package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strconv"
	"strings"
)

//  Получаем конфигурацию. Приоритет: флаги > переменные окружения > файл

// Config - конфиг сервиса.
// 			Общая структура которую используем для
//				- чтения переменных с помощю либы github.com/caarlos0/env
//				- парсинга флагов
//				- анмаршала JSON для чтения из файла конфига
type Config struct {
	// Адрес и порт на котором будет запущен сервис: 127.0.0.1:8080
	ServerAddress   string 	`env:"SERVER_ADDRESS"    json:"server_address"`
	// Основной FQDN/Адрес для сервиса сокращения URL: http://127.0.0.1/
	BaseURL         string 	`env:"BASE_URL"          json:"base_url"`
	// Путь до файла БД
	FileStoragePath string 	`env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	// DNS для подключения к Postgres
	DatabaseDSN     string 	`env:"DATABASE_DSN"      json:"database_dsn"`
	// Если True то будет запущен HTTPS. Допустимые значения: true, false
	TLS 			bool 	`env:"ENABLE_HTTPS"      json:"enable_https"`
	// Путь до файла конфигурации. Файл конфигурации обладает наименьшим приоритетом
	fileConfig      string 	`env:"CONFIG"`
	// Длина сокращенного URL
	URLLength       int    	`env:"URLLength"         json:"url_length"        envDefault:"5"`
}

// NewConfig создает объект для доступа к конфигу.
func NewConfig(logger *logrus.Logger) (Config, error) {
	var cfgs []Config

	// парсим флаги
	cfgFlag := parseFlag()
	logger.Infof("read config from Flags (cfgFlag): %+v", cfgFlag)
	cfgs = append(cfgs, cfgFlag)

	// парсим переменные среды
	cfgEnv, err := parseEnv()
	if err != nil {
		return cfgEnv, err
	}
	logger.Infof("read config from EnvVar (cfgEnv): %+v", cfgEnv)
	cfgs = append(cfgs, cfgEnv)

	// Парсим файл, если путь до него указан в переданом флаге или переменной среде
	if len(cfgFlag.fileConfig) > 0 {
		logger.Infof("find path for config file: cfgFlag.fileConfig=%s", cfgFlag.fileConfig)
		cfgFile, err := parseConfigFile(cfgFlag.fileConfig)
		if err != nil {
			return cfgFile, err
		}
		cfgs = append(cfgs, cfgFile)

	} else if len(cfgEnv.fileConfig) > 0 {
		logger.Infof("find path for config file: cfgEnv.fileConfig=%s", cfgEnv.fileConfig)
		cfgFile, err := parseConfigFile(cfgEnv.fileConfig)
		if err != nil {
			return cfgFile, err
		}
		cfgs = append(cfgs, cfgFile)
	}

	// Сравниваем переменные по приоритету cfgFlag > cfgEnv > cfgFile
	cfg := selectConfig(cfgs)

	// Если конфиг пустой, останавливаем выполнение программы
	msg := isEmptyConfig(cfg)
	if len(msg) != 0 {
		return cfg, fmt.Errorf("please set follow config item for run service: %s", msg)
	}

	logger.Infof("%+v\n", cfg)
	//logger.Info("the config success init")
	return cfg, nil
}

// selectConfig - выбирает приоритетное значение конфига: Flag > Env > File
func selectConfig(cfgs []Config) Config {
	logrus.Infof("Compare configs: %+v", cfgs)
	// Если какое ни будь поле конфига пустое то подставляем менее приоритетное
	cfg := Config{}

	for _, v := range cfgs {
		if len(cfg.ServerAddress) == 0 {
			cfg.ServerAddress = v.ServerAddress
		}
		if len(cfg.BaseURL) == 0 {
			cfg.BaseURL = v.BaseURL
		}
		if len(cfg.FileStoragePath) == 0 {
			cfg.FileStoragePath = v.FileStoragePath
		}
		if len(cfg.DatabaseDSN) == 0 {
			cfg.DatabaseDSN = v.DatabaseDSN
		}
		if len(strconv.FormatBool(cfg.TLS)) == 0 {
			cfg.TLS = v.TLS
		}
	}
	logrus.Infof("Final configs: %+v", cfg)
	return cfg
}

// parseFlag парсит флаги
func parseFlag() Config {
	cfg := Config{}
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "set server address, by example: 127.0.0.1:8080")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "set base URL, by example: http://127.0.0.1:8080")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "set file path for storage, by example: db.txt")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "set database string for Postgres, by example: 'host=localhost port=5432 user=example password=123 dbname=example sslmode=disable connect_timeout=5'")
	flag.BoolVar(&cfg.TLS, "s", cfg.TLS, "user -s for run HTTPS")
	flag.StringVar(&cfg.fileConfig, "c", cfg.fileConfig, "user -c for use config.json")
	// Парсим флаги
	flag.Parse()

	return cfg
}

// parseEnv парсим переменные окружения
func parseEnv() (Config, error) {
	cfg := Config{}

	// Читаем из переменных окружения
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

// parseConfigFile читает конфигурационный файл в формате JSON, парсит, возвращает структуру Config
func parseConfigFile(filePath string) (Config, error) {

	var cfg Config
	f, err := ioutil.ReadFile(filePath)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(f, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

// isEmptyConfig Проверяем наличие обязательного конфига для запуска сервиса
func isEmptyConfig(cfg Config) string {
	var message []string
	if len(cfg.ServerAddress) == 0 {
		message = append(message, "ServerAddress")
	}
	if len(cfg.BaseURL) == 0 {
		message = append(message, "BaseURL")
	}
	return strings.Join(message, ", ")
}