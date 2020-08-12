package main

import (
	"bytes"
	"encoding/json"
	"errors"
	repo "gitlab.com/amatusm/simpleREST/pkg/repo"
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
	return repo.Product{}, errors.New("NYI")
}

func (r *mockRepo) RemoveProduct(product repo.Product) error {
	for index, item := range r.Products {
		if (item.Id == product.Id) && (item.Name == product.Name) {
			r.Products = append(r.Products[:index], r.Products[index+1:]...)
			return nil
		}
	}

	return errors.New("couldn't update product")
}

var repository mockRepo
const contentTypeHeader = "Content-Type"
const contentType = "application/json"
const errorMsgStatuscode = "unexpected status code, got %d expected %d"
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
		t.Errorf(errorMsgStatuscode, status, http.StatusOK)
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
	newProducts := make([]repo.Product, 0)
	newProducts = append(newProducts, repository.Products...)

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
		t.Errorf(errorMsgStatuscode, status, http.StatusOK)
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
		t.Errorf(errorMsgStatuscode, status, http.StatusInternalServerError)
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

func TestMakeProductsHandler(t *testing.T) {
	initMockRepo()

}
