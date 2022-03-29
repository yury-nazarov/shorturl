package server

import (
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yury-nazarov/shorturl/internal/app/storage"
)

func TestURLService_URLHandler(t *testing.T) {
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
			name: "test #1: POST: Success request",
			request: request{
				httpMethod: http.MethodPost,
				url: "/",
				body: "https://www.youtube.com/watch?v=09nmlZjxRFs",
			},
			want: want {
				header: header{ contentType: "text/plain" },
				statusCode: 201,
				body: "http://127.0.0.1:8080/KJYUS",
			},
		},
		{
			name: "test #2: POST: Empty body",
			request: request{
				httpMethod: http.MethodPost,
				url: "/",
			},
			want: want {
				statusCode: http.StatusBadRequest,
				body: "",
			},
		},
		{
			name: "test #3: GET: Success short url from test #1",
			request: request{
				httpMethod: http.MethodGet,
				url: "/KJYUS",
			},
			want: want {
				header: header{locations: "https://www.youtube.com/watch?v=09nmlZjxRFs"},
				statusCode: http.StatusTemporaryRedirect,
			},
		},
		{
			name: "test #4: GET: Short url not found",
			request: request{
				httpMethod: http.MethodGet,
				url: "/qqWW",
			},
			want: want {
				//header: header{contentType: "text/plain; charset=utf-8"},
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "test #5: Default: Some other method without POST, GET",
			request: request{
				httpMethod: http.MethodPut,
				url: "/KJYUS",
			},
			want: want {
				//header: header{contentType: "text/plain; charset=utf-8"},
				statusCode: http.StatusBadRequest,
			},
		},
	}

	db := storage.New()
	s := New("127.0.0.1", 8080, 5, db)

	for _, tt := range tests{
		t.Run(tt.name, func(t *testing.T){
			request := httptest.NewRequest(tt.request.httpMethod, tt.request.url, strings.NewReader(tt.request.body))
			defer request.Body.Close()

			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.URLHandler)
			h.ServeHTTP(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.header.contentType, result.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.header.locations, result.Header.Get("Location"))
			resultBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.body, string(resultBody))

		})
	}
}
