package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
)


func (u *URLService) URLHandler(w http.ResponseWriter, r * http.Request) {
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
		originURL := string(bodyData)
		// TODO: поправить нейминг переменных/функции
		shortPath := ShortLink(originURL, u.URLLength)
		u.DB.Add(shortPath, originURL)


		// HTTP Response
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		// Подготавливаем сокращенный URL с адресом нашего сервиса, на пример: http://127.0.0.1:8080/qweEER
		shortURL := fmt.Sprintf("%s://%s:%d/%s", "http", u.ListenAddress, u.Port, shortPath)
		// Отправляем и обрабатываем HTTP Response
		_, err = w.Write([]byte(shortURL))
		if err != nil {
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return

	case "GET":
		// Получаем path/urlID из URL для дальнейшего поиска по нему в БД
		urlID := r.URL.Path[1:]
		//urlID := chi.URLParam("urlID")
		// Получаем оригинальный URL
		originURL, err := u.DB.Get(urlID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
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
