package storage

import (
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository/filedb"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository/inmemorydb"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository/pg"

	"github.com/sirupsen/logrus"
)


type DBConfig struct {
	FileName 	string
	PGConnStr	string
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
func New(conf DBConfig, logger *logrus.Logger) repository.Repository {
	if len(conf.PGConnStr) != 0 {
		// Создаем экземпляр подключения к БД и инициируем схему, если её нет
		db := pg.New(conf.PGConnStr)
		if err := db.SchemeInit(); err != nil {
			logger.Fatal(err)
		}
		logger.Println("DB Postgres is connecting")
		return db
	}
	if len(conf.FileName) != 0 {
		logger.Println("DB File is connecting")
		return filedb.NewFileDB(conf.FileName)
	}
	logger.Println("DB InMemory is connecting")
	return inmemorydb.NewInMemoryDB()
}
