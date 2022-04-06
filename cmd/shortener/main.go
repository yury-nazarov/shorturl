package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/yury-nazarov/shorturl/internal/app/handler"
	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/app/storage"
	"github.com/yury-nazarov/shorturl/internal/app/storage/filedb"
	"github.com/yury-nazarov/shorturl/internal/app/storage/inmemorydb"
)

func main() {

	// Получаем конфигурацию из переменных окружения
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if len(serverAddress) == 0 {
		serverAddress = "127.0.0.1:8080"
	}

	baseURL := os.Getenv("BASE_URL")
	if len(baseURL) == 0 {
		baseURL = "http://127.0.0.1:8080"
	}

	// Если указан файл для хранения данных - храним в нем,
	// иначе храним в RAM/inMemoryDB
	fileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	var db storage.Repository
	if len(fileStoragePath) == 0 {
		db = inmemorydb.NewInMemoryDB()
	} else {
		db = filedb.NewFileDB(fileStoragePath)
	}

	r := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	lc := service.NewLinkCompressor(5,  baseURL)
	c := handler.NewController(db, lc)


	r.HandleFunc("/", c.DefaultHandler)
	r.Post("/api/shorten", c.AddJSONURLHandler)
	r.Get("/{urlID}", c.GetURLHandler)
	r.Post("/", c.AddURLHandler)

	log.Fatal(http.ListenAndServe(serverAddress, r))
}

