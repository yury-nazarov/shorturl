package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"

	"github.com/yury-nazarov/shorturl/internal/app/storage/repository"
)

// HTTPCookieAuth - middleware - устанавливает подписаный токен для клиента, ели его нет.
func HTTPCookieAuth(db repository.Repository) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Получаем токен из Request
			token, err := r.Cookie("session_token")
			// Если токена нет
			if err != nil {
				// Генерим, шифруем и устанавливаем в куку
				cookieToken := setCookieEncryptToken()
				// Добавляем куку в Response для ответа клиенту
				http.SetCookie(w, cookieToken)
				// Добавляем куку в Request для дальнейшей обработке в хендлерах и добавления в БД
				r.AddCookie(cookieToken)
				next.ServeHTTP(w, r)
				return
			}

			// Проверяем валидность токена, найдя его в БД
			tokenExist, err := db.GetToken(token.Value)
			if err != nil {
				log.Print(err)
			}
			if !tokenExist {
				cookieToken := setCookieEncryptToken()
				http.SetCookie(w, cookieToken)
				r.AddCookie(cookieToken)
				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
			return
		}
		return http.HandlerFunc(fn)
	}
}

// setCookieEncryptToken - Генерит новый токен, шифрует, устанавливает в cookie
func setCookieEncryptToken() *http.Cookie{
	// Если токена нет - генерим, подписываем, добавляем в куку и передаем HTTP Request дальше
	uuid := uniqueUserID()
	// Подписываем его
	sessionToken := encryptToken([]byte(uuid))

	// Устанавливаем cookie
	cookieToken := &http.Cookie{
		Name: "session_token",
		Value: sessionToken,
	}
	return cookieToken
}

// uniqueUserID - Генерит рандомный токен для пользователя
func  uniqueUserID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		// TODO: Как правильно обрабатывать ошибки в mw ?
		log.Print(err)
	}
	return hex.EncodeToString(uuid)
}

func encryptToken(uuid []byte) string{
	// TODO: Вынести в конфиг ключ
	key := []byte("qwe")
	h := hmac.New(sha256.New, key)
	h.Write(uuid)
	dst := h.Sum(nil)

	return hex.EncodeToString(dst)
}

