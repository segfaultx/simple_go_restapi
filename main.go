package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gitlab.com/amatusm/simpleREST/pkg/repo"
	"log"
	"net/http"
	"strconv"
)

func MakeProductsHandler(repository repo.ProductRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			log.Fatal(err)
		}
		switch r.Method {
		case "GET":
			{
				for _, item := range repository.AllProducts() {
					if item.Id == id {
						resp, _ := json.Marshal(item)
						_, err2 := w.Write(resp)
						if err2 != nil {
							log.Fatal(err2)
						}
						break
					}
				}
			}
		case "DELETE":
			{
				dberr := repository.RemoveProduct(repo.Product{Id: id})
				if dberr != nil {
					log.Fatal(err)
				}
			}

		}

	}
}

func MakeAllProductsHandler(repository repo.ProductRepository) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "GET":
			{
				products := repository.AllProducts()
				resp, _ := json.Marshal(products)
				_, ok := writer.Write(resp)
				if ok != nil {
					log.Fatal(ok)
				}
			}
		case "POST":
			{
				product := repo.Product{}
				dec := json.NewDecoder(request.Body)
				err := dec.Decode(&product)
				if err != nil {
					writer.WriteHeader(500)
					log.Fatal(err)
				}
				repository.AddProduct(product)
				respo, _ := json.Marshal(product)
				_, _ = writer.Write(respo)
			}
		}
	}
}

func main() {
	r := mux.NewRouter()
	repository := repo.DefaultRepository{Products: nil}
	repository.InitRepo()
	r.HandleFunc("/catalog/products/{id}", MakeProductsHandler(&repository)).Methods("GET", "DELETE")
	r.HandleFunc("/catalog/products", MakeAllProductsHandler(&repository)).Methods("GET", "POST")
	_ = http.ListenAndServe(":8080", r)
}
