package repository

import "context"

// Repository - общее представление интерфейса для работы с БД
type Repository interface {
	Add(shortURL string, longURL string, token string) error
	Get(shortURL string, token string) (string, error)
	GetToken(token string) (bool, error)
	GetOwnerToken(token string) Owner
	GetUserURL(token string) ([]RecordURL, error)
	GetShortURLByIdentityPath(identityPath string, token string) int
	URLMarkDeleted(id int)
	Ping() bool
	OriginURLExists(ctx context.Context, originURL string) (bool, error)
}

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