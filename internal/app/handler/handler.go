package handler

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"

	"github.com/yury-nazarov/shorturl/internal/app/storage"
)

type Controller struct {
	db *storage.InMemoryDB 	// TODO: Заменить на интерфейс
	URLLength int
	ListenAddress string
	Port int
}

func NewController(db *storage.InMemoryDB, urlLength int) *Controller {
	c := &Controller{
		db: db,
		URLLength: urlLength,
	}
	return c
}

func (c *Controller) AddURLHandler(w http.ResponseWriter, r * http.Request) {
	// Читаем присланые данные
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
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
	shortURL := fmt.Sprintf("%s://%s/%s", "http", r.Host, shortPath)
	// Отправляем и обрабатываем HTTP Response
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (c *Controller) GetURLHandler(w http.ResponseWriter, r * http.Request) {

	// Получаем urlID из URL path для дальнейшего поиска по нему в БД
	urlID := chi.URLParam(r, "urlID")

	// Получаем оригинальный URL из БД
	originURL, err := c.db.Get(urlID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
	// Отправляем ответ
	w.Header().Set("Location", originURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (c *Controller) DefaultHandler(w http.ResponseWriter, r * http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
