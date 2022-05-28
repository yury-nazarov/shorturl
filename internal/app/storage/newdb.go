package storage

import (
	"context"
	"log"


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
		// Создаем экземпляр подключения к БД и инициируем схему, если её нет
		db := pg.New(conf.Ctx, conf.PGConnStr)
		if err := db.SchemeInit(); err != nil {
			log.Fatal(err)
		}
		log.Println("DB Postgres is connecting")
		return db
	}
	if len(conf.FileName) != 0 {
		log.Println("DB File is connecting")
		return filedb.NewFileDB(conf.FileName)
	}
	log.Println("DB InMemory is connecting")
	return inmemorydb.NewInMemoryDB()
}
