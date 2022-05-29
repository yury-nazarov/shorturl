package repository

import "context"

// Repository - общее представление интерфейса для работы с БД
type Repository interface {
	Add(ctx context.Context, shortURL string, longURL string, token string) error
	Get(ctx context.Context, shortURL string, token string) (string, error)
	GetToken(ctx context.Context, token string) (bool, error)
	GetUserURL(ctx context.Context, token string) ([]RecordURL, error)
	GetShortURLByIdentityPath(ctx context.Context, identityPath string, token string) int
	URLBulkDelete(ctx context.Context, urlsID chan int) error
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