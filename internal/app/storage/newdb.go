package storage

import (
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository/filedb"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository/inmemorydb"
)


type DBConfig struct {
	FileName 	string
}

// Record - описывает каждую запись в БД как json
type Record struct {
	ShortURL  	string `json:"short_url"`
	OriginURL 	string `json:"origin_url"`
	Token 		string `json:"token"`
}

func New(conf DBConfig) repository.Repository {
	if len(conf.FileName) != 0 {
		return filedb.NewFileDB(conf.FileName)
	}
	return inmemorydb.NewInMemoryDB()
}
