package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/segfaultx/simple_rest/pkg/auth"
	"github.com/segfaultx/simple_rest/pkg/repo"
	"log"
	"net/http"
	"strconv"
)

func MakeProductsHandler(repository repo.ProductRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch r.Method {
		case "GET":
			handleGet(repository, w, id)
		case "DELETE":
			handleDelete(repository, w, id)
		case "PUT":
			handlePut(repository, w, r, id)
		}
	}
}

func handleGet(repository repo.ProductRepository, w http.ResponseWriter, id int) {
	product, err := repository.GetProductById(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("couldn't find the requested object"))
		return
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
		log.Print(err)
	}
}

func handlePut(repository repo.ProductRepository, w http.ResponseWriter, r *http.Request, id int) {
	product := repo.Product{}
	err := decodeRequestBody(&product, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Malformed JSON request"))
		log.Print(err)
	}
	product.Id = id
	err = repository.UpdateProduct(product)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Error updating product"))
		log.Print(err)
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
				_, _ = writer.Write(resp)
			}
		case "POST":
			{
				product := repo.Product{}
				err := decodeRequestBody(&product, request)
				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					_, _ = writer.Write([]byte(err.Error()))
					return
				}
				// TODO: add min length to db schema
				if len(product.Name) <= 3 {
					writer.WriteHeader(http.StatusBadRequest)
					_, _ = writer.Write([]byte("invalid product name"))
					return
				}
				err = repository.AddProduct(product)
				if err != nil {
					writer.WriteHeader(http.StatusInternalServerError)
					_, _ = writer.Write([]byte("Error adding Product to database"))
					return
				}
				resp, _ := json.Marshal(product)
				_, _ = writer.Write(resp)
			}
		}
	}
}

func MakeAuthenticationHandler(service auth.AuthenticationService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "POST":
			handleAuthPost(service, writer, request)
		}
	}
}

func handleAuthPost(service auth.AuthenticationService, writer http.ResponseWriter, r *http.Request) {
	credentials := auth.Credentials{}
	err := decodeRequestBody(&credentials, r)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}
	err = service.RegisterUser(credentials.Username, credentials.Password)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func decodeRequestBody(t interface{}, request *http.Request) error {
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(t)
	return err
}
