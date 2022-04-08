package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/app/storage"
)

type Controller struct {
	db storage.Repository
	lc service.LinkCompressor
}

type URL struct {
	Request 	string `json:"url,omitempty"` 	 	// Не учитываем поле при Marshal
	Response  	string `json:"result,omitempty"`	// Не учитываем поле при Unmarshal
}

// NewController - вернет объект для доступа к эндпоинтам
func NewController(db storage.Repository,  lc service.LinkCompressor) *Controller {
	c := &Controller{
		db: db,
		lc: lc,
	}
	return c
}

func (c *Controller) AddJSONURLHandler(w http.ResponseWriter, r *http.Request) {
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

	// Unmarshal JSON
	var url URL
	if err = json.Unmarshal(bodyData, &url); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// Сокращаем url и добавляем в БД
	shortURL := c.lc.SortURL(url.Request)
	if err = c.db.Add(shortURL, url.Request); err != nil {
		log.Print(err)
	}

	// Сериализуем контент
	jsonShortURL, err := json.Marshal(URL{Response: shortURL})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Указываем заголовки в зависмости от типа контента
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// HTTP Response
	_, err = w.Write(jsonShortURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
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
	if err = c.db.Add(shortURL, originURL); err != nil {
		log.Print(err)
	}

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
