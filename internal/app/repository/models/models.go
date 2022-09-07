package models

// из NewDB

// Record - описывает каждую запись в БД как json
//			Используем:
//				repository.file 		- read / write to file
//				repository.inmemory  	- read / write to file
// 				repository.pg.GetUserURL - парсинг отваета SQL запроса
type Record struct {
	ShortURL  	string `json:"short_url"`
	//OriginURL 	string `json:"origin_url"`
	OriginURL 	string `json:"original_url"`
	Token 		string `json:"token"`
}

//

// RecordURL - Структура описывает возращаемые занчения для пакета repository
//             Используем:
//					pg.GetUserURL - парсинг отваета SQL запроса
//type RecordURL struct {
//	ShortURL  	string `json:"short_url"`
//	OriginURL 	string `json:"original_url"`
//}

// Owner Представление таблицы shorten_url.owner
// 		 используем
//				 pg.GetToken - парсинг отваета SQL запроса
type Owner struct {
	ID    int
	Token string
}

// структуры из handler

type URL struct {
	Request  string `json:"url,omitempty"`    // Не учитываем поле при Marshal
	Response string `json:"result,omitempty"` // Не учитываем поле при Unmarshal
}


type URLBatch struct {
	CorrelationID 	string `json:"correlation_id"`
	OriginalURL 	string `json:"original_url,omitempty"`
	ShortURL 		string `json:"short_url,omitempty"`
}


// Вариант использовать пару общих структур для передаи данных между слоями
// 		   и сериализации/десериализации коммуникаций с пользователем.

// URLService структура представляет таблицу shorten_url в БД
//type URLService struct {
//	ID 		int
//	Origin 	string
//	Short  	string `json:"short_url"`
//	Owner 	string `json:"original_url"`
//	Delete 	bool
//}