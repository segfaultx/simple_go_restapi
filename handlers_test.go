package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"gitlab.com/amatusm/simpleREST/pkg/repo"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockRepo struct {
	Products []repo.Product
}

func (r *mockRepo) AddProduct(product repo.Product) error {
	if product.Id == -1 {
		return errors.New("-1 is the signal from test to throw an error")
	}
	r.Products = append(r.Products, product)
	return nil
}

func (r *mockRepo) UpdateProduct(product repo.Product) error {
	for _, item := range r.Products {
		if item.Id == product.Id {
			item.Name = product.Name
			return nil
		}
	}
	return errors.New("could not find requested item")
}

func (r *mockRepo) AllProducts() []repo.Product {
	return r.Products
}

func (r *mockRepo) InitRepo() {
	return
}

func (r *mockRepo) GetProductById(id int) (repo.Product, error) {
	if id != 1 {
		return repo.Product{}, errors.New("couldn't find the requested item")
	}
	return r.Products[0], nil
}

func (r *mockRepo) RemoveProduct(product repo.Product) error {
	for index, item := range r.Products {
		if item.Id == product.Id {
			r.Products = append(r.Products[:index], r.Products[index+1:]...)
			return nil
		}
	}
	return errors.New("couldn't remove product")
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

func initRouter(handler http.HandlerFunc, method string) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc(baseUrl+"/{id}", handler).Methods(method)
	return router
}

func TestMakeAllProductsHandlerGET(t *testing.T) {
	initMockRepo()

	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := MakeAllProductsHandler(&repository)
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

	newProduct := repo.Product{Id: 2, Name: "Schuhe"}
	newProductJson, _ := json.Marshal(newProduct)
	reader := bytes.NewReader(newProductJson)
	req, err := http.NewRequest("POST", baseUrl, reader)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(contentTypeHeader, contentType)
	rr := httptest.NewRecorder()
	handler := MakeAllProductsHandler(&repository)
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
	reader := bytes.NewReader([]byte("aiusazdvawldkab"))
	req, err := http.NewRequest("POST", baseUrl, reader)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(contentTypeHeader, contentType)
	rr := httptest.NewRecorder()
	handler := MakeAllProductsHandler(&repository)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf(errorMsgStatusCode, status, http.StatusInternalServerError)
	}
	failproduct, _ := json.Marshal(repo.Product{Id: -1, Name: "bloe"})
	reader = bytes.NewReader(failproduct)
	req, err = http.NewRequest("POST", baseUrl, reader)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(contentTypeHeader, contentType)
	rr = httptest.NewRecorder()
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