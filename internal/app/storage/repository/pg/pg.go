package pg

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/yury-nazarov/shorturl/internal/app/storage/repository"
)

type pg struct {
	db *sql.DB

}

// New - врнет ссылку на соединение с PG
func New(connStr string) *pg {
	db, err := sql.Open("pgx", connStr)

	if err != nil {
		log.Fatal(err)
	}
	// TODO: defer db.Close() ???
	dbConnect := &pg{
		db: db,
	}
	return dbConnect
}

// SchemeInit Создает таблицы в БД если они не созданы.
func (p *pg) SchemeInit() error {
	// Контекст для инициализации БД
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Общая таблица таблицы.
	_, err := p.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS shorten_url (
                          id serial PRIMARY KEY,
						  origin VARCHAR (255) NOT NULL,
						  short VARCHAR (255) NOT NULL,
						  owner VARCHAR (255) NOT NULL,
						  delete BOOLEAN DEFAULT FALSE)`)
	if err != nil {
		return fmt.Errorf("create table `shorten_url`: %w", err)
	}
	return nil
}

//func (p *pg) SchemeInit() error {
//	// Контекст для инициализации БД
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	// Общая таблица содержащая ссылки на остальные  таблицы.
//	_, err := p.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS shorten_url (
//                          id serial PRIMARY KEY,
//						  url INT NOT NULL,
//						  owner INT NOT NULL,
//						  delete BOOLEAN DEFAULT FALSE)`)
//	if err != nil {
//		return fmt.Errorf("create table `shorten_url`: %w", err)
//	}
//
//
//	// Таблица для URL
//	_, err = p.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS url (
//                          id serial PRIMARY KEY,
//						  origin VARCHAR (255) NOT NULL,
//						  short VARCHAR (255) NOT NULL)`)
//	if err != nil {
//		return fmt.Errorf("create table `url`: %w", err)
//	}
//
//	// Создаем уникальный индекс для таблицы URL
//	_, err = p.db.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS url_index ON url (origin)`)
//	if err != nil {
//		return fmt.Errorf("create index `url_index`: %w", err)
//	}
//
//	// Таблица для пользовательских данных
//	_, err = p.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS owner (
//                          id serial PRIMARY KEY,
//						  token  VARCHAR (255) NOT NULL)`)
//	if err != nil {
//		return fmt.Errorf("create table `owner`: %w", err)
//	}
//	return nil
//}

// URL - представление объекта URL
type URL struct {
	id int
	shortURL string // TODO: shortURL -> short
	origin string
	delete bool // default false
}

// Ping - Проверка соединения с БД
func (p *pg) Ping() bool {
	if err := p.db.Ping(); err != nil {
		return false
	}
	return true
}

