package pg

import (
	"context"
	"fmt"
	"log"

	"database/sql"
	_ "github.com/jackc/pgx/v4/stdlib"
	//_ "github.com/jackc/pgx/v4"

	"github.com/yury-nazarov/shorturl/internal/app/storage/repository"
)

type pg struct {
	db *sql.DB
	//db *pgxpool.Pool
	ctx context.Context
}

// New - врнет ссылку на пулл соединений с PG
func New(ctx context.Context, connStr string) *pg {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: defer db.Close() ???
	dbConnect := &pg{
		db: db,
		ctx: ctx,
	}
	return dbConnect
}
//func New(ctx context.Context, connStr string) *pg {
//	poolConfig, _ := pgxpool.ParseConfig(connStr)
//	poolConfig.MinConns = 5
//	poolConfig.MaxConns = 5
//
//	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
//	if err != nil {
//		log.Print(err)
//	}
//	dbPoolConnect := &pg {
//		db: pool,
//		ctx: ctx,
//	}
//	return dbPoolConnect
//}

// SchemeInit Создает таблицы в БД если они не созданы.
func (p *pg) SchemeInit() error {
	// Общая таблица содержащая ссылки на остальные  таблицы.
	//_, err := p.db.Exec(p.ctx, `CREATE TABLE IF NOT EXISTS shorten_url (
    //                      id serial PRIMARY KEY,
	//					  url INT NOT NULL,
	//					  owner INT NOT NULL,
	//					  delete BOOLEAN DEFAULT FALSE)`)
	_, err := p.db.ExecContext(p.ctx, `CREATE TABLE IF NOT EXISTS shorten_url (
                          id serial PRIMARY KEY,
						  url INT NOT NULL,
						  owner INT NOT NULL,
						  delete BOOLEAN DEFAULT FALSE)`)
	if err != nil {
		return fmt.Errorf("create table `shorten_url`: %w", err)
	}


	// Таблица для URL
	//_, err = p.db.Exec(p.ctx, `CREATE TABLE IF NOT EXISTS url (
    //                      id serial PRIMARY KEY,
	//					  origin VARCHAR (255) NOT NULL,
	//					  short VARCHAR (255) NOT NULL)`)
	_, err = p.db.ExecContext(p.ctx, `CREATE TABLE IF NOT EXISTS url (
                          id serial PRIMARY KEY,
						  origin VARCHAR (255) NOT NULL,
						  short VARCHAR (255) NOT NULL)`)
	if err != nil {
		return fmt.Errorf("create table `url`: %w", err)
	}

	// Создаем уникальный индекс для таблицы URL
	//_, err = p.db.Exec(p.ctx, `CREATE UNIQUE INDEX IF NOT EXISTS url_index ON url (origin)`)
	_, err = p.db.ExecContext(p.ctx, `CREATE UNIQUE INDEX IF NOT EXISTS url_index ON url (origin)`)
	if err != nil {
		return fmt.Errorf("create index `url_index`: %w", err)
	}

	// Таблица для пользовательских данных
	//_, err = p.db.Exec(p.ctx, `CREATE TABLE IF NOT EXISTS owner (
    //                      id serial PRIMARY KEY,
	//					  token  VARCHAR (255) NOT NULL)`)
	_, err = p.db.ExecContext(p.ctx, `CREATE TABLE IF NOT EXISTS owner (
                          id serial PRIMARY KEY,
						  token  VARCHAR (255) NOT NULL)`)
	if err != nil {
		return fmt.Errorf("create table `owner`: %w", err)
	}
	return nil
}

// URL - представление объекта URL
type URL struct {
	id int
	shortURL string // TODO: shortURL -> short
	origin string
	delete bool // default false
}

