package models

// из NewDB

// Record - описывает каждую запись в БД как json
type Record struct {
	ShortURL  	string `json:"short_url"`
	OriginURL 	string `json:"origin_url"`
	Token 		string `json:"token"`
}


//

// RecordURL - Структура описывает возращаемые занчения для пакета repository
type RecordURL struct {
	ShortURL  	string `json:"short_url"`
	OriginURL 	string `json:"original_url"`
}

// Owner Представление таблицы shorten_url.owner
type Owner struct {
	ID    int
	Token string
}

// URLService структура представляет таблицу shorten_url в БД
//type URLService struct {
//	ID 		int
//	Origin 	string
//	Short  	string `json:"short_url"`
//	Owner 	string `json:"original_url"`
//	Delete 	bool
//}