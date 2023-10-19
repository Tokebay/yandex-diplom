package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Tokebay/yandex-diplom/api/handlers"
)

var ErrUnauthorized = errors.New("unauthorized")

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверка токена
		cookie, err := r.Cookie(handlers.CookieName)
		if err != nil {
			//http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Printf("authMiddleware. cookie %s \n", cookie)

		userID, err := handlers.ExtractUserIDFromToken(cookie.Value)
		if err != nil {
			//http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		fmt.Printf("authMiddleware. userID %d \n", userID)
		//userID, err := handlers.GetUserCookie(r)
		//if err != nil || userID == -1 {
		//	//http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
		//	w.WriteHeader(http.StatusUnauthorized)
		//	return
		//}

		// Устанавливаем userID в контексте запроса
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		r = r.WithContext(ctx)

		// Прошло проверку, передаем запрос следующему обработчику
		next.ServeHTTP(w, r)
	})
}
