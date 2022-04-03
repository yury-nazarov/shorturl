package handler

import (
	"fmt"
	"io/ioutil"
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

func NewTestServer() *httptest.Server{
	router := chi.NewRouter()
	db := storage.NewInMemoryDB()
	lc := service.NewLinkCompressor(5, "")
	c := NewController(db, lc)

	router.HandleFunc("/", c.DefaultHandler)
	router.Get("/{urlID}", c.GetURLHandler)
	router.Post("/", c.AddURLHandler)

	return httptest.NewServer(router)
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

func TestController_AddUrlHandler(t *testing.T) {
	type header struct {
		contentType string
		locations string
	}
	type request struct {
		httpMethod string
		url string
		body string
	}
	type want struct {
		statusCode int
		body string
	}
	tests := []struct{
		name string
		request request
		want want
	}{
		{
			name: "test_1: POST: Success request",
			request: request{
				httpMethod: http.MethodPost,
				url:        "/",
				body:       "https://www.youtube.com/watch?v=09nmlZjxRFs",
			},
			want: want{
				statusCode: 201,
				body: "/KJYUS",
			},
		},
		{
			name: "test_2: POST: Empty body",
			request: request{
				httpMethod: http.MethodPost,
				url: "/",
			},
			want: want {
				statusCode: http.StatusBadRequest,
				body: "",
			},
		},
	}

	ts := NewTestServer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s%s", ts.URL, tt.request.url)
			resp, body := testRequest(t, tt.request.httpMethod, url, tt.request.body)
			defer resp.Body.Close() // go vet test
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			// При добавлении URL сервис возвращает ответ в виде короткой ссылки с адресом сервиса: "http://{ SERVICE_NAME }/KJYUS"
			// Предпосылки специфичные для тест кейсов:
			//		1. Тестовый сервер запускается на произвольном порту: http://127.0.0.1:XXXXX,
			//		   я не могу захардкодить конкретный URL, передав его в конструктор NewLinkCompressor().
			//		2. Не могу использовать локальный DNS для поддержки доменного имени для сервиса, в рамках текущих тесткейсов.
			// Тем самым, проверяем, если ожидаемый body не пустой (т.е. должен содержать URL в ответе),
			// то подставляем hostname:port тестового сервиса для подготовки ожидаемого клиентов URL: "http://{ SERVICE_NAME }/KJYUS"
			if len(tt.want.body) > 0 {
				serviceResponse := fmt.Sprintf("%s%s", ts.URL, body)
				wantURL := fmt.Sprintf("%s%s", ts.URL, tt.want.body)
				assert.Equal(t, wantURL,  serviceResponse)
			}
		})
	}
}

func TestController_GetUrlHandler(t *testing.T) {
	type header struct {
		contentType string
		locations string
	}
	type request struct {
		httpMethod string
		url string
		body string
	}
	type want struct {
		header header
		statusCode int
		body string
	}
	tests := []struct{
		name string
		request request
		want want
	}{
		{
			name: "Prepare: Add test url into DB",
			request: request{
				httpMethod: http.MethodPost,
				url:        "/",
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
				url: "/KJYUS",
			},
			want: want {
				header: header{locations: "https://www.youtube.com/watch?v=09nmlZjxRFs"},
				statusCode: http.StatusTemporaryRedirect,
				//statusCode: http.StatusOK,
			},
		},
		{
			name: "test_2: GET: Short url not found",
			request: request{
				httpMethod: http.MethodGet,
				url: "/qqWW",
			},
			want: want {
				statusCode: http.StatusNotFound,
			},
		},
	}

	ts := NewTestServer()


	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s%s", ts.URL, tt.request.url)
			resp, _ := testRequest(t, tt.request.httpMethod, url, tt.request.body)
			//client.CheckRedirect = func(req *Request, via []*Request) error{
			//	return nil
			//}
			defer resp.Body.Close() // go vet test from github
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}


func TestController_DefaultHandler(t *testing.T) {
	type header struct {
		contentType string
		locations string
	}
	type request struct {
		httpMethod string
		url string
		body string
	}
	type want struct {
		header header
		statusCode int
		body string
	}
	tests := []struct{
		name string
		request request
		want want
	}{
		{
			name: "test_1: Default: Some other method without POST, GET",
			request: request{
				httpMethod: http.MethodPut,
				url: "/",
			},
			want: want {
				statusCode: http.StatusBadRequest,
			},
		},
	}

	ts := NewTestServer()

	for _, tt := range tests{
		t.Run(tt.name, func(t *testing.T){
			url := fmt.Sprintf("%s%s", ts.URL, tt.request.url)
			resp, body := testRequest(t, tt.request.httpMethod, url, tt.request.body)
			defer resp.Body.Close() // go vet test from github
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.body, body)
		})
	}
}

