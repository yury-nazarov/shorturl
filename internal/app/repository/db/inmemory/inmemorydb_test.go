package inmemorydb

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// BenchmarkInMemoryDB бенчмарки на основные методы работы с БД.
func BenchmarkInMemoryDB(b *testing.B) {
	db := NewInMemoryDB()
	token := "123456789qwertyXYZ"
	ctx := context.Background()
	b.Run("inmemory_Add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.Add(ctx, fmt.Sprintf("http://127.0.0.1:8080/KJYUS_%d", i), fmt.Sprintf("%s_%d", "https://www.youtube.com/watch?v=09nmlZjxRFs", i), fmt.Sprintf("%s_%d", token, i))
		}
	})
	b.Run("inmemory_Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.Get(ctx, fmt.Sprintf("http://127.0.0.1:8080/KJYUS_%d", i), token)
		}
	})
	b.Run("inmemory_GetToken", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.GetToken(ctx, fmt.Sprintf("%s_%d", token, i))
		}
	})
	b.Run("inmemory_GetUserURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.GetUserURL(ctx, fmt.Sprintf("%s_%d", token, i))
		}
	})
	b.Run("inmemory_OriginURLExists", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.OriginURLExists(ctx, fmt.Sprintf("http://127.0.0.1:8080/KJYUS_%d", i))
		}
	})
}

// Тесты

func TestInMemoryDB_Add(t *testing.T) {
	db := NewInMemoryDB()
	ctx := context.Background()
	err := db.Add(ctx, "shortURL1", "shortURL1", "token1")
	if err != nil {
		t.Errorf("Add return err %v;", err)
	}
}

// TestInMemoryDB_Get - по сокращенному URL получаем оригинальный
func TestInMemoryDB_Get(t *testing.T) {
	// Создаем объект
	db := NewInMemoryDB()
	ctx := context.Background()

	// Добавляем тестовые данные
	originURL := "longURL1"
	err := db.Add(ctx, "shortURL1", originURL, "token1")
	if err != nil {
		t.Errorf("Add return err %v;", err)
	}
	// Получаем оригинальный URL по сокрашенному
	data, err := db.Get(ctx, "shortURL1", "token1")
	if err != nil {
		t.Errorf("Add return err %v;", err)
	}
	assert.Equal(t, originURL, data)

	// Не верный URL отсутствует
	wrongShortURL := "wrongShortURL"
	_, err = db.Get(ctx, wrongShortURL, "token1")
	assert.Equal(t, fmt.Errorf("shorturl %s not found", wrongShortURL), err)
}

// TestInMemoryDB_GetToken Проверяем наличие токена в БД
func TestInMemoryDB_GetToken(t *testing.T) {
	// Создаем объект
	db := NewInMemoryDB()
	ctx := context.Background()

	// Добавляем тестовые данные
	tokens := []string{"token1", "", "-1"}
	for i, token := range tokens {
		err := db.Add(ctx, fmt.Sprintf("%s_%d", "shortURL", i), fmt.Sprintf("%s_%d", "originURL", i), token)
		if err != nil {
			t.Errorf("Add return err %v;", err)
		}
	}
	// Токены сущетсвуют
	for _, token := range tokens {
		// Получаем токен из БД
		tokenIsValid, err := db.GetToken(ctx, token)
		if err != nil {
			t.Errorf("Add return err %v;", err)
		}
		assert.Equal(t, true, tokenIsValid)
	}

	// Токена не существует
	tokenNotValid, err := db.GetToken(ctx, "wrongToken")
	if err != nil {
		t.Errorf("Add return err %v;", err)
	}
	assert.Equal(t, false, tokenNotValid)
}

// TestInMemoryDB_GetUserURL проверяем получение всех URL пользователя
func TestInMemoryDB_GetUserURL(t *testing.T) {
	// Создаем объект
	db := NewInMemoryDB()
	ctx := context.Background()

	// Добавляем тестовые данные
	for i := 0; i < 3; i++ {
		err := db.Add(ctx, fmt.Sprintf("%s_%d", "shortURL_", i), fmt.Sprintf("%s_%d", "originURL_", i),"token")
		if err != nil {
			t.Errorf("Add demo data return err %v;", err)
		}
	}

	// Получаем все URL пользователя по токену
	data, err := db.GetUserURL(ctx,"token")
	if err != nil {
		t.Errorf("GetUserURL return err %v;", err)
	}
	// Проверяем что сокращенный и оригинальный URL совпадают с теми, которые были добавлены в БД
	for i, record := range data{
		assert.Equal(t, fmt.Sprintf("%s_%d", "originURL_", i), record.OriginURL)
		assert.Equal(t, fmt.Sprintf("%s_%d", "shortURL_", i), record.ShortURL)
	}
}

// TestInMemoryDB_Ping Проверяем наличие соединения с БД
func TestInMemoryDB_Ping(t *testing.T) {
	// Создаем объект
	db := NewInMemoryDB()
	pingOk := db.Ping()
	assert.Equal(t, true, pingOk)
}

// Ниже тесты заглушки для функций реализующих обратную совместимость интерфейса с Postgres
// В оригинальном задании первой части курса для InMemory, fileDB их реализовывать не нужно,
// но код этот есть и учавствует в общем зачете покрытия тестами.

// TestInMemoryDB_OriginURLExists -
func TestInMemoryDB_OriginURLExists(t *testing.T) {
	// Создаем объект
	db := NewInMemoryDB()
	ctx := context.Background()

	ok, _ := db.OriginURLExists(ctx, "")
	assert.NotEqual(t, true, ok)
}


// TestInMemoryDB_GetShortURLByIdentityPath -
func TestInMemoryDB_GetShortURLByIdentityPath(t *testing.T) {
	// Создаем объект
	db := NewInMemoryDB()
	ctx := context.Background()

	num := db.GetShortURLByIdentityPath(ctx, "", "")
	assert.Equal(t, 0, num)
}

// TestInMemoryDB_URLBulkDelete -
func TestInMemoryDB_URLBulkDelete(t *testing.T) {
	// Создаем объект
	db := NewInMemoryDB()
	ctx := context.Background()
	ch := make(chan int)
	err := db.URLBulkDelete(ctx, ch)
	assert.Equal(t, nil, err)
}