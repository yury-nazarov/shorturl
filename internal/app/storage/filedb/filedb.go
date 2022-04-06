package filedb

import (
	"errors"
	"io"
	"log"
)

// Record - описывает каждую запись в БД как json
type record struct {
	ShortURL 	string 	`json:"short_url"`
	OriginURL 	string 	`json:"origin_url"`
}

type fileDB struct {
	name string
}

// NewFileDB - 	возвращает объект
//			   	Метод Add() открывает файл, записывает новую строчку
// 				Метод Get() ищет нужное значение в файле и возвращает его
func NewFileDB(fileName string) *fileDB {
	f := &fileDB{
		name: fileName,
	}
	return f
}

// Add - добавляем запись в БД
func (f *fileDB) Add(shortURL string, originURL string) error{
	// Создаем новую запись
	data := &record{
		ShortURL: shortURL,
		OriginURL: originURL,
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
	// Новыю строку
	if err = p.write(data); err != nil {
		return err
	}
	return nil
}


// Get Поиск в БД
func (f *fileDB) Get(shortURL string) (string, error){
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
			return "", errors.New("the URL not found")
		}
		if r.ShortURL == shortURL {
			return r.OriginURL, nil
		}
		if err != nil {
			return "", err
		}
	}
}
