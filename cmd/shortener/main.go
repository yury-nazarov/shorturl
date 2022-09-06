package main

import (
	"net/http"


	"github.com/yury-nazarov/shorturl/internal/app/handler"
	appMiddleware "github.com/yury-nazarov/shorturl/internal/app/middleware"
	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/app/storage"
	"github.com/yury-nazarov/shorturl/internal/config"
	"github.com/yury-nazarov/shorturl/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)


func main() {
	run()
}

func run() {

	log := logger.New()

	// Инициируем конфиг: аргументы cli > env
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Инициируем БД
	db := storage.New(storage.DBConfig{FileName: cfg.FileStoragePath, PGConnStr: cfg.DatabaseDSN})

	// Инициируем Router
	r := chi.NewRouter()


	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	//Собственные middleware
	r.Use(appMiddleware.HTTPResponseCompressor)
	r.Use(appMiddleware.HTTPRequestDecompressor)
	// Передае в middleware соеденение с БД
	r.Use(appMiddleware.HTTPCookieAuth(db))

	// Создаем объект для доступа к методам компрессии URL
	lc := service.NewLinkCompressor(5, cfg.BaseURL)
	// Инициируем объект для доступа к хендлерам
	c := handler.NewController(db, lc)


	// API endpoints
	r.HandleFunc("/", c.DefaultHandler)
	r.Post("/", c.AddURLHandler)
	r.Get("/{urlID}", c.GetURLHandler)
	r.Route("/api", func(r chi.Router) {
		r.Delete("/user/urls", c.DeleteURLs)
		r.Get("/user/urls", c.GetUserURLs)
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", c.AddJSONURLHandler)
			r.Post("/batch", c.AddJSONURLBatchHandler)
		})
	})
	r.HandleFunc("/ping", c.PingDB)


	// Запускаем сервер
	log.Println("run server on", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}

// serverConfigInit - возвращает приоритетное значение из переданых аргументов
func serverConfigInit(flag string, env string, defaultValue string) string {
	if len(flag) != 0 {
		return flag
	}
	if len(env) != 0 {
		return env
	}
	return defaultValue

}
