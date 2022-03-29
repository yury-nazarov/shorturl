package server

import (
	"fmt"
	"net/http"

	"github.com/yury-nazarov/shorturl/internal/app/storage"
)

type URLService struct {
	ListenAddress 	string
	Port 			int
	URLLength 		int
	DB 				*storage.URLDB
	Bind 			*http.Server
}

// New - конструктор обертка для стандартного http сервера,
// 		 необходим что бы внутри объекта были поля доступа к методам storage и конфигурационным параметрам запуска приложения
func New(ListenAddress string, Port int, URLLength int, db *storage.URLDB, router http.Handler) *URLService{

	s := &URLService{
		DB: db,
		URLLength: URLLength,
		ListenAddress: ListenAddress,
		Port: Port,
		Bind: &http.Server{
			Addr: fmt.Sprintf("%s:%d", ListenAddress, Port),
			Handler: router,
		},
	}
	return s
}

