package handler

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yury-nazarov/shorturl/internal/app/repository/db"
	"github.com/yury-nazarov/shorturl/internal/app/service"
	"github.com/yury-nazarov/shorturl/internal/config"
	"github.com/yury-nazarov/shorturl/internal/logger"
)

// NewTestServer - конфигурируем тестовый сервер,
func NewTestServer(dbName string, PGConnStr string) *httptest.Server {
	// Инициируем логгер
	logger := logger.New()

	// В дальнейшем на этот адрес/url будут завязаны тест кейсы
	cfg := config.Config{}
	cfg.ServerAddress = "127.0.0.1:8080"
	cfg.BaseURL = "http://127.0.0.1:8080"
	cfg.FileStoragePath = dbName
	cfg.DatabaseDSN = PGConnStr
	cfg.URLLength = 5

	linkCompressor := service.NewLinkCompressor(cfg, logger)

	// Инициируем БД
	db, err := db.New(cfg, logger)
	if err != nil {
		logger.Fatal(err)
	}
	controller := NewController(db, linkCompressor, logger)

	r := NewRouter(controller, db, logger)

	// Настраиваем адрес/порт который будут слушать тестовый сервер
	listener, err := net.Listen("tcp", cfg.ServerAddress)
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
func testRequest(t *testing.T, method, path string, body string, headers map[string]string) (*http.Response, string) {
	// Подготавливаем HTTP Request для тестового сервера
	req, err := http.NewRequest(method, path, strings.NewReader(body))
	require.NoError(t, err)

	// Устанавливаем нужные хедеры для HTTP Request
	for name, value := range headers {
		req.Header.Set(name, value)
	}

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

// gzipCompressor - вспомогательная функция,
//				    позволяет компресить в формате gzip
func gzipCompressor(payload string) *bytes.Buffer {
	b := []byte(payload)
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err := g.Write(b); err != nil {
		log.Fatal(err)
	}
	if err := g.Close(); err != nil {
		log.Print(err)
	}
	return &buf
}

// gzipDecompressor - вспомогательная функция,
//					  позволяет извлекать данные из формата gzip
func gzipDecompressor(body string) string {
	gz, err := gzip.NewReader(strings.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	defer gz.Close()

	result, err := io.ReadAll(gz)
	if err != nil {
		log.Fatal(err)
	}
	return string(result)
}

func TestController_AddJSONURLHandler(t *testing.T) {
	// Параметры для настройки тестового HTTP Request
	type request struct {
		httpMethod string
		url        string
		headers    map[string]string
		body       string
	}
	// Ожидаемый ответ сервера
	type want struct {
		statusCode int
		headers    map[string]string
		body       string
	}
	// Список тесткейсов
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "test_1: POST: Success JSON request",
			request: request{
				httpMethod: http.MethodPost,
				url:        "http://127.0.0.1:8080/api/shorten",
				body:       `{"url":"https://www.youtube.com/watch?v=09nmlZjxRFs"}`,
				headers:    map[string]string{"Content-Type": "application/json"},
			},
			want: want{
				statusCode: http.StatusCreated,
				body:       `{"result":"http://127.0.0.1:8080/KJYUS"}`,
				headers:    map[string]string{"Content-Type": "application/json"},
			},
		},
		{
			name: "test_2: POST: Empty request",
			request: request{
				httpMethod: http.MethodPost,
				url:        "http://127.0.0.1:8080/api/shorten",
				body:       "",
				headers:    map[string]string{"Content-Type": "application/json"},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "test_3: POST: Server incoming request in the gzip format, return the text format.",
			request: request{
				httpMethod: http.MethodPost,
				url:        "http://127.0.0.1:8080/api/shorten",
				body:       gzipCompressor(`{"url":"https://www.youtube.com/watch?v=09nmlZjxRFs"}`).String(),
				headers:    map[string]string{"Content-Encoding": "gzip"},
			},
			want: want{
				statusCode: http.StatusCreated,
				body:       `{"result":"http://127.0.0.1:8080/KJYUS"}`,
				headers:    map[string]string{"Content-Type": "application/json"},
			},
		},
		{
			name: "test_4: POST: Server incoming request in the text format with Header: 'Accept-Encoding: gzip', return the gzip format.",
			request: request{
				httpMethod: http.MethodPost,
				url:        "http://127.0.0.1:8080/api/shorten",
				body:       `{"url":"https://www.youtube.com/watch?v=09nmlZjxRFs"}`,
				headers:    map[string]string{"Accept-Encoding": "gzip"},
			},
			want: want{
				statusCode: http.StatusCreated,
				body:       `{"result":"http://127.0.0.1:8080/KJYUS"}`,
				// TODO: Не совсем понятно, нужно ли при этом еще ставить заголовки указывающие что внутри JSON
				headers: map[string]string{"Content-Encoding": "gzip"},
			},
		},
	}

	// Запускаем одинаковые тесты на разной конфигурации сервера: inMemoryDB, fileDB
	tsDBName := []string{"inMemoryDB", "fileDB"}
	for _, dbName := range tsDBName {
		ts := NewTestServer(dbName, "")
		ts.Start()
		for _, tt := range tests {
			testName := fmt.Sprintf("%s: DB: %s", tt.name, dbName)
			t.Run(testName, func(t *testing.T) {
				// Выполняем тестовый HTTP Request
				resp, body := testRequest(t, tt.request.httpMethod, tt.request.url, tt.request.body, tt.request.headers)
				defer resp.Body.Close() // go vet test

				// Проверяем все возможные хедеры для ожидаемого ответа, которые указали в тест кейсах выше.
				// Если хедер не указан, то будет возвращено пустое значение.
				// Возможно не самое элегантное решение, но пока я лучше не придумал :-/
				wantContentTypeHeader := tt.want.headers["Content-Type"]
				wantAcceptEncodingHeader := tt.want.headers["Accept-Encoding"]
				wantContentEncodingHeader := tt.want.headers["Content-Encoding"]

				// Если мы ожидаем сжатый в gzip ответ. Т.е. в HTTP Response пришел херед: "Content-Encoding: gzip"
				// В этом случае нужно распаковать из gzip body и сравнить результат с ожидаемым
				if resp.Header.Get("Content-Encoding") == "gzip" {
					assert.Equal(t, wantContentEncodingHeader, resp.Header.Get("Content-Encoding"))
					assert.Equal(t, tt.want.body, gzipDecompressor(body))
				} else {
					assert.Equal(t, wantContentTypeHeader, resp.Header.Get("Content-Type"))
					assert.Equal(t, wantAcceptEncodingHeader, resp.Header.Get("Accept-Encoding"))

					assert.Equal(t, tt.want.statusCode, resp.StatusCode)
					assert.Equal(t, tt.want.body, body)
				}
			})
		}
		ts.Close()
	}
	defer os.Remove("test_db.txt")
}

func TestController_AddUrlHandler(t *testing.T) {
	// Параметры для настройки тестового HTTP Request
	type request struct {
		httpMethod string
		url        string
		headers    map[string]string
		body       string
	}
	// Ожидаемый ответ сервера
	type want struct {
		statusCode int
		headers    map[string]string
		body       string
	}
	// Список тесткейсов
	tests := []struct {
		name    string
		request request
		want    want
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
				body:       "http://127.0.0.1:8080/KJYUS",
				headers:    map[string]string{"Content-Type": "text/plain"},
			},
		},
		{
			name: "test_2: POST: Empty body",
			request: request{
				httpMethod: http.MethodPost,
				url:        "http://127.0.0.1:8080",
				body:       "",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "test_3: POST: Server incoming request in the gzip format, return the text format.",
			request: request{
				httpMethod: http.MethodPost,
				url:        "http://127.0.0.1:8080",
				body:       gzipCompressor("https://www.youtube.com/watch?v=09nmlZjxRFs").String(),
				headers:    map[string]string{"Content-Encoding": "gzip"},
			},
			want: want{
				statusCode: http.StatusCreated,
				body:       "http://127.0.0.1:8080/KJYUS",
				headers:    map[string]string{"Content-Type": "text/plain"},
			},
		},
		{
			name: "test_4: POST: Server incoming request in the text format with Header: 'Accept-Encoding: gzip', return the gzip format.",
			request: request{
				httpMethod: http.MethodPost,
				url:        "http://127.0.0.1:8080",
				body:       "https://www.youtube.com/watch?v=09nmlZjxRFs",
				headers:    map[string]string{"Accept-Encoding": "gzip"},
			},
			want: want{
				statusCode: http.StatusCreated,
				body:       "http://127.0.0.1:8080/KJYUS",
				headers:    map[string]string{"Content-Encoding": "gzip"},
			},
		},
	}

	// Запускаем одинаковые тесты на разной конфигурации сервера: inMemoryDB, fileDB
	tsDBName := []string{"inMemoryDB", "fileDB"}
	for _, dbName := range tsDBName {
		ts := NewTestServer(dbName, "")
		ts.Start()
		for _, tt := range tests {
			testName := fmt.Sprintf("%s: DB: %s", tt.name, dbName)
			t.Run(testName, func(t *testing.T) {
				// Выполняем тестовый HTTP Request
				resp, body := testRequest(t, tt.request.httpMethod, tt.request.url, tt.request.body, tt.request.headers)
				defer resp.Body.Close() // go vet test

				// Проверяем все возможные хедеры для ожидаемого ответа, которые указали в тест кейсах выше.
				// Если хедер не указан, то будет возвращено пустое значение.
				// Возможно не самое элегантное решение, но пока я лучше не придумал :-/
				wantContentTypeHeader := tt.want.headers["Content-Type"]
				wantAcceptEncodingHeader := tt.want.headers["Accept-Encoding"]
				wantContentEncodingHeader := tt.want.headers["Content-Encoding"]

				// Если мы ожидаем сжатый в gzip ответ. Т.е. в HTTP Response пришел херед: "Content-Encoding: gzip"
				// В этом случае нужно распаковать из gzip body и сравнить результат с ожидаемым
				if resp.Header.Get("Content-Encoding") == "gzip" {
					assert.Equal(t, wantContentEncodingHeader, resp.Header.Get("Content-Encoding"))
					assert.Equal(t, tt.want.body, gzipDecompressor(body))
				} else {
					assert.Equal(t, wantContentTypeHeader, resp.Header.Get("Content-Type"))
					assert.Equal(t, wantAcceptEncodingHeader, resp.Header.Get("Accept-Encoding"))

					assert.Equal(t, tt.want.statusCode, resp.StatusCode)
					assert.Equal(t, tt.want.body, body)
				}
			})
		}
		ts.Close()
	}
	defer os.Remove("test_db.txt")
}

func TestController_GetUrlHandler(t *testing.T) {
	// Параметры для настройки тестового HTTP Request
	type request struct {
		httpMethod string
		url        string
		body       string
		headers    map[string]string
	}
	// Ожидаемый ответ сервера
	type want struct {
		headers    map[string]string
		statusCode int
		body       string
	}
	// Список тесткейсов
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "Prepare: Add test url into DB",
			request: request{
				httpMethod: http.MethodPost,
				url:        "http://127.0.0.1:8080",
				body:       "https://www.youtube.com/watch?v=09nmlZjxRFs",
			},
			want: want{
				headers:    map[string]string{"Content-Type": "text/plain"},
				statusCode: 201,
				body:       "http://127.0.0.1:8080/KJYUS",
			},
		},
		{
			name: "test_1: GET: Success short url from test #1",
			request: request{
				httpMethod: http.MethodGet,
				url:        "http://127.0.0.1:8080/KJYUS",
			},
			want: want{
				headers:    map[string]string{"Location": "https://www.youtube.com/watch?v=09nmlZjxRFs"},
				statusCode: http.StatusTemporaryRedirect,
			},
		},
		{
			name: "test_2: GET: Short url not found",
			request: request{
				httpMethod: http.MethodGet,
				url:        "http://127.0.0.1:8080/qqWW",
			},
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
	}

	// Прогоняем одинаковые тесты на разной конфигурации сервера: inMemoryDB, fileDB
	tsDBName := []string{"inMemoryDB", "fileDB"}
	for _, dbName := range tsDBName {
		ts := NewTestServer(dbName, "")
		ts.Start()
		for _, tt := range tests {
			testName := fmt.Sprintf("%s: DB: %s", tt.name, dbName)
			t.Run(testName, func(t *testing.T) {
				resp, _ := testRequest(t, tt.request.httpMethod, tt.request.url, tt.request.body, map[string]string{})
				defer resp.Body.Close() // go vet test from github

				wantLocationHeader := tt.want.headers["Location"]

				assert.Equal(t, tt.want.statusCode, resp.StatusCode)
				assert.Equal(t, wantLocationHeader, resp.Header.Get("Location"))
			})
		}
		ts.Close()
	}
	defer os.Remove("test_db.txt")
}

