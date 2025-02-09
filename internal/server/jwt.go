package server

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	//TODO: move to cfg!
	secretKey  = []byte("supersecretkey")
	cookieName = "auth_token"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int `json:"user_id"`
}

// Функция создания JWT
func createJWT(userID int) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}

func getUserIDFromCookie(c echo.Context) (int, error) {
	cookie, err := c.Cookie(cookieName)
	if err != nil {
		return 0, err
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
}

func authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			jwtToken string
			err      error
		)

		_, err = c.Cookie(cookieName)
		if err != nil {
			userID := uuid.New().ClockSequence()
			jwtToken, err = createJWT(userID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "err generate token")
			}

			c.SetCookie(&http.Cookie{
				Name:     cookieName,
				Value:    jwtToken,
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(24 * time.Hour),
			})
		}
		return next(c)
	}
}
