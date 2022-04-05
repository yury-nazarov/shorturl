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

	r := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	db := storage.NewInMemoryDB()
	lc := service.NewLinkCompressor(5,  "http://127.0.0.1:8080")
	c := handler.NewController(db, lc)


	r.HandleFunc("/", c.DefaultHandler)
	r.Post("/api/shorten", c.AddJSONURLHandler)
	r.Get("/{urlID}", c.GetURLHandler)
	r.Post("/", c.AddURLHandler)

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", r))
}

