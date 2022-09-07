package main

import (
	"net/http"

	"github.com/yury-nazarov/shorturl/internal/app/handler"
	"github.com/yury-nazarov/shorturl/internal/app/repository"
	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/config"
	"github.com/yury-nazarov/shorturl/internal/logger"
)

func main() {
	// Инициируем логгер
	logger := logger.New()

	// Инициируем конфиг: аргументы cli > env
	cfg, err := config.NewConfig(logger)
	if err != nil {
		logger.Fatal(err)
	}
	// Инициируем БД
	db := repository.New(cfg, logger)
	// Создаем объект для доступа к методам компрессии URL
	linkCompressor := service.NewLinkCompressor(cfg, logger)
	// Инициируем объект для доступа к хендлерам
	controller := handler.NewController(db, linkCompressor, logger)
	// Инициируем роутер
	r := handler.NewRouter(controller, db, logger)
	// Запускаем сервер
	logger.Info("the server run on ", cfg.ServerAddress)
	logger.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}

