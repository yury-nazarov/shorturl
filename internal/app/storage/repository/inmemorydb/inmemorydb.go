package inmemorydb

import (
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
func (u *inMemoryDB) Get(shortURL string) (string, error) {
	urlInfo, ok := u.db[shortURL]
	if !ok {
		return "", fmt.Errorf("shorturl %s not found", shortURL)
	}
	return urlInfo.longURL, nil
}

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


func (u *inMemoryDB) Ping() bool {
	return true
}

func (u *inMemoryDB) OriginURLExists(originURL string) (bool, error) {
	return false, nil
}