func TestController_DefaultHandler(t *testing.T) {
	// Параметры для настройки тестового HTTP Request
	type request struct {
		httpMethod string
		url        string
		body       string
	}
	// Ожидаемый ответ сервера
	type want struct {
		statusCode int
		body       string
	}
	// Список тесткейсов
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "test_1: Default: HTTP PUT request",
			request: request{
				httpMethod: http.MethodPut,
				url:        "http://127.0.0.1:8080",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "test_1: Default: HTTP Delete request",
			request: request{
				httpMethod: http.MethodDelete,
				url:        "http://127.0.0.1:8080",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	// Прогоняем одинаковые тесты на разной конфигурации сервера: inMemoryDB, fileDB
	tsDBName := []string{"inMemoryDB", "fileDB"}
	for _, dbName := range tsDBName {
		ts := NewTestServer(dbName, "")
		ts.Start()
		for _, tt := range tests {
			testName := fmt.Sprintf("%s: DB: %s", tt.name, dbName)
			t.Run(testName, func(t *testing.T) {
				resp, body := testRequest(t, tt.request.httpMethod, tt.request.url, tt.request.body, map[string]string{})
				defer resp.Body.Close() // go vet test from github

				assert.Equal(t, tt.want.statusCode, resp.StatusCode)
				assert.Equal(t, tt.want.body, body)
			})
		}
		ts.Close()
	}
	defer os.Remove("test_db.txt")
}
