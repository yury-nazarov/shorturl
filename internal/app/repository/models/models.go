package models

// Структуры для работы с БД

// Record - описывает каждую запись в БД как json
//			Используем:
//				repository.file 		- read / write to file
//				repository.inmemory  	- read / write to file
// 				repository.pg.GetUserURL - парсинг отваета SQL запроса
type Record struct {
	ShortURL  	string `json:"short_url"`
	OriginURL 	string `json:"original_url"`
	Token 		string `json:"token"`
}

// структуры для handler помогающие обрабатывать и сериализовать коммуникации с клиентом

// URL
//		десериализуем данные пришедшие по HTTP
// 		Так же с помощью этой структуры сериализуем ответ клиенту
type URL struct {
	Request  string `json:"url,omitempty"`    // Не учитываем поле при Marshal
	Response string `json:"result,omitempty"` // Не учитываем поле при Unmarshal
}

// URLBatch
// 		 десериализуем данные пришедшие по HTTP
// 		 Так же с помощью этой структуры сериализуем ответ клиенту
type URLBatch struct {
	CorrelationID 	string `json:"correlation_id"`
	OriginalURL 	string `json:"original_url,omitempty"`
	ShortURL 		string `json:"short_url,omitempty"`
}


// Вариант использовать пару общих структур для передаи данных между слоями
// 		   и сериализации/десериализации коммуникаций с пользователем.

////URLService структура представляет таблицу shorten_url в БД
//type URLService struct {
//	ID 		int
//	Origin 	string
//	Short  	string `json:"short_url"`
//	Owner 	string `json:"original_url"`
//	Delete 	bool
//}