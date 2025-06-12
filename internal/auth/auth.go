// Package auth provides utilities for creating and parsing JWT tokens
// used for user authentication in the URL shortener service.
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// SecretKey is the secret key used to sign JWT tokens.
//
// In production, this should be replaced with a secure secret stored outside of the source code.
var SecretKey = []byte("supersecretkey")

// CookieName is the name of the cookie used to store the JWT token in HTTP responses.
var CookieName = "auth_token"

// Claims represents the JWT payload used for authentication.
//
// It embeds jwt.RegisteredClaims and adds a custom UserID field.
type Claims struct {
	jwt.RegisteredClaims
	UserID int `json:"user_id"`
}

// CreateJWT generates a new JWT token with:
// - a randomly generated UserID (based on UUID ClockSequence),
// - an expiration time of 24 hours.
//
// Returns the signed token string or an error if the token could not be generated.
func CreateJWT() (string, error) {
	userID := uuid.New().ClockSequence()

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(SecretKey)
}

// ParseJWT parses the given JWT token string and extracts the UserID.
//
// If the token is valid and correctly signed, returns the UserID from the token claims.
// If the token is invalid, returns an error.
//
// The token must be signed using the HMAC SHA-256 algorithm (HS256).
func ParseJWT(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return SecretKey, nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return 0, jwt.ErrSignatureInvalid
	}

	return claims.UserID, nil
}
