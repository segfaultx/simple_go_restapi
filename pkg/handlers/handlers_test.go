package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/segfaultx/simple_rest/pkg/auth"
	"github.com/segfaultx/simple_rest/pkg/repo"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockRepo struct {
	Products []repo.Product
}

func (mockRepo *mockRepo) AddProduct(product repo.Product) error {
	if product.Id == -1 {
		return errors.New("-1 is the signal from test to throw an error")
	}
	mockRepo.Products = append(mockRepo.Products, product)
	return nil
}

func (mockRepo *mockRepo) UpdateProduct(product repo.Product) error {
	for _, item := range mockRepo.Products {
		if item.Id == product.Id {
			item.Name = product.Name
			return nil
		}
	}
	return errors.New("could not find requested item")
}

func (mockRepo *mockRepo) AllProducts() []repo.Product {
	return mockRepo.Products
}

func (mockRepo *mockRepo) InitRepo(user, passwd, dbname string) error {
	return nil
}

func (mockRepo *mockRepo) GetProductById(id int) (repo.Product, error) {
	if id != 1 {
		return repo.Product{}, errors.New("couldn't find the requested item")
	}
	return mockRepo.Products[0], nil
}

func (mockRepo *mockRepo) RemoveProduct(product repo.Product) error {
	for index, item := range mockRepo.Products {
		if item.Id == product.Id {
			mockRepo.Products = append(mockRepo.Products[:index], mockRepo.Products[index+1:]...)
			return nil
		}
	}
	return errors.New("couldn't remove product")
}

func (mockRepo *mockRepo) Close() {
	// dummy
}

var repository mockRepo

const contentTypeHeader = "Content-Type"
const contentType = "application/json"
const errorMsgStatusCode = "unexpected status code, got %d expected %d"
const errorMsgResponseBody = "unexpected response body, got %s wanted %s"
const baseUrl = "/catalog/products"

func initMockRepo() {
	testProducts := make([]repo.Product, 0)
	testProduct := repo.Product{
		Id:   1,
		Name: "Hosen",
	}
	testProducts = append(testProducts, testProduct)
	repository = mockRepo{
		Products: testProducts,
	}
}

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


func prepareAuthService() auth.AuthenticationService {
	return &auth.BasicJwtAuthService{Repo: &MockUserRepo{}}
}

func initRouter(handler http.HandlerFunc, method string) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc(baseUrl+"/{id}", handler).Methods(method)
	return router
}

func authenticate(req *http.Request, service auth.AuthenticationService){
	_ = service.RegisterUser("hugo", "test")
	token, _ := service.GenerateToken(auth.Credentials{Username: "hugo", Password: "test"})
	expiration := time.Now().Add(time.Minute * 10)
	cookie := http.Cookie{Name: "token",
		Value:    token,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   false,
		Path:     "/"}
	req.AddCookie(&cookie)
}

func TestMakeAllProductsHandlerGET(t *testing.T) {
	initMockRepo()
	service := prepareAuthService()
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := MakeAllProductsHandler(&repository, service)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf(errorMsgStatusCode, status, http.StatusOK)
	}
	expected, _ := json.Marshal(repository.Products)
	if response := rr.Body.String(); response != string(expected) {
		t.Errorf("unexpected responsebody, got %s wanted %s", response, expected)
	}
}

func TestMakeAllProductsHandlerPOST(t *testing.T) {
	initMockRepo()
	service := prepareAuthService()
	newProduct := repo.Product{Id: 2, Name: "Schuhe"}
	newProductJson, _ := json.Marshal(newProduct)
	reader := bytes.NewReader(newProductJson)
	req, err := http.NewRequest("POST", baseUrl, reader)
	if err != nil {
		t.Fatal(err)
	}
	authenticate(req, service)
	req.Header.Set(contentTypeHeader, contentType)
	rr := httptest.NewRecorder()
	handler := MakeAllProductsHandler(&repository, service)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf(errorMsgStatusCode, status, http.StatusOK)
		t.FailNow()
	}

	expected, _ := json.Marshal(newProduct)
	if response := rr.Body.String(); response != string(expected) {
		t.Errorf(errorMsgResponseBody, response, expected)
	}
}

