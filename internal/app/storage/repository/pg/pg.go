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
// 			  TODO: Возвращать ошибку и завершать работу приложения
func (p *pg) SchemeInit() error {
	// Общая таблица содержащая ссылки на остальные  таблицы.
	_, err := p.db.Exec(p.ctx, `CREATE TABLE IF NOT EXISTS shorten_url (
                          id serial PRIMARY KEY,
						  url INT NOT NULL,
						  owner INT NOT NULL)`)
	if err != nil {
		return fmt.Errorf("create table `shorten_url`: %w", err)
	}


	// Таблица для URL
	_, err = p.db.Exec(p.ctx, `CREATE TABLE IF NOT EXISTS url (
                          id serial PRIMARY KEY,
						  origin VARCHAR (255) NOT NULL,
						  short VARCHAR (255) NOT NULL)`)
	if err != nil {
		return fmt.Errorf("create table `url`: %w", err)
	}

	// Создаем уникальный индекс для таблицы URL
	_, err = p.db.Exec(p.ctx, `CREATE UNIQUE INDEX IF NOT EXISTS url_index ON url (origin)`)
	if err != nil {
		return fmt.Errorf("create index `url_index`: %w", err)
	}

	// Таблица для пользовательских данных
	_, err = p.db.Exec(p.ctx, `CREATE TABLE IF NOT EXISTS owner (
                          id serial PRIMARY KEY,
						  token  VARCHAR (255) NOT NULL)`)
	if err != nil {
		return fmt.Errorf("create table `owner`: %w", err)
	}
	return nil
}

// Owner Представление таблицы shorten_url.owner
type Owner struct {
	id int
	token string
}

// URL - представление объекта URL
type URL struct {
	id int
	shortURL string // TODO: shortURL -> short
	origin string
}

// Add - добавляет новую запись в таблицу: shorten_url
func (p *pg) Add(shortURL string, longURL string, token string) error {
	// Проверяем наличие пользователя в БД с определенным token, получаем id для дальнейшего insert
	owner := p.ownerTokenExist(token)

	// Если пользователя нет, добавляем токен в БД и получаем его id для дальнейшего insert
	if owner.id == 0 {
		if err := p.db.QueryRow(p.ctx, `INSERT INTO owner (token) VALUES ($1) RETURNING id;`, token).Scan(&owner.id); err != nil {
			return fmt.Errorf("sql insert into token err: %w", err)
		}
		log.Printf("token '%s' was added into DB with 'id'=%d", token, owner.id)
	}

	// Добавляем URL
	url := URL{}
	//  Выполнит INSERT и вернет id добавленой записи, либо обновит существующую запись и вернет ее id
	err := p.db.QueryRow(p.ctx, `INSERT INTO url (origin, short) 
									 VALUES ($1, $2) 
									 ON CONFLICT (origin) DO UPDATE SET origin=$1 
									 RETURNING id;`, longURL, shortURL).Scan(&url.id)
	if err != nil {
		return fmt.Errorf("sql insert into url: %w", err)
	}

	// Добавляем owner.id, url.id в общую таблицу
	_, err = p.db.Exec(p.ctx, `INSERT INTO shorten_url (url, owner) VALUES ($1, $2);`, url.id, owner.id)
	if err != nil {
		return fmt.Errorf("sql insert into table `shorten_url`: %w", err)
	}

	return nil
}

// Get - Возвращает оригинальный URL
func (p *pg) Get(shortURL string) (string, error) {
	url := URL{}
	if err := p.db.QueryRow(p.ctx, `SELECT origin FROM url WHERE short=$1 LIMIT 1`, shortURL).Scan(&url.origin); err != nil {
		return "", fmt.Errorf("url not found: %w", err)
	}
	return url.origin, nil
}

// GetToken - Проверяет наличие токена в БД
func (p *pg) GetToken(token string) (bool, error) {
	owner := Owner{}
	if err := p.db.QueryRow(p.ctx, `SELECT id FROM owner WHERE token=$1 LIMIT 1;`, token).Scan(&owner.id); err != nil {
		return false, fmt.Errorf("token not found: %w", err)
	}
	return true, nil
}

// GetUserURL - Возвращает все url для конкретного token
func (p *pg) GetUserURL(token string) ([]repository.RecordURL, error) {
	// Слайс который будем возвращать как результат работы метода
	var urls []repository.RecordURL

	// Проверяем наличие пользователя в БД с определенным token, получаем id для дальнейшего select
	owner := p.ownerTokenExist(token)
	if owner.id == 0 {
		return urls, fmt.Errorf("owner with token: %s not exist", token)
	}

	// Получаем все id для url для конкретного owner
	rows, err := p.db.Query(p.ctx, `SELECT url FROM shorten_url WHERE owner=$1`, owner.id)
	if err != nil {
		return urls, err
	}
	defer rows.Close()

	// Достаем по id конкретные URL: origin, short.
	for rows.Next() {
		// Получаем id url
		url := URL{}
		rows.Scan(&url.id)

		// Получаем из БД пару URL: origin, short и добавляем  результирующий слайс
		ownerURL, err := p.getURLById(url.id)
		if err != nil {
			return urls, err
		}
		urls = append(urls, ownerURL)
	}
	return urls, nil
}

// Ping - Проверка соединения с БД
func (p *pg) Ping() bool {
	if err := p.db.Ping(p.ctx); err != nil {
		return false
	}
	return true
}

// getURLById - по ID получаем пару URL: origin, short
func (p *pg) getURLById(id int) (repository.RecordURL, error) {
	url := repository.RecordURL{}
	if err := p.db.QueryRow(p.ctx, `SELECT origin, short FROM url WHERE id=$1 LIMIT 1`, id).Scan(&url.OriginURL, &url.ShortURL); err != nil {
		return url, err
	}
	return url, nil
}

// Получает информацию о пользователе из БД по токену
// Если пользователя не существует, вернет структуру Owner с дефолтными значениями полей
func (p *pg) ownerTokenExist(token string) Owner {
	owner := Owner{}
	// Проверяем наличие пользователя в БД с определенным token
	if err := p.db.QueryRow(p.ctx, `SELECT id FROM owner WHERE token=$1 LIMIT 1;`, token).Scan(&owner.id); err != nil {
		log.Printf("sql select token err: %s", err)
	}
	return owner
}

// OriginUrlExists - проверяет наличие URL в БД
func (p *pg) OriginUrlExists(originURL string) (bool, error) {
	url := URL{}
	err := p.db.QueryRow(p.ctx, `SELECT origin FROM url WHERE origin=$1 LIMIT 1`, originURL).Scan(&url.origin)
	if err != nil {
		return false, err
	}
	if len(url.origin) == 0 {
		return false, nil
	}
	return true, nil

}