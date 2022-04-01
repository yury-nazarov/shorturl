package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/yury-nazarov/shorturl/internal/app/handler"
	"github.com/yury-nazarov/shorturl/internal/app/storage"
)

func main() {
	router := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	db := storage.New()
	c := handler.NewController(db, 5)

	router.Get("/{urlID}", c.GetUrlHandler)
	router.Post("/", c.AddUrlHandler)
	router.HandleFunc("/", c.AddUrlHandler)

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", router))
}

