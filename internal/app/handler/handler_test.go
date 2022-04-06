package handler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/app/storage"
)


// NewTestServer - конфигурируем тестовый сервер,
func NewTestServer() *httptest.Server{
	ServiceAddress := "127.0.0.1:8080"

	r := chi.NewRouter()
	db := storage.NewInMemoryDB()
	//db := storage.NewFileDB("test_db.txt") // TODO: Подумать как лучше сделать оба теста
											 // TODO: Убирать за собой тестовые данные|файлы
	lc := service.NewLinkCompressor(5, fmt.Sprintf("http://%s", ServiceAddress))
	c := NewController(db, lc)

	// Handler routing
	r.HandleFunc("/", c.DefaultHandler)
	r.Post("/api/shorten", c.AddJSONURLHandler)
	r.Get("/{urlID}", c.GetURLHandler)
	r.Post("/", c.AddURLHandler)

	// Настраиваем адрес/порт который будут слушать тестовый сервер
	listener, err := net.Listen("tcp", ServiceAddress)
	if err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(r)
	// Закрываем созданый httptest.NewUnstartedServer Listener и назначаем подготовленный нами ранее
	// В тесткейсе нужно будет запустить и остановить сервер: ts.Start(), ts.Close()
	ts.Listener.Close()
	ts.Listener = listener
	return ts
}

// Функция HTTP клиент для тестовых запросов
func testRequest(t *testing.T,  method, path string, body string) (*http.Response, string){

	req, err := http.NewRequest(method, path, strings.NewReader(body))
	require.NoError(t, err)

	// Убираем редирект в HTTP клиенте, для коректного тестирования HTTP хендлеров c Header Location
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestController_AddJSONURLHandler(t *testing.T) {
	// Вспомогательная структура, описывает HTTP Headers для структур: request и/или want
	type header struct {
		contentType string
		locations string
	}
	// Параметры для настройки тестового HTTP Request
	type request struct {
		httpMethod string
		url string
		header header
		body string
	}
	// Ожидаемый ответ сервера
	type want struct {
		statusCode int
		header header
		body string
	}
	// Список тесткейсов
	tests := []struct{
		name string
		request request
		want want
	}{
		{
			name: "test_1: POST: Success request",
			request: request{
				httpMethod: http.MethodPost,
				url:        "http://127.0.0.1:8080/api/shorten",
				body:       `{"url":"https://www.youtube.com/watch?v=09nmlZjxRFs"}`,
				header: header{
					contentType: "application/json",
				},
			},
			want: want{
				statusCode: 201,
				body:       `{"result":"http://127.0.0.1:8080/KJYUS"}`,
				header: header{
					contentType: "application/json",
				},
			},
		},
		{
			name: "test_1: POST: Empty request",
			request: request{
				httpMethod: http.MethodPost,
				url:        "http://127.0.0.1:8080/api/shorten",
				body:       "",
				header: header{
					contentType: "application/json",
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	ts := NewTestServer()
	ts.Start()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, tt.request.httpMethod, tt.request.url, tt.request.body)
			defer resp.Body.Close() // go vet test

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.header.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.body, body)
		})
	}
	ts.Close()
}

func TestController_AddUrlHandler(t *testing.T) {
	// Вспомогательная структура, описывает HTTP Headers для структур: request и/или want
	type header struct {
		contentType string
		locations string
	}
	// Параметры для настройки тестового HTTP Request
	type request struct {
		httpMethod string
		url string
		body string
	}
	// Ожидаемый ответ сервера
	type want struct {
		statusCode int
		header header
		body string
	}
	// Список тесткейсов
	tests := []struct{
		name string
		request request
		want want
	}{
		{
			name: "test_1: POST: Success request",
			request: request{
				httpMethod: http.MethodPost,
				url:        "http://127.0.0.1:8080",
				body:       "https://www.youtube.com/watch?v=09nmlZjxRFs",
			},
			want: want{
				statusCode: http.StatusCreated,
				body: "http://127.0.0.1:8080/KJYUS",
				header: header{contentType: "text/plain"},

			},
		},
		{
			name: "test_2: POST: Empty body",
			request: request{
				httpMethod: http.MethodPost,
				url: "http://127.0.0.1:8080",
				body: "",
			},
			want: want {
				statusCode: http.StatusBadRequest,
			},
		},
	}

	ts := NewTestServer()
	ts.Start()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, tt.request.httpMethod, tt.request.url, tt.request.body)
			defer resp.Body.Close() // go vet test

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.header.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.body,  body)
		})
	}
	ts.Close()
}

func TestController_GetUrlHandler(t *testing.T) {
	// Вспомогательная структура, описывает HTTP Headers для структур: request и/или want
	type header struct {
		contentType string
		locations string
	}
	// Параметры для настройки тестового HTTP Request
	type request struct {
		httpMethod string
		url string
		body string
	}
	// Ожидаемый ответ сервера
	type want struct {
		header header
		statusCode int
		body string
	}
	// Список тесткейсов
	tests := []struct{
		name string
		request request
		want want
	}{
		{
			name: "Prepare: Add test url into DB",
			request: request{
				httpMethod: http.MethodPost,
				url:        "http://127.0.0.1:8080",
				body:       "https://www.youtube.com/watch?v=09nmlZjxRFs",
			},
			want: want{
				header:     header{contentType: "text/plain"},
				statusCode: 201,
				body:       "http://127.0.0.1:8080/KJYUS",
			},
		},
		{
			name: "test_1: GET: Success short url from test #1",
			request: request{
				httpMethod: http.MethodGet,
				url: "http://127.0.0.1:8080/KJYUS",
			},
			want: want {
				header: header{locations: "https://www.youtube.com/watch?v=09nmlZjxRFs"},
				statusCode: http.StatusTemporaryRedirect,
			},
		},
		{
			name: "test_2: GET: Short url not found",
			request: request{
				httpMethod: http.MethodGet,
				url: "http://127.0.0.1:8080/qqWW",
			},
			want: want {
				statusCode: http.StatusNotFound,
			},
		},
	}

	ts := NewTestServer()
	ts.Start()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := testRequest(t, tt.request.httpMethod, tt.request.url, tt.request.body)
			defer resp.Body.Close() // go vet test from github

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.header.locations, resp.Header.Get("Location"))
		})
	}
	ts.Close()
}


func TestController_DefaultHandler(t *testing.T) {
	// Вспомогательная структура, описывает HTTP Headers для структур: request и/или want
	type header struct {
		contentType string
		locations string
	}
	// Параметры для настройки тестового HTTP Request
	type request struct {
		httpMethod string
		url string
		body string
	}
	// Ожидаемый ответ сервера
	type want struct {
		header header
		statusCode int
		body string
	}
	// Список тесткейсов
	tests := []struct{
		name string
		request request
		want want
	}{
		{
			name: "test_1: Default: HTTP PUT request",
			request: request{
				httpMethod: http.MethodPut,
				url: "http://127.0.0.1:8080",
			},
			want: want {
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "test_1: Default: HTTP Delete request",
			request: request{
				httpMethod: http.MethodDelete,
				url: "http://127.0.0.1:8080",
			},
			want: want {
				statusCode: http.StatusBadRequest,
			},
		},
	}

	ts := NewTestServer()
	ts.Start()
	for _, tt := range tests{
		t.Run(tt.name, func(t *testing.T){
			resp, body := testRequest(t, tt.request.httpMethod, tt.request.url, tt.request.body)
			defer resp.Body.Close() // go vet test from github

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.body, body)
		})
	}
	ts.Close()
}

