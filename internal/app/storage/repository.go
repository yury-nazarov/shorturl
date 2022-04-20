package storage

// Repository - общее представление интерфейса для работы с БД
type Repository interface {
	Add(shortPath string, longURL string, token string) error
	Get(shortPath string) (string, error)
	GetToken(token string) (bool, error)
}