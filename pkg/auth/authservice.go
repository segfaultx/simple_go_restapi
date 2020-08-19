package auth

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"os"
	"time"
)

var jwtKey = []byte(os.Getenv("API_SECRET"))

type (
	AuthenticationService interface {
		GenerateToken(credentials Credentials) (string, error)
		GetTokenFromString(tokenstring string) (*jwt.Token, error)
	}

	Credentials struct {
		Password string   `json:"password"`
		Username string   `json:"username"`
		Roles    []string `json:"roles"`
	}

	BasicJwtAuthService struct {
	}
)

func (authService *BasicJwtAuthService) GenerateToken(credentials Credentials) (string, error) {
	claims := make(jwt.MapClaims)
	claims["authorized"] = true
	claims["userId"] = credentials.Username
	claims["exp"] = time.Now().Add(10 * time.Minute).Unix()
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString(jwtKey)
}

func (authService *BasicJwtAuthService) GetTokenFromString(tokenstring string) (*jwt.Token, error) {
	token, ok := jwt.ParseWithClaims(tokenstring, &jwt.MapClaims{}, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
		}
		return jwtKey, nil
	})
	if ok != nil {
		log.Fatal(ok)
	}
	if _, err := token.Claims.(*jwt.MapClaims); err && token.Valid {
		return token, nil
	}
	return &jwt.Token{}, jwt.ErrSignatureInvalid
}
