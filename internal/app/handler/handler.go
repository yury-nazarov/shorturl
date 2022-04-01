package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/yury-nazarov/shorturl/internal/app/storage"
)

type Controller struct {
	db *storage.URLDB 	// TODO: Заменить на интерфейс
	URLLength int
	ListenAddress string
	Port int
}

func NewController(db *storage.URLDB, urlLength int) *Controller {
	c := &Controller{
		db: db,
		URLLength: urlLength,
		ListenAddress: "127.0.0.1", // TODO адрес и порт брать из рантайма!
		Port: 8080,
	}
	return c
}


func (c *Controller) AddUrlHandler(w http.ResponseWriter, r * http.Request) {
	// Читаем присланые данные
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
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
	shortPath := ShortPath(originURL, c.URLLength)
	c.db.Add(shortPath, originURL)


	// HTTP Response
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	// Подготавливаем сокращенный URL с адресом нашего сервиса, на пример: http://127.0.0.1:8080/qweEER
	shortURL := fmt.Sprintf("%s://%s:%d/%s", "http", c.ListenAddress, c.Port, shortPath)
	// Отправляем и обрабатываем HTTP Response
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	return
}

func (c *Controller) GetUrlHandler(w http.ResponseWriter, r * http.Request) {

	// Получаем path/urlID из URL для дальнейшего поиска по нему в БД
	urlID := r.URL.Path[1:]
	//urlID := chi.URLParam("urlID")
	// Получаем оригинальный URL
	originURL, err := c.db.Get(urlID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Отправляем ответ
	w.Header().Set("Location", originURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	return
}

func (c *Controller) DefaultHandler(w http.ResponseWriter, r * http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	return
}
