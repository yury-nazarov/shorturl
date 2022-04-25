package storage

import (
	"context"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository/filedb"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository/inmemorydb"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository/pg"
)


type DBConfig struct {
	FileName 	string
	PGConnStr	string
	Ctx 		context.Context
}

// Record - описывает каждую запись в БД как json
type Record struct {
	ShortURL  	string `json:"short_url"`
	OriginURL 	string `json:"origin_url"`
	Token 		string `json:"token"`
}

// New - возвращает подключение к БД, приоритеты:
//		 1. Postgres
//		 2. FileDB
//		 3. Inmemory
func New(conf DBConfig) repository.Repository {
	if len(conf.PGConnStr) != 0 {
		return pg.New(conf.Ctx, conf.PGConnStr)
	}
	if len(conf.FileName) != 0 {
		return filedb.NewFileDB(conf.FileName)
	}
	return inmemorydb.NewInMemoryDB()
}
