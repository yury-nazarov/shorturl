package app

import (
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// UrlDB - БД для URL
type UrlDB struct {
	Db map[string]string
	ShortURLLength int
}

// Add Добавляет новый url в БД
func (udb *UrlDB) Add(longUrl string) string {
	urlId := udb.shortUrl(longUrl)
	udb.Db[urlId] = longUrl
	return urlId
}

// Get Достает из БД URL
func (udb *UrlDB) Get(urlId string) (string, error) {
	longUrl, ok := udb.Db[urlId]
	if !ok {
		return "", errors.New(fmt.Sprintf("url %s not found", urlId))
	}
	return longUrl, nil
}

// shortUrl наивный алгоритм сокращения url:
// 			считаем md5 хеш от оригинального url
//			Дальше забираем из него рандомно 5 символов, собираем в строку
// 			для избежания коллизий дополнительно в цикле проверяем наличие shortUrl в db map
func (udb *UrlDB) shortUrl(longUrl string) string {
	for {
		rand.Seed(time.Now().Unix())
		// Считаем md5
		charSet := fmt.Sprintf("%x", md5.Sum([]byte(longUrl)))
		var shortUrl []string
		// Достаем N рандомных символов из хеша собираем в строку
		for i := 0; i < udb.ShortURLLength; i++ {
			random := rand.Intn(len(charSet))
			randomChar := charSet[random]
			shortUrl = append(shortUrl, string(randomChar))
		}
		res := strings.Join(shortUrl, "")
		// Проверяем наличие сгенерированого url в БД, если нет - возвращаем
		if udb.urlIsPresent(res) {
			return res
		}
	}
}

// urlIsPresent проверяет наличие укороченого url в БД
func (i *UrlDB) urlIsPresent(shortUrl string) bool {
	_, ok := i.Db[shortUrl]
	if !ok {
		return true
	}
	return false
}
