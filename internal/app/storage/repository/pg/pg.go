package pg

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/yury-nazarov/shorturl/internal/app/storage/repository"
)

type pg struct {
	db *pgxpool.Pool
	ctx context.Context
}

// New - врнет ссылку на пулл соединений с PG
func New(ctx context.Context, connStr string) *pg {
	poolConfig, _ := pgxpool.ParseConfig(connStr)
	poolConfig.MinConns = 10
	poolConfig.MaxConns = 10

	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		log.Print(err)
	}
	dbPoolConnect := &pg {
		db: pool,
		ctx: ctx,
	}
	return dbPoolConnect
}


func (p *pg) Add(shortPath string, longURL string, token string) error {
	return nil
}

func (p *pg) Get(shortPath string) (string, error) {
	return "", nil
}

func (p *pg) GetToken(token string) (bool, error) {
	return true, nil
}

func (p *pg) GetUserURL(token string) ([]repository.RecordURL, error) {
	return []repository.RecordURL{}, nil
}

func (p *pg) Ping() bool {
	if err := p.db.Ping(p.ctx); err != nil {
		return false
	}
	return true
}