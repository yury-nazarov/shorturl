package handler

import (
	"context"
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
	ctx context.Context
}

type URL struct {
	Request  string `json:"url,omitempty"`    // Не учитываем поле при Marshal
	Response string `json:"result,omitempty"` // Не учитываем поле при Unmarshal
}


type URLBatch struct {
	CorrelationID 	string `json:"correlation_id"`
	OriginalURL 	string `json:"original_url,omitempty"`
	ShortURL 		string `json:"short_url,omitempty"`
}

// NewController - вернет объект для доступа к хендлерам
func NewController(ctx context.Context, db repository.Repository, lc service.LinkCompressor) *Controller {
	c := &Controller{
		db: db,
		lc: lc,
		ctx: ctx,
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
	// Проверяем если в БД уже есть оригинальный URL, нуже для верной установки заголовков ответа
	originURLExists, err := c.db.OriginURLExists(url.Request)
	if err != nil {
		log.Print("OriginURLExists: ", err)
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
	if originURLExists {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

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

	// Проверяем если в БД уже есть оригинальный URL, нуже для верной установки заголовков ответа
	originURL := string(bodyData)
	originURLExists, err := c.db.OriginURLExists(originURL)
	if err != nil {
		log.Print("OriginURLExists: ", err)
	}
	// Сокращаем url и добавляем в БД: сокращенный url, оригинальный url, token идентификатор пользователя
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
	if originURLExists {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// GetURLHandler по сокращенному  URL
//				вернет оригинальный URL
//				установит заголоко Location: originURL + HTTP 307
func (c *Controller) GetURLHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор пользователя
	// 		Пустая строка userToken нужна для обратной совместимости с inMemory и fileDB
	// 		т.к если куки не найдены, r.Cookie() не вернет объект токен со строковым свойством token.Value
	var userToken string
	token, err := r.Cookie("session_token")
	// Если токен существует нужно объявить это в userToken
	if err == nil {
		userToken = token.Value
	}

	// Получаем оригинальный URL из БД
	shortURL := fmt.Sprintf("%s%s", c.lc.ServiceName, r.URL.Path)
	originURL, err := c.db.Get(shortURL, userToken)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	// HTTP 410 если url помечен как удаленный
	if len(originURL) == 0 {
		w.WriteHeader(http.StatusGone)
		return
	}
	// HTTP 307 Если url есть в БД
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

// TODO Debug: iteration14_test.go:221: Не удалось дождаться удаления переданных URL в течении 20 секунд

// DeleteURLs помечает удаленными URL по идентификатору (сокращенная часть url)
//			  202 Accepted - успешное выполнение запроса пользователем его создавшем
func (c *Controller) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	// 1. Прочитать из body [ "a", "b", "c", "d", ...] сериализовать в JSON
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if len(bodyData) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Конвертируем в JSON данные из body
	var urlIdentityList []string
	if err = json.Unmarshal(bodyData, &urlIdentityList); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// 1.1. Получаем токен пользователя пользователя
	// 		если токена нет - удалять нечего
	token, err := r.Cookie("session_token")
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// fmt.Fprint(w, token.Value) // 3daa3dfdf17db865188026b7cc02e1b1f5c96bee2de9d247734f8b06c325a6be

	// 2. В цикле, пройтись по списку:
	//	  2.1:
	// 		TODO: Можно реализовать паттерн Fan-Out поместив в канал идентификаторы и выполняя SELECT в несколько потоков
	// 		 	  (но скорее всего бутылочное горлышко будет в тестах практикума)
	//
	var urlsID []int
	for _, identity := range urlIdentityList {
		id := c.db.GetShortURLByIdentityPath(identity, token.Value)
		urlsID = append(urlsID, id)
	}

	// Помечаем как удаленные
	for _, id := range urlsID {
		go c.db.URLMarkDeleted(id)
	}
	log.Println(urlsID)
	w.WriteHeader(http.StatusAccepted)

	//     2.2:
	// 		TODO: Релизовать паттерн Fan-In читая из канала и
	//		UPDATE shorten_url SET delete=true WHERE id=id
	//	   2.3
	//  	TODO: * формируем буфер batch update (pgx prepare statement)

}

func (c *Controller) DefaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

// PingDB - Проверка соединения с БД
func (c *Controller) PingDB(w http.ResponseWriter, r *http.Request) {
	if !c.db.Ping() {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) AddJSONURLBatchHandler(w http.ResponseWriter, r *http.Request) {
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
	var urls []URLBatch
	if err = json.Unmarshal(bodyData, &urls); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// Сокращаем url и добавляем в БД, подготавливаем ответ
	var response []URLBatch
	for _, item := range urls {
		shortURL := c.lc.SortURL(item.OriginalURL)
		token, err := r.Cookie("session_token")
		if err != nil {
			log.Print("AddURLHandler: err:",err)
		}
		if err = c.db.Add(shortURL, item.OriginalURL, token.Value); err != nil {
			log.Print(err)
		}

		// Сразу подготавливаем слайс для ответа пользователю
		response = append(response, URLBatch{
			CorrelationID: item.CorrelationID,
			ShortURL: shortURL,
		})
	}

	// Сериализуем ответ
	jsonShortURL, err := json.Marshal(response)
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