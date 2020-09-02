package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/segfaultx/simple_rest/pkg/auth"
	"github.com/segfaultx/simple_rest/pkg/handlers"
	"github.com/segfaultx/simple_rest/pkg/repo"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func setupRepo() *repo.DefaultRepository {
	repository := repo.New()
	user := os.Getenv("POSTGRES_USER")
	passwd := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DBNAME")
	err := repository.InitRepo(user, passwd, dbname)
	if err != nil {
		panic(err)
	}
	return repository
}

func errorFunc() {
	r := recover()
	if r != nil {
		log.Fatal(r)
	}
}

func setupRoutes(router *mux.Router, repository repo.ProductRepository, service auth.AuthenticationService) {
	router.HandleFunc("/catalog/products/{id}", handlers.MakeProductsHandler(repository)).Methods("GET", "DELETE", "PUT")
	router.HandleFunc("/catalog/products", handlers.MakeAllProductsHandler(repository, service)).Methods("GET", "POST")
	router.HandleFunc("/register", handlers.MakeRegisterHandler(service)).Methods("POST")
	router.HandleFunc("/login", handlers.MakeLoginHandler(service)).Methods("POST")
}

func listenAndServe(server *http.Server) {
	log.Println("starting API server...")
	if err := server.ListenAndServe(); err != nil {
		log.Println("shutting down server, cleaning up...")
	}
}

func shutdownOnInterrupt(server *http.Server) {
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt)
	<-stopSignal
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}
}

func main() {
	router := mux.NewRouter()
	repository := setupRepo()
	authService := auth.New(repository)
	defer log.Println("done")
	defer errorFunc()
	defer repository.Close()

	setupRoutes(router, repository, authService)

	server := &http.Server{Addr: ":8080", Handler: router}

	go listenAndServe(server)

	shutdownOnInterrupt(server)
}
