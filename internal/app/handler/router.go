package handler

import (
	"net/http"

	appMiddleware "github.com/yury-nazarov/shorturl/internal/app/middleware"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(c *Controller, db repository.Repository) http.Handler {
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
	return r
}