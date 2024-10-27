package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateAccessToken(accountID uint, roles []string) (string, error) {
    claims := jwt.MapClaims{
        "account_id": accountID,
        "roles":      roles,
        "exp":        time.Now().Add(time.Hour * 1).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))
}

func GenerateRefreshToken(accountID uint) (string, error) {
    claims := jwt.MapClaims{
        "account_id": accountID,
        "exp":        time.Now().Add(time.Hour * 24 * 7).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("REFRESH_TOKEN_SECRET")))
}

func ValidateToken(tokenString string, secret string) (*jwt.Token, error) {
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })
}
