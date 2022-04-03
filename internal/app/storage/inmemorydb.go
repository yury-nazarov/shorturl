package storage

import (
	"fmt"
)

// InMemoryDB - БД для URL
type InMemoryDB struct {
	DB map[string]string
}

func NewInMemoryDB() *InMemoryDB{
	db := &InMemoryDB{
		DB: map[string]string{},
	}
	return db
}

// Add Добавляет новый url в БД
func (u *InMemoryDB) Add(shortPath string, longURL string) {
	u.DB[shortPath] = longURL
	//return shortPath
}

// Get Достает из БД URL
func (u *InMemoryDB) Get(shortPath string) (string, error) {
	originURL, ok := u.DB[shortPath]
	if !ok {
		return "", fmt.Errorf("shorturl %s not found", shortPath)
	}
	return originURL, nil
}
