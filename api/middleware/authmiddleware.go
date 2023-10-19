package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/Tokebay/yandex-diplom/api/handlers"
)

var ErrUnauthorized = errors.New("unauthorized")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := handlers.GetUserCookie(r)
		if err != nil || userID == -1 {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		// Проверка токена
		cookie, err := r.Cookie(handlers.CookieName)
		if err != nil {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		_, err = handlers.ExtractUserIDFromToken(cookie.Value)
		if err != nil {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		// Устанавливаем userID в контексте запроса
		ctx := context.WithValue(r.Context(), "userID", userID)
		r = r.WithContext(ctx)

		// Прошло проверку, передаем запрос следующему обработчику
		next.ServeHTTP(w, r)
	})
}
