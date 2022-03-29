package main

import (
	"log"
	"net/http"

	"github.com/yury-nazarov/shorturl/internal/app/server"
	"github.com/yury-nazarov/shorturl/internal/app/storage"
)

func main() {

	db := storage.New()
	s := server.New("127.0.0.1", 8080, 5, db)
	http.HandleFunc("/", s.UrlHandler)
	log.Fatal(s.Bind.ListenAndServe())
}

