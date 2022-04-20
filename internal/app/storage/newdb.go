package storage

import (
	"github.com/yury-nazarov/shorturl/internal/app/storage/filedb"
	"github.com/yury-nazarov/shorturl/internal/app/storage/inmemorydb"
)


type DBConfig struct {
	FileName string
}

func New(conf DBConfig) Repository{
	if len(conf.FileName) != 0 {
		return filedb.NewFileDB(conf.FileName)
	}
	return inmemorydb.NewInMemoryDB()
}
