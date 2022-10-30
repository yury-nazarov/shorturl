package filedb

import (
	"encoding/json"
	"os"

	"github.com/yury-nazarov/shorturl/internal/app/repository/models"
)

// Чтение данных из файла

type consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func newConsumer(fileName string) (*consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *consumer) read() (*models.Record, error) {
	record := &models.Record{}
	if err := c.decoder.Decode(&record); err != nil {
		return nil, err
	}
	return record, nil
}

func (c *consumer) close() error {
	return c.file.Close()
}
