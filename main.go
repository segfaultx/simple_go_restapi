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
						w.WriteHeader(200)
						_, err = w.Write(resp)
						if err != nil {
							w.WriteHeader(500)
							_, _ = w.Write([]byte("Error writing response"))
							log.Fatal(err)
						}
						break
					}
				}
			}
		case "DELETE":
			{
				dberr := repository.RemoveProduct(repo.Product{Id: id})
				if dberr != nil {
					w.WriteHeader(500)
					_, _ = w.Write([]byte("Error removing Product"))
					log.Fatal(err)
				}
			}
		case "PUT":
			{
				product := repo.Product{}
				err := decodeRequestBody(&product, r)
				if err != nil {
					w.WriteHeader(500)
					_, _ = w.Write([]byte("Malformed JSON request"))
					log.Fatal(err)
				}
				product.Id = id
				err = repository.UpdateProduct(product)
				if err != nil {
					w.WriteHeader(500)
					_, _ = w.Write([]byte("Error updating product"))
					log.Fatal(err)
				}
				w.WriteHeader(200)
				resp, _ := json.Marshal(product)
				_, _ = w.Write(resp)

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
				_, err := writer.Write(resp)
				if err != nil {
					log.Fatal(err)
				}
			}
		case "POST":
			{
				product := repo.Product{}
				err := decodeRequestBody(&product, request)
				if err != nil {
					writer.WriteHeader(500)
					log.Fatal(err)
				}
				err = repository.AddProduct(product)
				if err != nil {
					_, _ = writer.Write([]byte("Error adding Product to database"))
					writer.WriteHeader(500)
				}
				respo, _ := json.Marshal(product)
				_, _ = writer.Write(respo)
			}
		}
	}
}

func decodeRequestBody(t interface{}, request *http.Request) error {
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(t)
	return err
}

func main() {
	r := mux.NewRouter()
	repository := repo.DefaultRepository{Products: nil}
	repository.InitRepo()
	r.HandleFunc("/catalog/products/{id}", MakeProductsHandler(&repository)).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/catalog/products", MakeAllProductsHandler(&repository)).Methods("GET", "POST")
	_ = http.ListenAndServe(":8080", r)
}
