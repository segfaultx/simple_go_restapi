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
				handleGet(repository, w, id)
				break
			}

		case "DELETE":
			{
				handleDelete(repository, w, id)
			}
		case "PUT":
			{
				handlePut(repository, w, r, id)
			}

		}

	}
}

func handleGet(repository repo.ProductRepository, w http.ResponseWriter, id int) {
	product, err := repository.GetProductById(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("couldn't find the requested object"))
	}
	resp, _ := json.Marshal(product)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

func handleDelete(repository repo.ProductRepository, w http.ResponseWriter, id int) {
	err := repository.RemoveProduct(repo.Product{Id: id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Error removing Product"))
		log.Fatal(err)
	}
}

func handlePut(repository repo.ProductRepository, w http.ResponseWriter, r *http.Request, id int) {
	product := repo.Product{}
	err := decodeRequestBody(&product, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Malformed JSON request"))
		log.Fatal(err)
	}
	product.Id = id
	err = repository.UpdateProduct(product)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Error updating product"))
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusOK)
	resp, _ := json.Marshal(product)
	_, _ = w.Write(resp)
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
					writer.WriteHeader(http.StatusInternalServerError)
					_, err = writer.Write([]byte(err.Error()))
					if err != nil {
						log.Fatal(err)
					}
					return
				}
				err = repository.AddProduct(product)
				if err != nil {
					writer.WriteHeader(http.StatusInternalServerError)
					_, _ = writer.Write([]byte("Error adding Product to database"))
					return
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
