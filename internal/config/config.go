package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"
)

//  Получаем конфигурацию из переменных или флагов

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	URLLength       int    `env:"URLLength" envDefault:"5"`
}

func NewConfig(logger *logrus.Logger) (Config, error) {
	cfg := Config{}
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "set server address, by example: 127.0.0.1:8080")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "set base URL, by example: http://127.0.0.1:8080")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "set file path for storage, by example: db.txt")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "set database string for Postgres, by example: 'host=localhost port=5432 user=example password=123 dbname=example sslmode=disable connect_timeout=5'")

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}
	flag.Parse()
	logger.Info("the config success init")
	return cfg, nil
}
