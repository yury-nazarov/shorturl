package main

import (
	"fmt"
	"github.com/yury-nazarov/shorturl/internal/app/url"
	"io"
	"log"
	"net/http"
)

var (
	db = &url.URLDB{
		DB: map[string]string{},
		URLLength: 5,
	}
	httpProtocol = "http"
	listenAddress = "127.0.0.1"
	port = 8080
	serviceName = fmt.Sprintf("%s://%s:%d", httpProtocol, listenAddress, port)
)


func main() {
	http.HandleFunc("/", urlHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", listenAddress, port), nil))
}

func urlHandler(w http.ResponseWriter, r * http.Request) {
	switch r.Method {
	case "POST":
		// Читаем присланые данные
		bodyData, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("ReadBody:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Проверяем пустой Body
		if len(bodyData) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Сокращаем url и добавляем в БД
		shortPath := db.Add(string(bodyData))

		// Отправляем ответ
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(fmt.Sprintf("%s/%s", serviceName, shortPath)))
		if err != nil {
			log.Println("writeResponse:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return

	case "GET":
		// Получаем path/urlID из URL для дальнейшего поиска по нему в БД
		urlID := r.URL.Path[1:]
		// Получаем оригинальный URL
		originURL, err := db.Get(urlID)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Отправляем ответ
		w.Header().Set("Location", originURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