func TestMakeAllProductsHandlerPOSTFail(t *testing.T) {
	initMockRepo()
	service := prepareAuthService()
	reader := bytes.NewReader([]byte("aiusazdvawldkab"))
	req, err := http.NewRequest("POST", baseUrl, reader)
	if err != nil {
		t.Fatal(err)
	}
	authenticate(req, service)
	req.Header.Set(contentTypeHeader, contentType)
	rr := httptest.NewRecorder()
	handler := MakeAllProductsHandler(&repository, service)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf(errorMsgStatusCode, status, http.StatusBadRequest)
	}
	failproduct, _ := json.Marshal(repo.Product{Id: -1, Name: "bloe"})
	reader = bytes.NewReader(failproduct)
	req, err = http.NewRequest("POST", baseUrl, reader)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(contentTypeHeader, contentType)
	rr = httptest.NewRecorder()
	authenticate(req, service)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("unexpected statuscode, got %d expected %d", status, http.StatusInternalServerError)
	}
}


func TestMakeProductsHandlerGET(t *testing.T) {
	initMockRepo()
	req, err := http.NewRequest("GET", baseUrl+"/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := MakeProductsHandler(&repository)
	router := initRouter(handler, "GET")
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf(errorMsgStatusCode, status, http.StatusOK)
		t.FailNow()
	}
	expected, _ := json.Marshal(repository.Products[0])
	if response := rr.Body.String(); response != string(expected) {
		t.Errorf(errorMsgResponseBody, response, expected)
	}

}

func TestMakeProductsHandlerGETNotExistingProduct(t *testing.T) {
	initMockRepo()
	req, err := http.NewRequest("GET", baseUrl+"/4", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := MakeProductsHandler(&repository)
	initRouter(handler, "GET").ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf(errorMsgStatusCode, status, http.StatusInternalServerError)
	}
}

func TestMakeProductsHandlerDELETE(t *testing.T) {
	initMockRepo()
	req, err := http.NewRequest("DELETE", baseUrl+"/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := MakeProductsHandler(&repository)
	initRouter(handler, "DELETE").ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf(errorMsgStatusCode, status, http.StatusOK)
	}

}

func TestMakeProductsHandlerDELETENotExistingProduct(t *testing.T) {
	initMockRepo()
	req, err := http.NewRequest("DELETE", baseUrl+"/5", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := MakeProductsHandler(&repository)
	initRouter(handler, "DELETE").ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf(errorMsgStatusCode, status, http.StatusInternalServerError)
	}
}

func TestMakeProductsHandlerPUT(t *testing.T) {
	initMockRepo()
	testProduct := repo.Product{Id: 1, Name: "Hemd"}
	testProductJson, err := json.Marshal(testProduct)
	if err != nil {
		t.Fatal(err)
	}
	reader := bytes.NewReader(testProductJson)
	req, err := http.NewRequest("PUT", baseUrl+"/1", reader)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(contentTypeHeader, contentType)

	rr := httptest.NewRecorder()
	handler := MakeProductsHandler(&repository)
	router := initRouter(handler, "PUT")
	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf(errorMsgStatusCode, status, http.StatusOK)
	}

}

func TestMakeProductsHandlerPUTFAILBadRequestBody(t *testing.T) {
	initMockRepo()
	reader := bytes.NewReader([]byte("asdasdajd"))
	req, err := http.NewRequest("PUT", baseUrl+"/1", reader)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := MakeProductsHandler(&repository)
	initRouter(handler, "PUT").ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf(errorMsgStatusCode, status, http.StatusInternalServerError)
	}
}

func TestMakeProductsHandlerPUTFAILBadRequestID(t *testing.T) {
	initMockRepo()
	testProduct := repo.Product{Id: 1, Name: "Hemd"}
	testProductJson, err := json.Marshal(testProduct)
	if err != nil {
		t.Fatal(err)
	}
	reader := bytes.NewReader(testProductJson)
	req, err := http.NewRequest("PUT", baseUrl+"/29", reader)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := MakeProductsHandler(&repository)
	initRouter(handler, "PUT").ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf(errorMsgStatusCode, status, http.StatusInternalServerError)
	}
}

func TestMakeProductsHandlerBADURLParam(t *testing.T) {
	req, err := http.NewRequest("PUT", baseUrl+"/abc", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := MakeProductsHandler(&repository)
	initRouter(handler, "PUT").ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf(errorMsgStatusCode, status, http.StatusBadRequest)
	}
}