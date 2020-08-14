package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/segfaultx/simple_rest/pkg/handlers"
	"github.com/segfaultx/simple_rest/pkg/repo"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	router := mux.NewRouter()
	repository := repo.DefaultRepository{Products: nil}
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt)
	defer log.Println("done")
	defer func() {
		r := recover()
		if r != nil {
			log.Fatal(r)
		}
	}()
	defer repository.Close()
	user := os.Getenv("POSTGRES_USER")
	passwd := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DBNAME")
	err := repository.InitRepo(user, passwd, dbname)
	if err != nil {
		panic(err)
	}
	router.HandleFunc("/catalog/products/{id}", handlers.MakeProductsHandler(&repository)).Methods("GET", "DELETE", "PUT")
	router.HandleFunc("/catalog/products", handlers.MakeAllProductsHandler(&repository)).Methods("GET", "POST")
	server := &http.Server{Addr: ":8080", Handler: router}
	go func() {
		log.Println("starting API server...")
		if err := server.ListenAndServe(); err != nil {
			log.Println("shutting down server, cleaning up...")
		}
	}()

	<- stopSignal
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}
}
