package main

import (
	"fmt"
	"github.com/yury-nazarov/shorturl/internal/app/url"
	"io"
	"log"
	"net/http"
	"strings"
)

var (
	db = &url.URLDB{
		DB: map[string]string{},
		URLLength: 5,
	}
	fqdn = "http://127.0.0.1:8080/"
)


func main() {
	http.HandleFunc("/", urlHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
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
		url := db.Add(string(bodyData))

		// Отправляем ответ
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(fmt.Sprintf("%s%s", fqdn, url)))
		if err != nil {
			log.Println("writeResponse:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return

	case "GET":
		// Получаем id из URL для дальнейшего поиска в БД
		urlID := strings.Split(r.URL.String(), "/")[1]
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
