package db

import (
	"context"

	filedb "github.com/yury-nazarov/shorturl/internal/app/repository/db/file"
	inmemorydb "github.com/yury-nazarov/shorturl/internal/app/repository/db/inmemory"
	"github.com/yury-nazarov/shorturl/internal/app/repository/db/pg"

	"github.com/yury-nazarov/shorturl/internal/app/repository/models"
	"github.com/yury-nazarov/shorturl/internal/config"

	"github.com/sirupsen/logrus"
)

// Repository - общее представление интерфейса для работы с БД
// 				имплементируем его для каждой реализации
type Repository interface {
	// TODO: В Add и Get можно передавать объект:

	Add(ctx context.Context, shortURL string, longURL string, token string) error
	Get(ctx context.Context, shortURL string, token string) (string, error)
	GetToken(ctx context.Context, token string) (bool, error)
	GetUserURL(ctx context.Context, token string) ([]models.Record, error)
	GetShortURLByIdentityPath(ctx context.Context, identityPath string, token string) int
	URLBulkDelete(ctx context.Context, urlsID chan int) error
	Ping() bool
	OriginURLExists(ctx context.Context, originURL string) (bool, error)
}

// TODO: Это же фабрика!

// New - возвращает подключение к БД, приоритеты:
//		 1. Postgres
//		 2. FileDB
//		 3. Inmemory
func New(cfg config.Config, logger *logrus.Logger) (Repository, error) {
	if len(cfg.DatabaseDSN) != 0 {
		// Создаем экземпляр подключения к БД и инициируем схему, если её нет
		db := pg.New(cfg.DatabaseDSN)
		if err := db.SchemeInit(); err != nil {
			return nil, err
		}
		logger.Println("DB Postgres is connecting")
		return db, nil
	}
	if len(cfg.FileStoragePath) != 0 {
		logger.Println("DB File is connecting")
		return filedb.NewFileDB(cfg.FileStoragePath), nil
	}
	logger.Println("DB InMemory is connecting")
	return inmemorydb.NewInMemoryDB(), nil
}
