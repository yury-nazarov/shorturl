package storage

// Repository - общее представление интерфейса для работы с БД
type Repository interface {
	Add(shortPath string, longURL string)
	Get(shortPath string) (string, error)
}