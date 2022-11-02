package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/yury-nazarov/shorturl/internal/app/grpcserver/server"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/yury-nazarov/shorturl/internal/app/grpcserver/proto"

	"github.com/yury-nazarov/shorturl/internal/app/handler"
	"github.com/yury-nazarov/shorturl/internal/app/repository/db"
	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/config"
	"github.com/yury-nazarov/shorturl/internal/logger"

	"google.golang.org/grpc"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	// Инициируем логгер.
	logger := logger.New()

	// Инициируем конфиг: аргументы cli > env
	cfg, err := config.NewConfig(logger)
	if err != nil {
		logger.Fatal(err)
	}
	// Инициируем БД.
	db, err := db.New(cfg, logger)
	if err != nil {
		logger.Fatal(err)
	}
	// Создаем объект для доступа к методам компрессии URL.
	linkCompressor := service.NewLinkCompressor(cfg, logger)
	// Инициируем объект для доступа к хендлерам.
	controller := handler.NewController(db, cfg, linkCompressor, logger)
	// Инициируем роутер.
	r := handler.NewRouter(controller, db, logger)

	// Публикуем служебную информацию
	logger.Info("Build version: ", buildVersion)
	logger.Info("Build date: ", buildDate)
	logger.Info("Build commit: ", buildCommit)

	// Подготавливаем сервер к запуску и остановке
	srv := http.Server{Addr: cfg.ServerAddress, Handler: r}

	// Через этот канал, сообщим основному потоку выполнения программы, что сетевые соединения закрыты
	// и можно корректно завершить выполнение запросов в БД, закрыть открытые файлы и т.д.
	idleConnectionClose := make(chan struct{})

	// канал для перенаправления прерываний
	sigint := make(chan os.Signal, 1)

	// регистрируем перенаправление прерываний которые будем обрабатывать
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// запускаем параллельно горутину для обработки пойманных прерываний
	go func() {
		// Закрываемсетевые соединения корректно завершая обработку HTTP запроов клиентов
		<-sigint
		if err = srv.Shutdown(context.Background()); err != nil {
			logger.Infof("HTTP Server Shutdown: %v", err)
		}
		// Закрываем канал, сообщая что все сетевые соединения завершены
		close(idleConnectionClose)
	}()

	///////// gRPC Server /////////
	go gRPCServer(db, linkCompressor, logger, cfg)

	///////// HTTP/HTTPS Server /////////
	if cfg.TLS {
		certFile := "internal/tls/cert.crt"
		keyFile := "internal/tls/private.key"
		logger.Info("the HTTPS server run on ", cfg.ServerAddress)
		if err = srv.ListenAndServeTLS(certFile, keyFile); err != http.ErrServerClosed {
			logger.Fatalf("HTTP Server ListenAndServer: %v", err)
		}
	} else {
		logger.Info("the HTTP server run on ", cfg.ServerAddress)
		if err = srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatalf("HTTP Server ListenAndServer: %v", err)
		}
	}

	/////// Graceful shutdown /////////
	<-idleConnectionClose
	// Завершаем работу с БД
	err = db.Close()
	if err != nil {
		logger.Infof("Close DB connection: %v", err)
	}
	logger.Infof("Server shutdown graseful")
}

// Установка необходимых зависимостей для работы с gRPC
// 		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
// 		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
// Генетит код из прото файла
/*
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
 	internal/app/grpcserver/proto/shorturl.proto
 */
func gRPCServer(db db.Repository, lc  service.LinkCompressor, logger *logrus.Logger, cfg config.Config) {
	// Определяем порт для сервера
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		logger.Fatalf("can't start GRPC server: %s", err)
	}
	// Создаем gRPC-сервер без зарегестрированной службы
	s := grpc.NewServer()
	// Регистрируем сервис ShorURLService на сервере gRPC
	pb.RegisterShortURLServer(s, &server.ShorURLService{
		DB: db,
		LC: lc,
		CFG: cfg,
	})

	logger.Info("The gRPC server has been started")

	// Получаем запрос gRPC
	if err := s.Serve(listen); err != nil {
		logger.Fatalf("stop GRPC server. Reason: %s", err)
	}
	// Display all ports on Mac
	// lsof -nP -iTCP -sTCP:LISTEN

}
