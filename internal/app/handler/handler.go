package handler

import (
	"encoding/json"
	"fmt"
	"github.com/yury-nazarov/shorturl/internal/app/repository/db"
	"io"
	"net/http"
	"sync"

	"github.com/yury-nazarov/shorturl/internal/app/repository/models"
	"github.com/yury-nazarov/shorturl/internal/app/service"

	"github.com/sirupsen/logrus"
)

// Controller структура для создания контроллера
type Controller struct {
	db db.Repository
	lc service.LinkCompressor
	logger 	*logrus.Logger
}

// NewController - вернет объект для доступа к хендлерам
func NewController(db db.Repository, lc service.LinkCompressor, logger *logrus.Logger) *Controller {
	c := &Controller{
		db: db,
		lc: lc,
		logger: logger,
	}
	logger.Info("the controller success init")
	return c
}

// AddJSONURLHandler - принимает URL в формате JSON
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
	var url models.URL
	if err = json.Unmarshal(bodyData, &url); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	// Проверяем если в БД уже есть оригинальный URL, нуже для верной установки заголовков ответа
	originURLExists, err := c.db.OriginURLExists(r.Context(), url.Request)
	if err != nil {
		c.logger.Print("OriginURLExists: ", err)
	}

	// Сокращаем url и добавляем в БД
	shortURL := c.lc.SortURL(url.Request)
	token, err := r.Cookie("session_token")
	if err != nil {
		c.logger.Print("AddURLHandler: err:",err)
	}
	if err = c.db.Add(r.Context(), shortURL, url.Request, token.Value); err != nil {
		c.logger.Print(err)
	}

	// Сериализуем контент
	jsonShortURL, err := json.Marshal(models.URL{Response: shortURL})
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

// AddURLHandler - принимает URL в текстовом формате
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

	// Проверяем если в БД уже есть оригинальный URL, ниже для верной установки заголовков ответа
	originURL := string(bodyData)
	originURLExists, err := c.db.OriginURLExists(r.Context(), originURL)
	if err != nil {
		c.logger.Print("OriginURLExists: ", err)
	}
	// Сокращаем url и добавляем в БД: сокращенный url, оригинальный url, token идентификатор пользователя
	shortURL := c.lc.SortURL(originURL)

	// Добавляем в БД только если URL нет в БД
	if !originURLExists {
		token, err := r.Cookie("session_token")
		if err != nil {
			c.logger.Print("AddURLHandler: err:",err)
		}
		if err = c.db.Add(r.Context(), shortURL, originURL, token.Value); err != nil {
			c.logger.Print(err)
		}
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
	originURL, err := c.db.Get(r.Context(), shortURL, userToken)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
	c.logger.Printf("DEBUG: User: %s get URL: %s -> %s\n", token,  shortURL, originURL)

	// HTTP 410 если url помечен как удаленный
	if len(originURL) == 0 {
		w.WriteHeader(http.StatusGone)
		return
	}
	// HTTP 307 Если url есть в БД
	w.Header().Set("Location", originURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// GetUserURLs - вернет список всех пользовательских URL
func (c *Controller) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	// Получаем токен из кук
	token, err := r.Cookie("session_token")
	if err != nil {
		c.logger.Println("DEBUG 1")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Достаем из БД все записи по токену
	userURL, err := c.db.GetUserURL(r.Context(), token.Value)
	if err != nil {
		c.logger.Print(err)
	}

	answer, err := json.Marshal(userURL)
	if err != nil {
		c.logger.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	if len(userURL) == 0 {
		c.logger.Println("DEBUG 2")
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

// DeleteURLs помечает удаленными URL по идентификатору (сокращенная часть url)
//			  202 Accepted - успешное выполнение запроса пользователем его создавшем
func (c *Controller) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	// Читаем из body [ "a", "b", "c", "d", ...] сериализовать в JSON
	bodyData, err := io.ReadAll(r.Body)
	c.logger.Println("bodyData:", string(bodyData))
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

	// Получаем токен пользователя пользователя, если токена нет - удалять нечего
	token, err := r.Cookie("session_token")
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Получаем id записей которые нужно пометить удаленными
	urlsID := make(chan int, len(urlIdentityList))


	var wg sync.WaitGroup
	for _, identity := range urlIdentityList {
		wg.Add(1)
		go func(identity string) {
			id := c.db.GetShortURLByIdentityPath(r.Context(), identity, token.Value)
			fmt.Printf("DEBUG: Prepare mark deleted URL identity: %s with ID: %d\n", identity, id)
			urlsID <- id
			wg.Done()
		}(identity)
	}
	// Закрываем канал когда он заполнился
	wg.Wait()
	close(urlsID)

	// Помечаем удаленными пачку записей
	if err = c.db.URLBulkDelete(r.Context(), urlsID); err != nil {
		c.logger.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	c.logger.Println("DEBUG: Stop URLBulkDelete:")
	w.WriteHeader(http.StatusAccepted)
}


// DefaultHandler - TODO
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

// AddJSONURLBatchHandler - добавляет пачку URL пришедших в формате JSON
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
	var urls []models.URLBatch
	if err = json.Unmarshal(bodyData, &urls); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// Сокращаем url и добавляем в БД, подготавливаем ответ
	var response []models.URLBatch
	for _, item := range urls {
		shortURL := c.lc.SortURL(item.OriginalURL)
		token, err := r.Cookie("session_token")
		if err != nil {
			c.logger.Print("AddURLHandler: err:",err)
		}
		if err = c.db.Add(r.Context(), shortURL, item.OriginalURL, token.Value); err != nil {
			c.logger.Print(err)
		}

		// Сразу подготавливаем слайс для ответа пользователю
		response = append(response, models.URLBatch{
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