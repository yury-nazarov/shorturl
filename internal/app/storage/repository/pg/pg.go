package pg

import (
	"context"
	"fmt"
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

// SchemeInit Создает таблицы в БД если они не созданы.
func (p *pg) SchemeInit(){
	_, err := p.db.Exec(p.ctx, `CREATE TABLE IF NOT EXISTS shorten_url (
                          id serial PRIMARY KEY,
						  short_url  VARCHAR (255) NOT NULL,
						  long_url VARCHAR (255) NOT NULL,
						  token_id INT NOT NULL)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = p.db.Exec(p.ctx, `CREATE TABLE IF NOT EXISTS users (
                          id serial PRIMARY KEY,
						  token  VARCHAR (255) NOT NULL)`)
	if err != nil {
		log.Fatal(err)
	}
}


type User struct {
	id int
	token string
}

type ShortenURL struct {
	id int
	shortURL string
	longURL string
}

// Add - добавляет новую запись в тадлицу: shorten_url
func (p *pg) Add(shortURL string, longURL string, token string) error {
	user := User{}
	// Проверяем наличие пользователя в БД с определенным token, получаем id для дальнейшего insert
	if err := p.db.QueryRow(p.ctx, `SELECT id FROM users WHERE token=$1 LIMIT 1;`, token).Scan(&user.id); err != nil {
		log.Printf("sql select token err: %s", err)
	}

	// Если пользователя нет, добавляем токен в БД и получаем его id для дальнейшего insert
	if user.id == 0 {
		if err := p.db.QueryRow(p.ctx, `INSERT INTO users (token) VALUES ($1) RETURNING id;`, token).Scan(&user.id); err != nil {
			return fmt.Errorf("sql insert into token err: %s", err)
		}
		log.Printf("token '%s' was added into DB with 'id'=%d", token, user.id)
	}

	// Добавляем в БД shortURL, longURL, token
	newURL := ShortenURL{}
	if err := p.db.QueryRow(p.ctx, `INSERT INTO shorten_url (short_url, long_url, token_id) VALUES ($1, $2, $3) RETURNING id;`, shortURL, longURL, user.id).Scan(&newURL.id); err != nil {
		return fmt.Errorf("sql insert into shorten_url err: %s", err)
	}
	log.Printf("shorten_urecord short_url:'%s', long_url:'%s', token_id: '%s' was added into shorten_url with 'id'=%d", shortURL, longURL, token, newURL.id)

	return nil
}

// Get - Возвращает оригинальный long_url из таблицы shorten_url
func (p *pg) Get(shortURL string) (string, error) {
	url := ShortenURL{}
	if err := p.db.QueryRow(p.ctx, `SELECT long_url FROM shorten_url WHERE short_url=$1 LIMIT 1`, shortURL).Scan(&url.longURL); err != nil {
		return "", fmt.Errorf("url not found: %s", err)
	}
	return url.longURL, nil
}

// GetToken - Проверяет наличие токена в БД
func (p *pg) GetToken(token string) (bool, error) {
	user := User{}
	if err := p.db.QueryRow(p.ctx, `SELECT id FROM users WHERE token=$1 LIMIT 1;`, token).Scan(&user.id); err != nil {
		return false, fmt.Errorf("token not found: %s", err)
	}
	return true, nil
}

// GetUserURL - Возвращает все url из таблицы shorten_url, для конкретного token
func (p *pg) GetUserURL(token string) ([]repository.RecordURL, error) {
	var urls []repository.RecordURL

	// Проверяем наличие пользователя в БД с определенным token, получаем id для дальнейшего select
	user := User{}
	if err := p.db.QueryRow(p.ctx, `SELECT id FROM users WHERE token=$1 LIMIT 1;`, token).Scan(&user.id); err != nil {
		log.Printf("sql select token err: %s", err)
	}

	if user.id == 0 {
		return urls, fmt.Errorf("user with token: %s not exist", token)
	}

	rows, err := p.db.Query(p.ctx, `SELECT short_url, long_url FROM shorten_url WHERE token_id=$1`, user.id)
	if err != nil {
		return urls, err
	}
	defer rows.Close()

	for rows.Next() {
		url := repository.RecordURL{}
		rows.Scan(&url.ShortURL, &url.OriginURL)
		fmt.Print(url)
		urls = append(urls, url)
	}
	return urls, nil
}

func (p *pg) Ping() bool {
	if err := p.db.Ping(p.ctx); err != nil {
		return false
	}
	return true
}