// Add - добавляет новую запись в таблицу: shorten_url
//		 записать в БД url и токен.
func (p *pg) Add(ctx context.Context, shortURL string, longURL string, token string) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO shorten_url (origin, short, owner) VALUES ($1, $2, $3)`, longURL, shortURL, token)
	if err != nil {
		log.Printf("sql | insert new url err %s\n", err)
	}

	log.Printf("DEBUG: User: %s add URL: %s -> %s\n", token,  longURL, shortURL)
	return nil
}

// Get - Возвращает оригинальный URL или 410 если он помечен удаленных (для всех пользователей)
func (p *pg) Get(ctx context.Context, shortURL string, token string) (string, error) {
	var originURL string
	var isDelete bool

	// Получаем оргинальный URL
	err := p.db.QueryRowContext(ctx, `SELECT origin, delete FROM shorten_url WHERE short=$1 LIMIT 1`, shortURL).Scan(&originURL, &isDelete)

	if err != nil {
		log.Printf("sql |  get origin url status err: %s", err)
	}

	// Возвращаем пустую строку если URL помечен как удаленный
	if isDelete {
		return "", nil
	}
	// Возвращаем оригинальный URL если помечен как не удаленный
	return originURL, nil
}

// GetUserURL - Возвращает все url для конкретного token
func (p *pg) GetUserURL(ctx context.Context, token string) ([]repository.RecordURL, error) {
	// Слайс который будем возвращать как результат работы метода
	var urls []repository.RecordURL

	// Получаем все url для конкретного owner
	rows, err := p.db.QueryContext(ctx, `SELECT origin, short FROM shorten_url WHERE owner=$1`, token)
	fmt.Println("token", token)
	if err != nil {
		return urls, err
	}
	defer rows.Close()

	log.Println("DEBUG 4:",)

	// Достаем по id конкретные URL: origin, short.
	for rows.Next() {
		log.Println("DEBUG 5:")
		var url repository.RecordURL
		rows.Scan(&url.OriginURL, &url.ShortURL)
		log.Println("-", url.ShortURL, url.OriginURL)
		urls = append(urls, url)
	}
	if err = rows.Err(); err != nil {
		log.Printf("sql | get users url err: %s\n", err)
	}
	log.Println("DEBUG 6:", urls)
	return urls, nil
}

// GetShortURLByIdentityPath вернет все записи пользователя по идентификатору короткого URL
func (p *pg) GetShortURLByIdentityPath(ctx context.Context, identityPath string, token string) int {
	var urlID int
	err := p.db.QueryRowContext(ctx, `SELECT id FROM shorten_url 
											WHERE short LIKE $1
											AND owner=$2`,
											"%"+identityPath, token).Scan(&urlID)

	if err != nil {
		log.Printf("sql | select short url by identity path err: %s", err)
	}
	return urlID
}



// URLBulkDelete помечает удаленным в таблице shorten_url. delete=true
func (p *pg) URLBulkDelete(ctx context.Context,  urlsID chan int) error {
	// шаг 1 — объявляем транзакцию
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("sql | transaction begin err: %w", err)
	}
	// если возникает ошибка, откатываем изменения
	defer tx.Rollback()

	// шаг 2 — готовим инструкцию
	stmt, err := tx.PrepareContext(ctx, "UPDATE shorten_url SET delete=true WHERE id=$1")
	if err != nil {
		return fmt.Errorf("sql | transaction prepare context err err %w", err)
	}
	defer stmt.Close()

	// TODO: FanIn использовать для максимального быстрого наполнения буфера объектов обновления
	// шаг 3 - указываем, что для каждого id в таблице shorten_url нужно обновить поле delete
	for id := range urlsID{
		fmt.Printf("DEBUG: transaction statement prepare delete url with ID:%d\n", id)
		if _, err = stmt.ExecContext(ctx, id); err != nil {
			return fmt.Errorf("sql | transaction statement exec context err %w", err)
		}
	}
	// шаг 4 — сохраняем изменения
	fmt.Printf("DEBUG: Commit transaction\n")
	return tx.Commit()
}

//func (p *pg) URLBulkDelete(ctx context.Context, idList []int) error {
//	// шаг 1 — объявляем транзакцию
//	tx, err := p.db.Begin()
//	if err != nil {
//		return fmt.Errorf("sql | transaction begin err: %w", err)
//	}
//	// если возникает ошибка, откатываем изменения
//	defer tx.Rollback()
//
//	// шаг 2 — готовим инструкцию
//	stmt, err := tx.PrepareContext(ctx, "UPDATE shorten_url SET delete=true WHERE id=$1")
//	if err != nil {
//		return fmt.Errorf("sql | transaction prepare context err err 2 %w", err)
//	}
//	defer stmt.Close()
//
//	for _, id := range idList {
//		// шаг 3 - указываем, что для каждого id в таблице shorten_url нужно обновить поле delete
//		if _, err = stmt.ExecContext(ctx, strconv.Itoa(id)); err != nil {
//			return fmt.Errorf("sql | transaction statement exec context err %w", err)
//		}
//	}
//	// шаг 4 — сохраняем изменения
//	return tx.Commit()
//}

////// Вспомогательные методы

// GetToken - Проверяет наличие токена в БД
func (p *pg) GetToken(ctx context.Context, token string) (bool, error) {
	owner := repository.Owner{}
	if err := p.db.QueryRowContext(ctx, `SELECT id FROM owner WHERE token=$1 LIMIT 1;`, token).Scan(&owner.ID); err != nil {
		return false, fmt.Errorf("sql | token not found: %w", err)
	}
	return true, nil
}


//// getURLByID - по ID получаем пару URL: origin, short
//func (p *pg) getURLByID(ctx context.Context, id int) (repository.RecordURL, error) {
//	url := repository.RecordURL{}
//
//	if err := p.db.QueryRowContext(ctx, `SELECT origin, short FROM url WHERE id=$1 LIMIT 1`, id).Scan(&url.OriginURL, &url.ShortURL); err != nil {
//		return url, err
//	}
//	return url, nil
//}

//// GetOwnerToken Получает информацию о пользователе из БД по токену
//// Если пользователя не существует, вернет структуру Owner с дефолтными значениями полей
//func (p *pg) GetOwnerToken(ctx context.Context, token string) repository.Owner {
//	owner := repository.Owner{}
//	// Проверяем наличие пользователя в БД с определенным token
//	if err := p.db.QueryRowContext(ctx, `SELECT id FROM owner WHERE token=$1 LIMIT 1;`, token).Scan(&owner.ID); err != nil {
//		log.Printf("sql | select token err: %s", err)
//	}
//	return owner
//}


// OriginURLExists - проверяет наличие URL в БД
func (p *pg) OriginURLExists(ctx context.Context, originURL string) (bool, error) {
	url := URL{}
	err := p.db.QueryRowContext(ctx, `SELECT origin FROM shorten_url WHERE origin=$1 LIMIT 1`, originURL).Scan(&url.origin)
	if err != nil {
		return false, err
	}
	if len(url.origin) == 0 {
		return false, nil
	}
	return true, nil
}