package inmemorydb

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/yury-nazarov/shorturl/internal/app/repository/models"
)

// URLInfo вспомогательная структура для записи JSON в мапу.
type URLInfo struct {
	longURL string
	token   string
}

type inMemoryDB struct {
	db map[string]URLInfo
	mu sync.RWMutex
}

// NewInMemoryDB - 	возвращает объект для записи и чтения из мапы в RAM.
func NewInMemoryDB() *inMemoryDB {
	db := &inMemoryDB{
		db: map[string]URLInfo{},
	}
	return db
}

// Add Добавляет новый url в БД.
func (u *inMemoryDB) Add(ctx context.Context, shortURL string, longURL string, token string) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.db[shortURL] = URLInfo{longURL: longURL, token: token}
	return nil
}

// Get Достает из БД URL.
func (u *inMemoryDB) Get(ctx context.Context, shortURL string, token string) (string, error) {
	urlInfo, ok := u.db[shortURL]
	if !ok {
		return "", fmt.Errorf("shorturl %s not found", shortURL)
	}
	return urlInfo.longURL, nil
}

// GetToken за O(n) ищет первую подходящую запись с токеном.
func (u *inMemoryDB) GetToken(ctx context.Context, token string) (bool, error) {
	for _, urlInfo := range u.db {
		//if strings.Contains(token, urlInfo.token) {
		if token == urlInfo.token {
			return true, nil
		}
	}
	return false, nil
}

// GetUserURL - вернет все url для пользователя.
func (u *inMemoryDB) GetUserURL(ctx context.Context, token string) ([]models.Record, error) {
	var result []models.Record
	for k, urlInfo := range u.db {
		if strings.Contains(token, urlInfo.token) {
			result = append(result, models.Record{ShortURL: k, OriginURL: urlInfo.longURL})
		}
	}
	return result, nil
}

// Ping Для обратной совместимости с Postgres.
func (u *inMemoryDB) Ping() bool {
	return true
}

// OriginURLExists Для обратной совместимости с Postgres.
func (u *inMemoryDB) OriginURLExists(ctx context.Context, originURL string) (bool, error) {
	return false, nil
}

// GetShortURLByIdentityPath Для обратной совместимости с Postgres.
func (u *inMemoryDB) GetShortURLByIdentityPath(ctx context.Context, identityPath string, token string) int {
	return 0
}

// URLMarkDeleted Для обратной совместимости с Postgres.
func (u *inMemoryDB) URLMarkDeleted(ctx context.Context, id int) {}

// URLBulkDelete Для обратной совместимости с Postgres.
func (u *inMemoryDB) URLBulkDelete(ctx context.Context, urlsID chan int) error {
	return nil
}