// Add - добавляет новую запись в таблицу: shorten_url
func (p *pg) Add(shortURL string, longURL string, token string) error {
	// Проверяем наличие пользователя в БД с определенным token, получаем id для дальнейшего insert
	owner := p.GetOwnerToken(token)
	// Если пользователя нет, добавляем токен в БД и получаем его id для дальнейшего insert
	if owner.ID == 0 {
		//if err := p.db.QueryRow(p.ctx, `INSERT INTO owner (token) VALUES ($1) RETURNING id;`, token).Scan(&owner.ID); err != nil {
		//	return fmt.Errorf("sql insert into token err: %w", err)
		//}
		if err := p.db.QueryRowContext(p.ctx, `INSERT INTO owner (token) VALUES ($1) RETURNING id;`, token).Scan(&owner.ID); err != nil {
			return fmt.Errorf("sql insert into token err: %w", err)
		}
		log.Printf("token '%s' was added into DB with 'id'=%d", token, owner.ID)
	}

	// Добавляем URL если нет
	url := URL{}
	//  Выполнит INSERT и вернет id добавленой записи, либо обновит существующую запись и вернет ее id
	//err := p.db.QueryRow(p.ctx, `INSERT INTO url (origin, short)
	//								 VALUES ($1, $2)
	//								 ON CONFLICT (origin) DO UPDATE SET origin=$1
	//								 RETURNING id;`, longURL, shortURL).Scan(&url.id)
	err := p.db.QueryRowContext(p.ctx, `INSERT INTO url (origin, short) 
									 VALUES ($1, $2) 
									 ON CONFLICT (origin) DO UPDATE SET origin=$1 
									 RETURNING id;`, longURL, shortURL).Scan(&url.id)
	if err != nil {
		return fmt.Errorf("sql insert into url: %w", err)
	}

	// Проверяем наличие записи, что бы не было дублей
	var existID int
	//err = p.db.QueryRow(p.ctx, `SELECT id FROM shorten_url WHERE url=$1 AND owner=$2`, url.id, owner.ID).Scan(&existID)
	err = p.db.QueryRowContext(p.ctx, `SELECT id FROM shorten_url WHERE url=$1 AND owner=$2`, url.id, owner.ID).Scan(&existID)
	if err != nil {
		log.Printf("sql check record err: %s", err)
	}
	log.Println("existID:", existID)

	// выполняем только в том случае, если записи в БД нет
	if existID == 0 {
		// Добавляем owner.id, url.id в общую таблицу
		//_, err = p.db.Exec(p.ctx, `INSERT INTO shorten_url (url, owner) VALUES ($1, $2);`, url.id, owner.ID)
		_, err = p.db.ExecContext(p.ctx, `INSERT INTO shorten_url (url, owner) VALUES ($1, $2);`, url.id, owner.ID)
		if err != nil {
			return fmt.Errorf("sql insert into table `shorten_url`: %w", err)
		}
	}

	return nil
}

// Get - Возвращает оригинальный URL
func (p *pg) Get(shortURL string, token string) (string, error) {
	var urlID int
	var originURL string
	//var isDelete bool

	// Получаем оргинальный URL
	//err := p.db.QueryRow(p.ctx, `SELECT id, origin FROM url WHERE short=$1 LIMIT 1`, shortURL).Scan(&urlID, &originURL)
	err := p.db.QueryRowContext(p.ctx, `SELECT id, origin FROM url WHERE short=$1 LIMIT 1`, shortURL).Scan(&urlID, &originURL)
	if err != nil {
		return "",  fmt.Errorf("sql url not found: %w", err)
	}


	// Получаем статус URL для конкретного пользователя (удален/не удален)
	//isDelete, err := p.db.Exec(p.ctx, `SELECT delete FROM shorten_url
	//									WHERE url=$1
	//									AND owner=(SELECT id FROM owner WHERE token=$2)
	//									LIMIT 1`, urlID, token)
	// TODO: Есть более изящный способ, вчера видел!!!
	//isDelete, err := p.db.ExecContext(p.ctx, `SELECT delete FROM shorten_url
	//									WHERE url=$1
	//									AND owner=(SELECT id FROM owner WHERE token=$2)
	//									LIMIT 1`, urlID, token)
	//if err != nil {
	//	return "",  fmt.Errorf("sql SELECT delete FORM shorten_url: %w", err)
	//}

	//// Возвращаем пустую строку если URL помечен как удаленный
	//if isDelete.String() == "true" {
	//	return "", nil
	//}
	// Возвращаем оригинальный URL если помечен как не удаленный
	return originURL, nil
}

// GetToken - Проверяет наличие токена в БД
func (p *pg) GetToken(token string) (bool, error) {
	owner := repository.Owner{}
	//if err := p.db.QueryRow(p.ctx, `SELECT id FROM owner WHERE token=$1 LIMIT 1;`, token).Scan(&owner.ID); err != nil {
	//	return false, fmt.Errorf("token not found: %w", err)
	//}
	if err := p.db.QueryRowContext(p.ctx, `SELECT id FROM owner WHERE token=$1 LIMIT 1;`, token).Scan(&owner.ID); err != nil {
		return false, fmt.Errorf("token not found: %w", err)
	}
	return true, nil
}


