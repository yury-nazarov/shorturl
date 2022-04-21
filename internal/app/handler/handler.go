package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository"
)

type Controller struct {
	db repository.Repository
	lc service.LinkCompressor
}

type URL struct {
	Request  string `json:"url,omitempty"`    // Не учитываем поле при Marshal
	Response string `json:"result,omitempty"` // Не учитываем поле при Unmarshal
}

// NewController - вернет объект для доступа к хендлерам
func NewController(db repository.Repository, lc service.LinkCompressor) *Controller {
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
	token, err := r.Cookie("session_token")
	if err != nil {
		log.Print("AddURLHandler: err:",err)
	}
	if err = c.db.Add(shortURL, url.Request, token.Value); err != nil {
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

func (c *Controller) AddURLHandler(w http.ResponseWriter, r *http.Request) {
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

	// Сокращаем url и добавляем в БД: сокращенный url, оригинальный url, token идентификатор пользователя
	originURL := string(bodyData)
	shortURL := c.lc.SortURL(originURL)
	token, err := r.Cookie("session_token")
	if err != nil {
		log.Print("AddURLHandler: err:",err)
	}
	if err = c.db.Add(shortURL, originURL, token.Value); err != nil {
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

func (c *Controller) GetURLHandler(w http.ResponseWriter, r *http.Request) {

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


func (c *Controller) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	// Получаем токен из кук
	token, err := r.Cookie("session_token")
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Достаем из БД все записи по токену
	userURL, err := c.db.GetUserURL(token.Value)
	if err != nil {
		log.Print(err)
	}

	answer, err := json.Marshal(userURL)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	if len(userURL) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Указываем заголовки в зависмости от типа отдаваемого контента
	w.Header().Add("Content-Type", "application/json")
	// HTTP Response
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(answer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (c *Controller) DefaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
