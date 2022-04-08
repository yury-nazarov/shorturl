package filedb

import (
	"encoding/json"
	"os"
)

// Запись данных в файл

type producer struct {
	file *os.File
	encoder *json.Encoder
}

func newProducer(fileName string) (*producer, error){
	file, err := os.OpenFile(fileName, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	return &producer{
		file: file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *producer) write(record *record) error{
	return p.encoder.Encode(&record)
}

func (p *producer) close() error{
	return p.file.Close()
}
