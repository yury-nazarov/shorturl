package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/yury-nazarov/shorturl/internal/logger"
	"os"
	"testing"
)

/*
Инициировать разные экземпляры конфига
1. По отдельности
	- Конфиг из флагов
    - Конфиг из переменных окружения
    - конфиг из файла - путь до файла передан из флага
    - конфиг из файла - путь до файла передан из переменной окружения

2. Каскадный тест
	-

3. Комплексный тест
*/

// Test_parseConfigFile тестируем чтение конфига из файла
func Test_parseConfigFile(t *testing.T) {
	// Ожидаемый результат
	wantCfg := Config{
		ServerAddress: "localhost:8181",
		BaseURL: "http://localhost",
		FileStoragePath: "/path/to/file.db",
		DatabaseDSN: "",
	}

	cfg, err := parseConfigFile("config.json")
	if err != nil {
		assert.Errorf(t, err, "parseConfigFile err")
	}
	assert.Equal(t, wantCfg, cfg)
}

// Test_parseConfigFile тест если файл не существует, должна вернтутся ошибка
func Test_parseConfigFile_ConfigFileNotExist(t *testing.T) {
	_, err := parseConfigFile("config_file_not_exist.json")
	assert.Errorf(t, err, "Configuration file must not exist!")
}

// Test_parseEnv тест чтения переменных окружения для инита конфига
func Test_parseEnv(t *testing.T) {
	// Ожидаемый результат
	wantCfg := Config{
		ServerAddress: "localhost-env:8181",
		BaseURL: "http://localhost-env",
		FileStoragePath: "/path/to/file.db",
		DatabaseDSN: "",
		URLLength: 5,
	}
	// Устанавливаем переменные окружения
	os.Setenv("SERVER_ADDRESS", "localhost-env:8181")
	os.Setenv("BASE_URL", "http://localhost-env")
	os.Setenv("FILE_STORAGE_PATH", "/path/to/file.db")

	cfg, err := parseEnv()
	if err != nil {
		assert.Errorf(t, err, "err")
	}
	assert.Equal(t, wantCfg, cfg)
}

// TODO: Пока не понятно как тестить запуск приложения с флагами
//// Test_parseFlag тест инит конфиг из флагов
//// Запустить нужно так же с флагами:
//func Test_parseFlag(t *testing.T) {
//	// Ожидаемый результат
//	wantCfg := Config{
//		ServerAddress: "localhost-flag:8182",
//		//BaseURL: "http://localhost-flag",
//		//FileStoragePath: "/path/to/file.db",
//		//DatabaseDSN: "",
//		//URLLength: 5,
//	}
//
//	cfg := parseFlag()
//	assert.Equal(t, wantCfg, cfg)
//}

// TestNewConfig Комплексный тест
//
func TestNewConfig(t *testing.T) {
	logger := logger.New()

	os.Setenv("SERVER_ADDRESS", "localhost-all:8080")
	// Значения из конфигурационного файла НЕ должны перезаписать SERVER_ADDRESS
	os.Setenv("CONFIG", "config.json")

	// Ожидаемый результат
	wantCfg := Config{
		// Значение из переменной окружения как более приоритетной
		ServerAddress: "localhost-all:8080",
		// Значения из файла конфигурации как из менее приритетного
		BaseURL: "http://localhost",
		FileStoragePath: "/path/to/file.db",
		DatabaseDSN: "",
		TLS: false,
	}


	cfg, err := NewConfig(logger)
	if err != nil {
		assert.Error(t, err)
	}
	assert.Equal(t, wantCfg, cfg)
}