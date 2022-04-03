package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/app/storage"
)

type Controller struct {
	db *storage.InMemoryDB
	lc service.LinkCompressor
}

func NewController(db *storage.InMemoryDB,  lc service.LinkCompressor) *Controller {
	c := &Controller{
		db: db,
		lc: lc,
	}
	return c
}

func (c *Controller) AddURLHandler(w http.ResponseWriter, r * http.Request) {
	// Читаем присланые данные
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Проверяем пустой Body
	if len(bodyData) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Сокращаем url и добавляем в БД
	originURL := string(bodyData)
	shortURL := c.lc.SortURL(originURL)
	c.db.Add(shortURL, originURL)

	// HTTP Response
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (c *Controller) GetURLHandler(w http.ResponseWriter, r * http.Request) {

	// Получаем оригинальный URL из БД
	shortURL := fmt.Sprintf("%s%s", c.lc.ServiceName, r.URL.Path)
	originURL, err := c.db.Get(shortURL)
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
