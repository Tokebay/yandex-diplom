package middleware

import (
	"context"
	"net/http"

	"github.com/Tokebay/yandex-diplom/api/handlers"
	"github.com/Tokebay/yandex-diplom/domain/models"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверка токена
		cookie, err := r.Cookie(handlers.CookieName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userID, err := handlers.ExtractUserIDFromToken(cookie.Value)
		if err != nil {
			//http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Устанавливаем userID в контексте запроса
		ctx := context.WithValue(r.Context(), models.UserIDKey, userID)
		r = r.WithContext(ctx)

		// Прошло проверку, передаем запрос следующему обработчику
		next.ServeHTTP(w, r)
	})
}
