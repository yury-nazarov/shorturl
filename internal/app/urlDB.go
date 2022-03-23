package app

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// URLDB - БД для URL
type URLDB struct {
	DB map[string]string
	ShortURLLength int
}

// Add Добавляет новый url в БД
func (u *URLDB) Add(longUrl string) string {
	urlID := u.shortURL(longUrl)
	u.DB[urlID] = longUrl
	return urlID
}

// Get Достает из БД URL
func (u *URLDB) Get(urlId string) (string, error) {
	longUrl, ok := u.DB[urlId]
	if !ok {
		return "", fmt.Errorf("url %s not found", urlId)
	}
	return longUrl, nil
}

// shortUrl наивный алгоритм сокращения url:
// 			считаем md5 хеш от оригинального url
//			Дальше забираем из него рандомно 5 символов, собираем в строку
// 			для избежания коллизий дополнительно в цикле проверяем наличие shortUrl в db map
func (u *URLDB) shortURL(longURL string) string {
	for {
		rand.Seed(time.Now().Unix())
		// Считаем md5
		charSet := fmt.Sprintf("%x", md5.Sum([]byte(longURL)))
		var shortUrl []string
		// Достаем N рандомных символов из хеша собираем в строку
		for i := 0; i < u.ShortURLLength; i++ {
			random := rand.Intn(len(charSet))
			randomChar := charSet[random]
			shortUrl = append(shortUrl, string(randomChar))
		}
		res := strings.Join(shortUrl, "")
		// Проверяем наличие сгенерированого url в БД, если нет - возвращаем
		if u.urlIsPresent(res) {
			return res
		}
	}
}

// urlIsPresent проверяет наличие укороченого url в БД
func (u *URLDB) urlIsPresent(shortUrl string) bool {
	_, ok := u.DB[shortUrl]
	// TODO: go vet test / statictest
	// 		 should use 'return !ok' instead of 'if !ok { return true }; return false'
	// 		 Пока не больше путает, чем помогает. Спросить у ментора.
	if !ok {
		return !ok
	}
	return ok
}
