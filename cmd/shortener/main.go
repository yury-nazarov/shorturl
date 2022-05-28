package main

import (
	"flag"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/yury-nazarov/shorturl/internal/app/handler"
	appMiddleware "github.com/yury-nazarov/shorturl/internal/app/middleware"
	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/app/storage"
)


func main() {
	run()
}

func run() {
	// Парсим аргументы командной строки
	serverAddressFlag 	:= flag.String("a", "", "set server address, by example: 127.0.0.1:8080")
	baseURLFlag 		:= flag.String("b", "", "set base URL, by example: http://127.0.0.1:8080")
	fileStoragePathFlag := flag.String("f", "", "set file path for storage, by example: db.txt")
	dataBaseStringFlag  := flag.String("d", "", "set database string for Postgres, by example: 'host=localhost port=5432 user=example password=123 dbname=example sslmode=disable connect_timeout=5'")
	flag.Parse()

	// Получаем переменные окружения
	serverAddressEnv 	:= os.Getenv("SERVER_ADDRESS")
	baseURLEnv 			:= os.Getenv("BASE_URL")
	fileStoragePathEnv 	:= os.Getenv("FILE_STORAGE_PATH")
	dataBaseStringEnv 	:= os.Getenv("DATABASE_DSN")

	// Устанавливаем конфигурационные параметры по приоритету:
	// 		1. Флаги;
	// 		2. Переменные окружения;
	// 		3. Дефолтное значение.
	serverAddress 	:= serverConfigInit(*serverAddressFlag, serverAddressEnv, "127.0.0.1:8080")
	baseURL 		:= serverConfigInit(*baseURLFlag, baseURLEnv, "http://127.0.0.1:8080")
	dbFileName 		:= serverConfigInit(*fileStoragePathFlag, fileStoragePathEnv, "")
	PGConnStr		:= serverConfigInit(*dataBaseStringFlag, dataBaseStringEnv, "")

	// Инициируем БД
	db := storage.New(storage.DBConfig{FileName: dbFileName, PGConnStr: PGConnStr})

	// Инициируем Router
	r := chi.NewRouter()


	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	//Собственные middleware
	r.Use(appMiddleware.HTTPResponseCompressor)
	r.Use(appMiddleware.HTTPRequestDecompressor)
	// Передае в middleware соеденение с БД
	r.Use(appMiddleware.HTTPCookieAuth(db))

	// Создаем объект для доступа к методам компрессии URL
	lc := service.NewLinkCompressor(5, baseURL)
	// Инициируем объект для доступа к хендлерам
	c := handler.NewController(db, lc)


	// API endpoints
	r.HandleFunc("/", c.DefaultHandler)
	r.Post("/", c.AddURLHandler)
	r.Get("/{urlID}", c.GetURLHandler)
	r.Route("/api", func(r chi.Router) {
		//r.Route("/user/urls", func(r chi.Router) {
		//	r.Delete("/", c.DeleteURLs)
		//	r.Get("/", c.GetUserURLs)
		//})
		r.Delete("/user/urls", c.DeleteURLs)
		r.Get("/user/urls", c.GetUserURLs)
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", c.AddJSONURLHandler)
			r.Post("/batch", c.AddJSONURLBatchHandler)
		})
	})
	r.HandleFunc("/ping", c.PingDB)


	// Запускаем сервер
	log.Fatal(http.ListenAndServe(serverAddress, r))
}

// serverConfigInit - возвращает приоритетное значение из переданых аргументов
func serverConfigInit(flag string, env string, defaultValue string) string {
	if len(flag) != 0 {
		return flag
	}
	if len(env) != 0 {
		return env
	}
	return defaultValue

}
