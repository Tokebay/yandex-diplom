package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

const TokenExp = time.Hour * 3
const SecretKey = "shortURL"
const CookieName = "token"

var ErrToken = errors.New("invalid token")
var ErrParseClaims = errors.New("error ParseWithClaims")
var ErrSignTokenString = errors.New("error create token string")

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func GetUserCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", nil
		}
		logger.Log.Error("GetUserCookie. error get cookie", zap.Error(err))
		return "", err
	}

	userID, err := ExtractUserIDFromToken(cookie.Value)
	if err != nil {
		return "", fmt.Errorf("error get userID: %w", err)
	}

	return userID, nil
}

func ExtractUserIDFromToken(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(SecretKey), nil
	})
	if err != nil {
		return "", ErrParseClaims
	}

	if !token.Valid {
		return "", ErrToken
	}

	return claims.UserID, nil
}

func BuildJWTString(userID string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", ErrSignTokenString
	}

	return tokenString, nil
}
