package filedb

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/yury-nazarov/shorturl/internal/app/repository/models"
)

// Хранение данных в файле

type fileDB struct {
	name string
}

// NewFileDB - 	возвращает объект для записи и чтения из файла БД
//			   	Метод Add() открывает файл, записывает новую строчку
// 				Метод Get() ищет нужное значение в файле и возвращает его
func NewFileDB(fileName string) *fileDB {
	f := &fileDB{
		name: fileName,
	}
	return f
}

// Add - добавляем запись в БД
func (f *fileDB) Add(ctx context.Context, shortURL string, originURL string, token string) error {
	// Создаем новую запись как JSON объект
	data := &models.Record{
		ShortURL:  shortURL,
		OriginURL: originURL,
		Token:     token,
	}
	// Открываем файл на запись
	p, err := newProducer(f.name)
	if err != nil {
		return err
	}
	// go vet test: should check returned error before deferring p.Close()
	defer func() {
		if err = p.close(); err != nil {
			log.Print(err)
		}
	}()
	// Записываем новую строку
	if err = p.write(data); err != nil {
		return err
	}
	return nil
}

// Get Поиск в БД
func (f *fileDB) Get(ctx context.Context, shortURL string, token string) (string, error) {
	// Открываем файл на чтение
	c, err := newConsumer(f.name)
	if err != nil {
		return "", err
	}
	defer c.close()
	// В цикле читаем каждую запись
	for {
		r, err := c.read()

		if err == io.EOF {
			return "", fmt.Errorf("the URL not found")
		}
		if r.ShortURL == shortURL {
			return r.OriginURL, nil
		}
		if err != nil {
			return "", err
		}
	}
}

func (f *fileDB) GetToken(ctx context.Context, token string) (bool, error) {
	// Открываем файл на чтение
	c, err := newConsumer(f.name)
	if err != nil {
		return false, err
	}
	defer c.close()
	// В цикле читаем каждую запись
	for {
		r, err := c.read()
		if err == io.EOF {
			return false, fmt.Errorf("the URL not found")
		}
		if r.Token == token {
			return true, nil
		}
		if err != nil {
			return false, err
		}
	}
}

// GetUserURL - вернет слайс из структур со всем URL пользователя
func (f *fileDB) GetUserURL(ctx context.Context, token string) ([]models.Record, error) {
	// Открываем файл на чтение
	c, err := newConsumer(f.name)
	if err != nil {
		return []models.Record{}, err
	}
	defer c.close()
	// В цикле читаем каждую запись
	var result []models.Record
	for {
		r, err := c.read()

		if err == io.EOF {
			break
		}

		if r.Token == token {
			result = append(result, models.Record{ShortURL: r.ShortURL, OriginURL: r.OriginURL})
		}
	}
	return result, nil
}

// Ping Для обратной совместимости с Postgres
func (f *fileDB) Ping() bool {
	return true
}

// OriginURLExists Для обратной совместимости с Postgres
func (f *fileDB) OriginURLExists(ctx context.Context, originURL string) (bool, error) {
	return false, nil
}

//// GetOwnerToken Для обратной совместимости с Postgres
//func (f *fileDB) GetOwnerToken(ctx context.Context, token string) repository.Owner {
//	owner := repository.Owner{}
//	return owner
//}

// GetShortURLByIdentityPath Для обратной совместимости с Postgres
func (f *fileDB) GetShortURLByIdentityPath(ctx context.Context, identityPath string, token string) int {
	return 0
}

// URLBulkDelete Для обратной совместимости с Postgres
//func (f *fileDB) URLBulkDelete(ctx context.Context, idList []int) error {
func (f *fileDB) URLBulkDelete(ctx context.Context, urlsID chan int) error {
	return nil
}
