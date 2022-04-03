package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/yury-nazarov/shorturl/internal/app/handler"
	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/app/storage"
)

func main() {

	router := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	db := storage.NewInMemoryDB()
	lc := service.NewLinkCompressor(5,  "http://127.0.0.1:8080")
	c := handler.NewController(db, lc)

	router.HandleFunc("/", c.DefaultHandler)
	router.Get("/{urlID}", c.GetURLHandler)
	router.Post("/", c.AddURLHandler)

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", router))
}

