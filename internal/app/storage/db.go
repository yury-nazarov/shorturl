package storage

import (
	"fmt"
)

// URLDB - БД для URL
type URLDB struct {
	DB map[string]string
}

func New() *URLDB{
	db := &URLDB{
		DB: map[string]string{},
	}
	return db
}

// Add Добавляет новый url в БД
func (u *URLDB) Add(shortPath string, longURL string) {
	u.DB[shortPath] = longURL
	//return shortPath
}

// Get Достает из БД URL
func (u *URLDB) Get(shortPath string) (string, error) {
	originURL, ok := u.DB[shortPath]
	if !ok {
		return "", fmt.Errorf("shorturl %s not found", shortPath)
	}
	return originURL, nil
}
