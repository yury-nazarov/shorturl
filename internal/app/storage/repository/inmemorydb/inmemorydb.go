package inmemorydb

import (
	"context"
	"fmt"
	"strings"

	"github.com/yury-nazarov/shorturl/internal/app/storage/repository"
)

// InMemoryDB - БД для URL

type URLInfo struct {
	longURL string
	token string
}

type inMemoryDB struct {
	db map[string]URLInfo
}


func NewInMemoryDB() *inMemoryDB {
	db := &inMemoryDB{
		db: map[string]URLInfo{},
	}
	return db
}

// Add Добавляет новый url в БД
func (u *inMemoryDB) Add(shortURL string, longURL string, token string) error {
	u.db[shortURL] = URLInfo{longURL: longURL, token: token}
	return nil
}

// Get Достает из БД URL
func (u *inMemoryDB) Get(shortURL string, token string) (string, error) {
	urlInfo, ok := u.db[shortURL]
	if !ok {
		return "", fmt.Errorf("shorturl %s not found", shortURL)
	}
	return urlInfo.longURL, nil
}
//func (u *inMemoryDB) Get(shortURL string) (repository.URL, error) {
//	var url repository.URL
//	urlInfo, ok := u.db[shortURL]
//	url.Origin = urlInfo.longURL
//	if !ok {
//		return url, fmt.Errorf("shorturl %s not found", shortURL)
//	}
//	return url, nil
//}

// GetToken за O(n) ищет первую подходящую запись с токеном
func (u *inMemoryDB) GetToken(token string) (bool, error) {
	for _, urlInfo := range u.db {
		if strings.Contains(token, urlInfo.token) {
			return true, nil
		}
	}
	return false, nil
}


// GetUserURL - вернет все url для пользователя
func (u *inMemoryDB) GetUserURL(token string) ([]repository.RecordURL, error) {
	var result []repository.RecordURL
	for k, urlInfo := range u.db {
		// fmt.Println("inmemoryItem",  k, urlInfo)
		if strings.Contains(token, urlInfo.token) {
			result = append(result, repository.RecordURL{ShortURL: k, OriginURL: urlInfo.longURL})
		}
	}
	return result, nil
}

// Ping Для обратной совместимости с Postgres
func (u *inMemoryDB) Ping() bool {
	return true
}

// OriginURLExists Для обратной совместимости с Postgres
func (u *inMemoryDB) OriginURLExists(ctx context.Context, originURL string) (bool, error) {
	return false, nil
}

// GetOwnerToken Для обратной совместимости с Postgres
func (u *inMemoryDB) GetOwnerToken(token string) repository.Owner {
	owner := repository.Owner{}

	return owner
}

// GetShortURLByIdentityPath Для обратной совместимости с Postgres
func (u *inMemoryDB) GetShortURLByIdentityPath(identityPath string, token string) int {
	return 0
}

// URLMarkDeleted Для обратной совместимости с Postgres
func (u *inMemoryDB) URLMarkDeleted(id int) {}