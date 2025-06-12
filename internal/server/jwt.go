package server

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/wickedv43/go-shortener/internal/auth"
)

// authMiddleware is an Echo middleware that checks for a valid JWT cookie.
// If the cookie does not exist, it creates one and assigns a new user ID.
// The user ID is stored in the context for downstream handlers.
func (s *Server) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var jwtToken string

		cookie, err := c.Cookie(auth.CookieName)
		if err != nil {
			jwtToken, err = auth.CreateJWT()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, "middleware err generate token")
			}

			cookie = &http.Cookie{
				Name:     auth.CookieName,
				Value:    jwtToken,
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(24 * time.Hour),
			}

			c.SetCookie(cookie)

			//var userID int
			//userID, err = auth.ParseJWT(cookie.Value)
			//if err != nil {
			//	return c.JSON(http.StatusInternalServerError, "Getting user from cookie")
			//}
			//
			//c.Set("userID", userID)
		}

		userID, err := auth.ParseJWT(cookie.Value)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Getting user from cookie")
		}

		c.Set("userID", userID)

		return next(c)
	}
}
