package main

import (
	"github.com/gorilla/mux"
	"github.com/segfaultx/simple_rest/pkg/handlers"
	"github.com/segfaultx/simple_rest/pkg/repo"
	"log"
	"net/http"
	"os"
)

func main() {
	r := mux.NewRouter()
	repository := repo.DefaultRepository{Products: nil}
	defer repository.Close()
	defer func() {
		r := recover(); if r != nil {
			log.Fatal(r)
		}
	}()
	user := os.Getenv("POSTGRES_USER")
	passwd := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DBNAME")
	err := repository.InitRepo(user, passwd, dbname)
	if err != nil {
		panic(err)
	}
	r.HandleFunc("/catalog/products/{id}", handlers.MakeProductsHandler(&repository)).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/catalog/products", handlers.MakeAllProductsHandler(&repository)).Methods("GET", "POST")
	_ = http.ListenAndServe(":8080", r)
}