// GetUserURL - Возвращает все url для конкретного token
func (p *pg) GetUserURL(token string) ([]repository.RecordURL, error) {
	// Слайс который будем возвращать как результат работы метода
	var urls []repository.RecordURL

	// Проверяем наличие пользователя в БД с определенным token, получаем id для дальнейшего select
	owner := p.GetOwnerToken(token)
	if owner.ID == 0 {
		return urls, fmt.Errorf("owner with token: %s not exist", token)
	}

	// Получаем все id для url для конкретного owner
	//rows, err := p.db.Query(p.ctx, `SELECT url FROM shorten_url WHERE owner=$1`, owner.ID)
	rows, err := p.db.QueryContext(p.ctx, `SELECT url FROM shorten_url WHERE owner=$1`, owner.ID)
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
		ownerURL, err := p.getURLByID(url.id)
		if err != nil {
			return urls, err
		}
		urls = append(urls, ownerURL)
	}
	return urls, nil
}

// Ping - Проверка соединения с БД
func (p *pg) Ping() bool {
	//if err := p.db.Ping(p.ctx); err != nil {
	//	return false
	//}
	if err := p.db.Ping(); err != nil {
		return false
	}
	return true
}

// getURLByID - по ID получаем пару URL: origin, short
func (p *pg) getURLByID(id int) (repository.RecordURL, error) {
	url := repository.RecordURL{}
	//if err := p.db.QueryRow(p.ctx, `SELECT origin, short FROM url WHERE id=$1 LIMIT 1`, id).Scan(&url.OriginURL, &url.ShortURL); err != nil {
	//	return url, err
	//}
	if err := p.db.QueryRowContext(p.ctx, `SELECT origin, short FROM url WHERE id=$1 LIMIT 1`, id).Scan(&url.OriginURL, &url.ShortURL); err != nil {
		return url, err
	}
	return url, nil
}

// GetOwnerToken Получает информацию о пользователе из БД по токену
// Если пользователя не существует, вернет структуру Owner с дефолтными значениями полей
func (p *pg) GetOwnerToken(token string) repository.Owner {
	owner := repository.Owner{}
	// Проверяем наличие пользователя в БД с определенным token
	//if err := p.db.QueryRow(p.ctx, `SELECT id FROM owner WHERE token=$1 LIMIT 1;`, token).Scan(&owner.ID); err != nil {
	//	log.Printf("sql select token err: %s", err)
	//}
	if err := p.db.QueryRowContext(p.ctx, `SELECT id FROM owner WHERE token=$1 LIMIT 1;`, token).Scan(&owner.ID); err != nil {
		log.Printf("sql select token err: %s", err)
	}
	return owner
}

// GetShortURLByIdentityPath вернет все записи пользователя по идентификатору короткого URL
func (p *pg) GetShortURLByIdentityPath(identityPath string, token string) int {
	var urlID int
	//err := p.db.QueryRow(p.ctx, `SELECT id FROM shorten_url
	//										WHERE url=(SELECT id FROM url WHERE short LIKE $1)
	//										AND owner=(SELECT id FROM owner WHERE token=$2);`,
	//										"%"+identityPath, token).Scan(&urlID)
	err := p.db.QueryRowContext(p.ctx, `SELECT id FROM shorten_url 
											WHERE url=(SELECT id FROM url WHERE short LIKE $1) 
											AND owner=(SELECT id FROM owner WHERE token=$2);`,
											"%"+identityPath, token).Scan(&urlID)
	if err != nil {
		log.Printf("sql select short url by identity path was err: %s", err)
	}
	return urlID
}

// URLMarkDeleted помечает удаленным в таблице shorten_url. delete=true
func (p *pg) URLMarkDeleted(id int) {
	//_, err := p.db.Exec(p.ctx, `UPDATE shorten_url SET delete=true WHERE id=$1`, id)
	_, err := p.db.ExecContext(p.ctx, `UPDATE shorten_url SET delete=true WHERE id=$1`, id)
	if err != nil {
		log.Println("sql mark delete err", err)
	}

}

// OriginURLExists - проверяет наличие URL в БД
func (p *pg) OriginURLExists(originURL string) (bool, error) {
	url := URL{}
	//err := p.db.QueryRow(p.ctx, `SELECT origin FROM url WHERE origin=$1 LIMIT 1`, originURL).Scan(&url.origin)
	err := p.db.QueryRowContext(p.ctx, `SELECT origin FROM url WHERE origin=$1 LIMIT 1`, originURL).Scan(&url.origin)
	if err != nil {
		return false, err
	}
	if len(url.origin) == 0 {
		return false, nil
	}
	return true, nil
}
