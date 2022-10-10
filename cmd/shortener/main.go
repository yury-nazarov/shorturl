package main

import (
	"net/http"

	"github.com/yury-nazarov/shorturl/internal/app/handler"
	"github.com/yury-nazarov/shorturl/internal/app/repository/db"
	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/config"
	"github.com/yury-nazarov/shorturl/internal/logger"
)

var (
	buildVersion = "N/A"
	buildDate = "N/A"
	buildCommit = "N/A"
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
	controller := handler.NewController(db, linkCompressor, logger)
	// Инициируем роутер.
	r := handler.NewRouter(controller, db, logger)
	// Запускаем сервер.
	logger.Info("Build version: ", buildVersion)
	logger.Info("Build date: ", buildDate)
	logger.Info("Build commit: ", buildCommit)

	if cfg.TLS {
		certFile := "internal/tls/cert.crt"
		keyFile  := "internal/tls/private.key"
		logger.Info("the HTTPS server run on ", cfg.ServerAddress)
		logger.Fatal(http.ListenAndServeTLS(cfg.ServerAddress, certFile, keyFile, r))
	} else {
		logger.Info("the HTTP server run on ", cfg.ServerAddress)
		logger.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
	}
}
