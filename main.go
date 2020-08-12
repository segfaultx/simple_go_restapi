package main

import (
	"github.com/gorilla/mux"
	"gitlab.com/amatusm/simpleREST/pkg/repo"
	"net/http"
)


func main() {
	r := mux.NewRouter()
	repository := repo.DefaultRepository{Products: nil}
	repository.InitRepo()
	r.HandleFunc("/catalog/products/{id}", MakeProductsHandler(&repository)).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/catalog/products", MakeAllProductsHandler(&repository)).Methods("GET", "POST")
	_ = http.ListenAndServe(":8080", r)
}
