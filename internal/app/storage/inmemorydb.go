package storage

import (
	"fmt"
)

// InMemoryDB - БД для URL
type inMemoryDB struct {
	db map[string]string
}

func NewInMemoryDB() *inMemoryDB{
	db := &inMemoryDB{
		db: map[string]string{},
	}
	return db
}

// Add Добавляет новый url в БД
func (u *inMemoryDB) Add(shortPath string, longURL string) {
	u.db[shortPath] = longURL
}

// Get Достает из БД URL
func (u *inMemoryDB) Get(shortPath string) (string, error) {
	originURL, ok := u.db[shortPath]
	if !ok {
		return "", fmt.Errorf("shorturl %s not found", shortPath)
	}
	return originURL, nil
}
