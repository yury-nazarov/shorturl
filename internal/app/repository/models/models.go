package models



// URL  Описывает:
// 		пришедшие от пользователя данные в поле Request.
//		отправляемые пользователю данные в поле Response.
// 		Используется в пакете: handler.
type URL struct {
	Request  string `json:"url,omitempty"`    // Не учитываем поле при Marshal
	Response string `json:"result,omitempty"` // Не учитываем поле при Unmarshal
}

// URLBatch Описывает:
//		пришедшие от пользователя данные в полях: CorrelationID, OriginalURL.
//		отправляемые пользователю данные в полях: CorrelationID, ShortURL.
// 		Используется в пакете: handler.
type URLBatch struct {
	CorrelationID 	string `json:"correlation_id"`
	OriginalURL 	string `json:"original_url,omitempty"`
	ShortURL 		string `json:"short_url,omitempty"`
}

// Структуры для работы с БД

// Record - Описывает:
//              структуру данных при работе с БД.
//          Используется в пакетах:
//              repository.file             - метод GetUserURL, что бы вернуть все url пользователя,
//                                            а так же для чтения и записи в файл в формате JSON.
//              repository.inmemory         - метод GetUserURL, что бы вернуть все url пользователя.
//              repository.pg.GetUserURL    - метод GetUserURL, что бы вернуть все url пользователя.
type Record struct {
	ShortURL  	string `json:"short_url"`
	OriginURL 	string `json:"original_url"`
	Token 		string `json:"token"`
}
