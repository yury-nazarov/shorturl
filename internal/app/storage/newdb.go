package storage

import (
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository/filedb"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository/inmemorydb"
	"github.com/yury-nazarov/shorturl/internal/app/storage/repository/pg"
	"github.com/yury-nazarov/shorturl/internal/config"

	"github.com/sirupsen/logrus"
)




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
func New(cfg config.Config, logger *logrus.Logger) repository.Repository {
	if len(cfg.DatabaseDSN) != 0 {
		// Создаем экземпляр подключения к БД и инициируем схему, если её нет
		db := pg.New(cfg.DatabaseDSN)
		if err := db.SchemeInit(); err != nil {
			logger.Fatal(err)
		}
		logger.Println("DB Postgres is connecting")
		return db
	}
	if len(cfg.FileStoragePath) != 0 {
		logger.Println("DB File is connecting")
		return filedb.NewFileDB(cfg.FileStoragePath)
	}
	logger.Println("DB InMemory is connecting")
	return inmemorydb.NewInMemoryDB()
}
