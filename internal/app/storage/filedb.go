package storage

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
)

type Record struct {
	ShortURL 	string 	`json:"short_url"`
	OriginURL 	string 	`json:"origin_url"`
}

type fileDB struct {
	name string
}

// NewFileDB - читает fileName и загружает в RAM
func NewFileDB(fileName string) *fileDB{
	f := &fileDB{
		name: fileName,
	}
	return f
}

// Add - добавляем запись в БД
func (f *fileDB) Add(shortURL string, originURL string) error{
	// Создаем новую запись
	data := &Record{
		ShortURL: shortURL,
		OriginURL: originURL,
	}
	// Открываем файл на запись
	p, err := NewProducer(f.name)
	if err != nil {
		return err
	}
	// go vet test: should check returned error before deferring p.Close()
	defer func() {
		if err = p.Close(); err != nil {
			log.Print(err)
		}
	}()
	// Новыю строку
	if err = p.Write(data); err != nil {
		return err
	}
	return nil
}


// Get Поиск в БД
func (f *fileDB) Get(shortPath string) (string, error){
	// Открываем файл на чтение
	c, err := NewConsumer(f.name)
	if err != nil {
		return "", err
	}
	defer c.Close()
	// В цикле читаем каждую запись
	for {
		r, err := c.Read()

		if err == io.EOF {
			return "", errors.New("the URL not found")
		}
		if r.ShortURL == shortPath {
			return r.OriginURL, nil
		}
		if err != nil {
			return "", err
		}
	}
}



type producer struct {
	file *os.File
	encoder *json.Encoder
}

func NewProducer(fileName string) (*producer, error){
	file, err := os.OpenFile(fileName, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	return &producer{
		file: file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *producer) Write(record *Record) error{
	return p.encoder.Encode(&record)
}

func (p *producer) Close() error{
	return p.file.Close()
}




type consumer struct {
	file *os.File
	decoder *json.Decoder
}

func NewConsumer(fileName string) (*consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY | os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &consumer{
		file: file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *consumer) Read() (*Record, error){
	record := &Record{}
	if err := c.decoder.Decode(&record); err != nil {
		return nil, err
	}
	return record, nil
}

func (c *consumer) Close() error {
	return c.file.Close()
}