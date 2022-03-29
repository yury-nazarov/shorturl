package main

import (
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/yury-nazarov/shorturl/internal/app/server"
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
	s := server.New("127.0.0.1", 8080, 5, db, router)

	router.HandleFunc("/{urlID}", s.URLHandler)
	router.HandleFunc("/", s.URLHandler)


	log.Fatal(s.Bind.ListenAndServe())
}

