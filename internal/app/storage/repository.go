package storage

// Repository - общее представление интерфейса для работы с БД
type Repository interface {
	Add(shortPath string, longURL string) error
	Get(shortPath string) (string, error)
}