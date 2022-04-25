package repository

// Repository - общее представление интерфейса для работы с БД
type Repository interface {
	Add(shortPath string, longURL string, token string) error
	Get(shortPath string) (string, error)
	GetToken(token string) (bool, error)
	GetUserURL(token string) ([]RecordURL, error)
	Ping() bool
}

// RecordURL - Структура описывает возращаемые занчения для пакета repository
type RecordURL struct {
	ShortURL  	string `json:"short_url"`
	OriginURL 	string `json:"original_url"`
}