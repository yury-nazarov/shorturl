package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

//  Получаем конфигурацию из переменных или флагов

type config struct {
	ServerAddress    string `env:"SERVER_ADDRESS"`
	BaseURL 		 string `env:"BASE_URL"`
	FileStoragePath  string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN      string `env:"DATABASE_DSN"`
}

func NewConfig() (config, error) {
	cfg := config{}
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "set server address, by example: 127.0.0.1:8080")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "set base URL, by example: http://127.0.0.1:8080")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "set file path for storage, by example: db.txt")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "set database string for Postgres, by example: 'host=localhost port=5432 user=example password=123 dbname=example sslmode=disable connect_timeout=5'")

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}
	flag.Parse()
	return cfg, nil
}
