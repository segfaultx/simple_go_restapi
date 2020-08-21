package auth

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/segfaultx/simple_rest/pkg/repo"
	"log"
	"os"
	"time"
)

var jwtKey = []byte(os.Getenv("API_SECRET"))

type (
	AuthenticationService interface {
		GenerateToken(credentials Credentials) (string, error)
		GetTokenFromString(tokenString string) (*jwt.Token, error)
	}

	Credentials struct {
		Password string `json:"password"`
		Username string `json:"username"`
		Role     Role `json:"roles"`
	}

	Role string

	BasicJwtAuthService struct {
		repo repo.UserRepository
	}
)

const (
	ADMIN Role = "ADMIN"
	USER Role = "USER"
)

func New(repository repo.UserRepository) AuthenticationService {
	authService := new(BasicJwtAuthService)
	authService.repo = repository
	return authService
}

func (authService *BasicJwtAuthService) GenerateToken(credentials Credentials) (string, error) {
	claims := make(jwt.MapClaims)
	claims["authorized"] = true
	claims["userId"] = credentials.Username
	claims["exp"] = time.Now().Add(10 * time.Minute).Unix()
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString(jwtKey)
}

func (authService *BasicJwtAuthService) GetTokenFromString(tokenString string) (*jwt.Token, error) {
	token, ok := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(tok *jwt.Token) (interface{}, error) {
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
