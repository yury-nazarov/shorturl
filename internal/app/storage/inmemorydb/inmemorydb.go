package inmemorydb

import (
	"fmt"
	"strings"
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
func (u *inMemoryDB) Add(shortPath string, longURL string, token string) error {
	u.db[shortPath] = URLInfo{longURL: longURL, token: token}
	return nil
}

// Get Достает из БД URL
func (u *inMemoryDB) Get(shortPath string) (string, error) {
	urlInfo, ok := u.db[shortPath]
	if !ok {
		return "", fmt.Errorf("shorturl %s not found", shortPath)
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