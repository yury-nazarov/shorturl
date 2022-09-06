package main

import (
	"log"
	"net/http"

	"github.com/yury-nazarov/shorturl/internal/app/handler"
	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/app/storage"
	"github.com/yury-nazarov/shorturl/internal/config"
	"github.com/yury-nazarov/shorturl/internal/logger"
)

func main() {
	// ИНициируем логгер
	logger := logger.New()

	// Инициируем конфиг: аргументы cli > env
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal(err)
	}
	// Инициируем БД
	db := storage.New(storage.DBConfig{FileName: cfg.FileStoragePath, PGConnStr: cfg.DatabaseDSN})
	// Создаем объект для доступа к методам компрессии URL
	lc := service.NewLinkCompressor(5, cfg.BaseURL)
	// Инициируем объект для доступа к хендлерам
	c := handler.NewController(db, lc)
	// Инициируем роутер
	r := handler.NewRouter(c, db)
	// Запускаем сервер
	log.Println("run server on", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}

