package auth

import (
	"errors"
	"github.com/segfaultx/simple_rest/pkg/repo"
	"reflect"
	"strings"
	"testing"
)

type MockUserRepo struct {
	Users []repo.User
}

func (mockRepo *MockUserRepo) AddUser(u repo.User) error {
	if u.Username == "fail" {
		return errors.New("should fail here")
	}
	mockRepo.Users = append(mockRepo.Users, u)
	return nil
}

func (mockRepo *MockUserRepo) GetByUsername(username string) (repo.User, error) {
	for _, user := range mockRepo.Users {
		if user.Username == username {
			return user, nil
		}
	}
	return repo.User{}, errors.New("user not found")
}

const (
	TokenFormatLength = 3
)

func prepareAuthService() AuthenticationService {
	return &BasicJwtAuthService{repo: &MockUserRepo{}}
}

func TestNew(t *testing.T) {
	mockRepo := &MockUserRepo{}
	service := New(mockRepo)
	if _, ok := service.(*BasicJwtAuthService); !ok {
		t.Fail()
	}
}

func TestBasicJwtAuthService_RegisterUser(t *testing.T) {
	service := prepareAuthService()
	err := service.RegisterUser("test", "hallo123")
	if err != nil {
		t.Errorf("expected %v, received %s ", nil, err)
		t.FailNow()
	}
	err = service.RegisterUser("fail", "")

	if err == nil {
		t.Errorf("expected %s, received %v", errors.New("user not found"), err)
		t.FailNow()
	}
}

func TestBasicJwtAuthService_GenerateToken(t *testing.T) {
	service := prepareAuthService()
	err := service.RegisterUser("test", "hallo123")
	if err != nil {
		t.Errorf("expected %v, received %s ", nil, err)
		t.FailNow()
	}
	creds := Credentials{Username: "test", Password: "hallo123"}
	token, err := service.GenerateToken(creds)
	if err != nil {
		t.Errorf("expected %v, received %s ", nil, err)
		t.FailNow()
	}
	if reflect.TypeOf(token).String() != "string" {
		t.Errorf("expected string, received %s ", reflect.TypeOf(token).String())
		t.FailNow()
	}
	if len(strings.Split(token, ".")) != TokenFormatLength {
		t.Error("unknown token format")
		t.FailNow()
	}
}

func TestBasicJwtAuthService_GetTokenFromString(t *testing.T) {
	service := prepareAuthService()
	err := service.RegisterUser("test", "hallo123")
	if err != nil {
		t.Errorf("expected %v, received %s ", nil, err)
		t.FailNow()
	}
	creds := Credentials{Username: "test", Password: "hallo123"}
	tokenString, err := service.GenerateToken(creds)
	if err != nil {
		t.Errorf("expected %v, received %s ", nil, err)
		t.FailNow()
	}
	token, err := service.GetTokenFromString(tokenString)
	if err != nil {
		t.Errorf("expected %v, received %v", nil, err)
		t.FailNow()
	}
	if !token.Valid {
		t.Errorf("expected %v, received %v", true, token.Valid)
		t.FailNow()
	}
}

func TestBasicJwtAuthService_RegisterUser_Username_taken(t *testing.T) {
	service := prepareAuthService()
	err := service.RegisterUser("hugo", "")
	if err != nil {
		t.Errorf("expected %v, received %v", nil, err)
		t.FailNow()
	}
	err = service.RegisterUser("hugo", "")
	if err == nil {
		t.Errorf("expected %v, received %v", errors.New("username already taken"), err)
		t.FailNow()
	}
}

func TestBasicJwtAuthService_GenerateToken_Invalid_Username(t *testing.T) {
	service := prepareAuthService()
	creds := Credentials{Username: "fail", Password: ""}
	_, err := service.GenerateToken(creds)
	if err == nil {
		t.Errorf("expected %v, received %v", errors.New("user not found"), err)
		t.FailNow()
	}
}

func TestBasicJwtAuthService_GenerateToken_Invalid_Password(t *testing.T) {
	service := prepareAuthService()
	err := service.RegisterUser("hugo", "test")
	if err != nil {
		t.Errorf("expected %v, received %v", nil, err)
		t.FailNow()
	}
	creds := Credentials{Username: "hugo", Password: "test123"}
	_, err = service.GenerateToken(creds)
	if err == nil {
		t.Errorf("expected %v, received %v", errors.New("user not found"), err)
		t.FailNow()
	}
}