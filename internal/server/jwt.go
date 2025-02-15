package server

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
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

func (s *Server) createJWT() (string, error) {
	userID := uuid.New().ClockSequence()

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		//generate userID
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}

func (s *Server) getUserIDFromCookie(cookie *http.Cookie) (int, error) {

	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}
	return 0, errors.Wrapf(err, "get userID from cookie %s", cookieName)
}

func (s *Server) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			jwtToken string
		)

		cookie, err := c.Cookie(cookieName)
		if err != nil {
			jwtToken, err = s.createJWT()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, "middleware err generate token")
			}

			cookie = &http.Cookie{
				Name:     cookieName,
				Value:    jwtToken,
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(24 * time.Hour),
			}

			c.SetCookie(cookie)

			userID, err := s.getUserIDFromCookie(cookie)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, "Getting user from cookie")
			}

			c.Set("userID", userID)
		}

		userID, err := s.getUserIDFromCookie(cookie)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Getting user from cookie")
		}

		c.Set("userID", userID)

		return next(c)
	}
}
