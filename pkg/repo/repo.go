package repo

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"sync"
)

type (
	ProductRepository interface {
		AddProduct(p Product) error
		RemoveProduct(p Product) error
		UpdateProduct(p Product) error
		AllProducts() []Product
		GetProductById(id int) (Product, error)
		InitRepo(user, passwd, dbname string) error
		Close()
	}

	UserRepository interface {
		AddUser(u User) error
		GetByUsername(username string) (string, error)
	}

	DefaultRepository struct {
		Products []Product
		Users    []User
		DB       *sql.DB
	}
	Product struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}

	User struct {
		Id       int    `json:"id"`
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
)

var readMutex = &sync.Mutex{}
var writeMutex = &sync.Mutex{}

// Products Methods

func (repo *DefaultRepository) GetProductById(id int) (Product, error) {
	for _, item := range repo.AllProducts() {
		if item.Id == id {
			return item, nil
		}
	}
	return Product{}, errors.New("no such item")
}

func (repo *DefaultRepository) UpdateProduct(p Product) error {
	writeMutex.Lock()
	defer writeMutex.Unlock()
	_, err := repo.DB.Exec("UPDATE products SET name = $1 where products.id = $2", p.Name, p.Id)
	return err
}

func (repo *DefaultRepository) AddProduct(p Product) error {
	writeMutex.Lock()
	defer writeMutex.Unlock()
	_, err := repo.DB.Exec("INSERT INTO products (name) VALUES ($1)", p.Name)
	if err != nil {
		return err
	}
	go repo.loadAllProducts()
	return nil
}

func (repo *DefaultRepository) loadAllProducts() {
	readMutex.Lock()
	defer readMutex.Unlock()
	defer func() {
		if rec := recover(); rec != nil {
			log.Fatal(rec)
		}
	}()
	rows, err := repo.DB.Query("SELECT * from products")
	if err != nil {
		panic(err)
	}
	repo.Products = make([]Product, 0)
	for rows.Next() {
		prod := Product{}
		err = rows.Scan(&prod.Id, &prod.Name)
		if err != nil {
			panic(err)
		}
		repo.Products = append(repo.Products, prod)
	}
}

func (repo *DefaultRepository) AllProducts() []Product {
	if repo.Products == nil {
		repo.loadAllProducts()
	}
	return repo.Products
}

func (repo *DefaultRepository) RemoveProduct(p Product) error {
	writeMutex.Lock()
	defer writeMutex.Unlock()
	_, err := repo.DB.Exec("DELETE FROM products WHERE id=$1", p.Id)
	if err != nil {
		return errors.New(err.Error())
	}
	go repo.loadAllProducts()
	return nil
}

// User methods

func (repo *DefaultRepository) loadAllUsers() {
	readMutex.Lock()
	defer readMutex.Unlock()
	repo.Users = make([]User, 0)

	rows, err := repo.DB.Query("SELECT * from users")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		user := User{}
		err = rows.Scan(&user.Id, &user.Username, &user.Password, &user.Role)
		if err != nil {
			panic(err)
		}
		repo.Users = append(repo.Users, user)
	}
}

func (repo *DefaultRepository) AddUser(u User) error {
	writeMutex.Lock()
	defer writeMutex.Unlock()
	return nil
}

func (repo *DefaultRepository) GetByUsername(username string) (string, error) {
	if repo.Users == nil {
		repo.loadAllUsers()
	}
	return "", nil
}

// Init and Close repo methods

func (repo *DefaultRepository) InitRepo(user, passwd, dbname string) error {
	dataSourceString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, passwd, dbname)
	db, err := sql.Open("postgres", dataSourceString)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	repo.DB = db
	return err
}

func (repo *DefaultRepository) Close() {
	err := repo.DB.Close()
	if err != nil {
		panic(err)
	}
}
