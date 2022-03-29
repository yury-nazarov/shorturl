package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yury-nazarov/shorturl/internal/app/storage"
)

func TestURLService_URLHandler(t *testing.T) {
	type request struct {
		httpMethod string
		url string
		body string
	}
	type want struct {
		contentType string
		statusCode int
		body string
	}
	tests := []struct{
		name string
		request request
		want want
	}{
		{
			name: "test #1: Success request",
			request: request{
				httpMethod: http.MethodPost,
				url: "/",
				body: "https://www.youtube.com/watch?v=09nmlZjxRFs",
			},
			want: want {
				contentType: "text/plain",
				statusCode: 201,
				body: "http://127.0.0.1:8080/KJYUS",
			},
		},
		{
			name: "test #2: Empty body",
			request: request{
				httpMethod: http.MethodPost,
				url: "/",
				body: "",
			},
			want: want {
				contentType: "",
				statusCode: 400,
				body: "",
			},
		},
	}

	db := storage.New()
	s := New("127.0.0.1", 8080, 5, db)

	for _, tt := range tests{
		t.Run(tt.name, func(t *testing.T){
			request := httptest.NewRequest(tt.request.httpMethod, tt.request.url, strings.NewReader(tt.request.body))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.URLHandler)
			h.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode) // 200
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
