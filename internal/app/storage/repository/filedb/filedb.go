package filedb

import (
	"fmt"
	"io"
	"log"

	"github.com/yury-nazarov/shorturl/internal/app/storage/repository"
)

// Хранение данных в файле

// record - описывает каждую запись в БД как json
type record struct {
	ShortURL  	string `json:"short_url"`
	OriginURL 	string `json:"origin_url"`
	Token 		string `json:"token"`
}

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
func (f *fileDB) Add(shortURL string, originURL string, token string) error {
	// Создаем новую запись как JSON объект
	data := &record{
		ShortURL:  shortURL,
		OriginURL: originURL,
		Token: token,
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
func (f *fileDB) Get(shortURL string) (string, error) {
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


func (f *fileDB) GetToken(token string) (bool, error) {
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
func (f *fileDB) GetUserURL(token string) ([]repository.RecordURL, error) {
	// Открываем файл на чтение
	c, err := newConsumer(f.name)
	if err != nil {
		return []repository.RecordURL{}, err
	}
	defer c.close()
	// В цикле читаем каждую запись
	var result []repository.RecordURL
	for {
		r, err := c.read()

		if err == io.EOF {
			break
		}

		if r.Token == token {
			result = append(result, repository.RecordURL{ShortURL: r.ShortURL, OriginURL: r.OriginURL})
		}
	}
	return result, nil
}

func (f *fileDB) Ping() bool {
	// TODO: Создать тестовый файл, записать в него, удалить
	return true
}