package main

import (
	"github.com/gorilla/mux"
	"gitlab.com/amatusm/simpleREST/pkg/repo"
	"log"
	"net/http"
	"os"
)

func main() {
	r := mux.NewRouter()
	repository := repo.DefaultRepository{Products: nil}
	defer repository.Close()
	user := os.Getenv("POSTGRES_USER")
	passwd := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DBNAME")
	err := repository.InitRepo(user, passwd, dbname)
	if err != nil {
		log.Fatal(err)
	}
	r.HandleFunc("/catalog/products/{id}", MakeProductsHandler(&repository)).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/catalog/products", MakeAllProductsHandler(&repository)).Methods("GET", "POST")
	_ = http.ListenAndServe(":8080", r)
}
