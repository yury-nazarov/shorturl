package filedb

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// BenchmarkFileDB бенчмарки на основные методы работы с БД.
func BenchmarkFileDB(b *testing.B) {
	db := NewFileDB("bm.txt")
	token := "123456789qwertyXYZ"
	ctx := context.Background()
	b.Run("file_Add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.Add(ctx, fmt.Sprintf("http://127.0.0.1:8080/KJYUS_%d", i), fmt.Sprintf("%s_%d", "https://www.youtube.com/watch?v=09nmlZjxRFs", i), fmt.Sprintf("%s_%d", token, i))
		}
	})
	b.Run("file_Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.Get(ctx, fmt.Sprintf("http://127.0.0.1:8080/KJYUS_%d", i), token)
		}
	})
	b.Run("file_GetToken", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.GetToken(ctx, fmt.Sprintf("%s_%d", token, i))
		}
	})
	b.Run("file_GetUserURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.GetUserURL(ctx, fmt.Sprintf("%s_%d", token, i))
		}
	})
	b.Run("file_OriginURLExists", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.OriginURLExists(ctx, fmt.Sprintf("http://127.0.0.1:8080/KJYUS_%d", i))
		}
	})
}

// Тесты
func TestFileDB_Add(t *testing.T) {
	db := NewFileDB("tests.txt")
	ctx := context.Background()
	err := db.Add(ctx, "shortURL1", "shortURL1", "token1")
	if err != nil {
		t.Errorf("Add return err %v;", err)
	}

	err = os.Remove("tests.txt")
	if err != nil {
		assert.Errorf(t, err, "can't remove file")
	}
}

// TestFileDB_Get - по сокращенному URL получаем оригинальный
func TestFileDB_Get(t *testing.T) {
	// Создаем объект
	db := NewFileDB("tests.txt")
	ctx := context.Background()

	// Добавляем тестовые данные
	originURL := "longURL2"
	err := db.Add(ctx, "shortURL2", originURL, "token2")
	if err != nil {
		t.Errorf("Add return err %v;", err)
	}
	// Получаем оригинальный URL по сокрашенному
	data, err := db.Get(ctx, "shortURL2", "token2")
	if err != nil {
		t.Errorf("Add return err %v;", err)
	}
	assert.Equal(t, originURL, data)

	// Не верный URL отсутствует
	wrongShortURL := "wrongShortURL"
	_, err = db.Get(ctx, wrongShortURL, "token2")
	assert.NotEqual(t, fmt.Errorf("shorturl %s not found", wrongShortURL), err)

	err = os.Remove("tests.txt")
	if err != nil {
		assert.Errorf(t, err, "can't remove file")
	}
}

// TestFileDB_GetToken Проверяем наличие токена в БД
func TestFileDB_GetToken(t *testing.T) {
	// Создаем объект
	db := NewFileDB("tests.txt")
	ctx := context.Background()

	// Добавляем тестовые данные
	tokens := []string{"token3", "", "-1"}
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
		assert.Equal(t, false, tokenNotValid)
	}

	err = os.Remove("tests.txt")
	if err != nil {
		assert.Errorf(t, err, "can't remove file")
	}
}

// TestFileDB_GetUserURL проверяем получение всех URL пользователя
func TestFileDB_GetUserURL(t *testing.T) {
	// Создаем объект
	db := NewFileDB("tests.txt")
	ctx := context.Background()

	// Добавляем тестовые данные
	for i := 0; i < 3; i++ {
		err := db.Add(ctx, fmt.Sprintf("%s_%d", "shortURL_", i), fmt.Sprintf("%s_%d", "originURL_", i), "token")
		if err != nil {
			t.Errorf("Add demo data return err %v;", err)
		}
	}

	// Получаем все URL пользователя по токену
	data, err := db.GetUserURL(ctx, "token")
	if err != nil {
		t.Errorf("GetUserURL return err %v;", err)
	}
	// Проверяем что сокращенный и оригинальный URL совпадают с теми, которые были добавлены в БД
	for i, record := range data {
		assert.Equal(t, fmt.Sprintf("%s_%d", "originURL_", i), record.OriginURL)
		assert.Equal(t, fmt.Sprintf("%s_%d", "shortURL_", i), record.ShortURL)
	}

	err = os.Remove("tests.txt")
	if err != nil {
		assert.Errorf(t, err, "can't remove file")
	}
}

// TestFileDB_Ping Проверяем наличие соединения с БД
func TestFileDB_Ping(t *testing.T) {
	// Создаем объект
	db := NewFileDB("tests.txt")
	pingOk := db.Ping()
	assert.Equal(t, true, pingOk)
}

// Ниже тесты заглушки для функций реализующих обратную совместимость интерфейса с Postgres
// В оригинальном задании первой части курса для InMemory, fileDB их реализовывать не нужно,
// но код этот есть и учавствует в общем зачете покрытия тестами.

// TestFileDB_OriginURLExists -
func TestFileDB_OriginURLExists(t *testing.T) {
	// Создаем объект
	db := NewFileDB("test.txt")
	ctx := context.Background()

	ok, _ := db.OriginURLExists(ctx, "")
	assert.NotEqual(t, true, ok)
}

// TestFileDB_GetShortURLByIdentityPath -
func TestFileDB_GetShortURLByIdentityPath(t *testing.T) {
	// Создаем объект
	db := NewFileDB("test.txt")
	ctx := context.Background()

	num := db.GetShortURLByIdentityPath(ctx, "", "")
	assert.Equal(t, 0, num)
}

// TestFileDB_URLBulkDelete -
func TestFileDB_URLBulkDelete(t *testing.T) {
	// Создаем объект
	db := NewFileDB("test.txt")
	ctx := context.Background()
	ch := make(chan int)
	err := db.URLBulkDelete(ctx, ch)
	assert.Equal(t, nil, err)

	err = os.Remove("tests.txt")
	if err != nil {
		assert.Errorf(t, err, "can't remove file")
	}
}
