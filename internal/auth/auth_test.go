package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

func TestCreateAndParseJWT(t *testing.T) {

	tokenStr, err := CreateJWT()
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	userID, err := ParseJWT(tokenStr)
	require.NoError(t, err)
	require.NotZero(t, userID)
}

func TestParseJWT_InvalidToken(t *testing.T) {

	invalidToken := "this.is.not.a.valid.token"

	userID, err := ParseJWT(invalidToken)
	require.Error(t, err)
	require.Equal(t, 0, userID)
}

func TestParseJWT_ExpiredToken(t *testing.T) {

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
		UserID: 42,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(SecretKey)
	require.NoError(t, err)

	userID, err := ParseJWT(tokenStr)
	require.Error(t, err)
	require.Equal(t, 0, userID)
}
