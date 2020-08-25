package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/segfaultx/simple_rest/pkg/auth"
	"github.com/segfaultx/simple_rest/pkg/repo"
	"log"
	"net/http"
	"strconv"
	"time"
)

func MakeProductsHandler(repository repo.ProductRepository) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			log.Print(err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		switch request.Method {
		case "GET":
			handleGet(repository, writer, id)
		case "DELETE":
			handleDelete(repository, writer, id)
		case "PUT":
			handlePut(repository, writer, request, id)
		}
	}
}

func handleGet(repository repo.ProductRepository, writer http.ResponseWriter, id int) {
	product, err := repository.GetProductById(id)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte("couldn't find the requested object"))
		return
	}
	resp, _ := json.Marshal(product)
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(resp)
}

func handleDelete(repository repo.ProductRepository, writer http.ResponseWriter, id int) {
	err := repository.RemoveProduct(repo.Product{Id: id})
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte("Error removing Product"))
		log.Print(err)
	}
}

func handlePut(repository repo.ProductRepository, writer http.ResponseWriter, request *http.Request, id int) {
	product := repo.Product{}
	err := decodeRequestBody(&product, request)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte("Malformed JSON request"))
		log.Print(err)
	}
	product.Id = id
	err = repository.UpdateProduct(product)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte("Error updating product"))
		log.Print(err)
	}
	writer.WriteHeader(http.StatusOK)
	resp, _ := json.Marshal(product)
	_, _ = writer.Write(resp)
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

func MakeRegisterHandler(service auth.AuthenticationService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		credentials := auth.Credentials{}
		err := decodeRequestBody(&credentials, request)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		err = service.RegisterUser(credentials.Username, credentials.Password)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}

func MakeLoginHandler(service auth.AuthenticationService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		credentials := auth.Credentials{}
		err := decodeRequestBody(&credentials, request)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		token, err := service.GenerateToken(credentials)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		expiration := time.Now().Add(time.Minute * 5)
		cookie := http.Cookie{Name: "token", Value: token, Expires: expiration, HttpOnly: true, Secure: true}
		http.SetCookie(writer, &cookie)
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(token))
	}
}

func decodeRequestBody(t interface{}, request *http.Request) error {
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(t)
	return err
}
