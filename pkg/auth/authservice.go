package auth

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/segfaultx/simple_rest/pkg/repo"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
	"time"
)

var jwtKey = []byte(os.Getenv("API_SECRET"))

type (
	AuthenticationService interface {
		GenerateToken(credentials Credentials) (string, error)
		GetTokenFromString(tokenString string) (*jwt.Token, error)
		RegisterUser(username, password string) error
		RefreshToken(token *jwt.Token) (string, error)
	}

	Credentials struct {
		Password string `json:"password"`
		Username string `json:"username"`
	}

	BasicJwtAuthService struct {
		Repo repo.UserRepository
	}
)

func New(repository repo.UserRepository) AuthenticationService {
	authService := new(BasicJwtAuthService)
	authService.Repo = repository
	return authService
}

func (authService *BasicJwtAuthService) RegisterUser(username, password string) error {
	_, err := authService.Repo.GetByUsername(username)
	if err == nil {
		return errors.New("username already taken")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	usr := repo.User{Username: username, Password: string(hashedPassword), Role: repo.USER}
	return authService.Repo.AddUser(usr)
}

func (authService *BasicJwtAuthService) GenerateToken(credentials Credentials) (string, error) {
	usr, err := authService.Repo.GetByUsername(credentials.Username)
	if err != nil {
		return "", err
	}
	err = checkPassword(usr, credentials)
	if err != nil {
		return "", err
	}
	claims := make(jwt.MapClaims)
	claims["authorized"] = true
	claims["userId"] = usr.Username
	claims["role"] = usr.Role
	claims["exp"] = time.Now().Add(10 * time.Minute).Unix()
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString(jwtKey)
}

func checkPassword(user repo.User, credentials Credentials) error {
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
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

func (authService *BasicJwtAuthService) RefreshToken(token *jwt.Token) (string, error) {
	claims := *(token.Claims.(*jwt.MapClaims))
	refreshClaims := make(jwt.MapClaims)
	refreshClaims["authorized"] = claims["authorized"]
	refreshClaims["userId"] = claims["userId"]
	refreshClaims["role"] = claims["role"]
	refreshClaims["exp"] = time.Now().Add(10 * time.Minute).Unix()
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	return refreshToken.SignedString(jwtKey)
}
