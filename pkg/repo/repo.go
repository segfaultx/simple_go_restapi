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
	DefaultRepository struct {
	Products []Product
	DB       *sql.DB
}
	Product struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}
)

var readMutex = &sync.Mutex{}
var writeMutex = &sync.Mutex{}

func (r *DefaultRepository) GetProductById(id int) (Product, error) {
	for _, item := range r.AllProducts() {
		if item.Id == id {
			return item, nil
		}
	}
	return Product{}, errors.New("no such item")
}

func (r *DefaultRepository) UpdateProduct(p Product) error {
	writeMutex.Lock()
	defer writeMutex.Unlock()
	_, err := r.DB.Exec("UPDATE products SET name = $1 where products.id = $2", p.Name, p.Id)
	return err
}

func (r *DefaultRepository) AddProduct(p Product) error {
	writeMutex.Lock()
	defer writeMutex.Unlock()
	_, err := r.DB.Exec("INSERT INTO products (name) VALUES ($1)", p.Name)
	if err != nil {
		return err
	}
	go r.loadAllProducts()
	return nil
}

func (r *DefaultRepository) loadAllProducts() {
	readMutex.Lock()
	defer readMutex.Unlock()
	defer func() {
		if rec := recover(); rec != nil {
			log.Fatal(rec)
		}
	}()
	rows, err := r.DB.Query("SELECT * from products")
	if err != nil {
		panic(err)
	}
	r.Products = make([]Product, 0)
	for rows.Next() {
		prod := Product{}
		err = rows.Scan(&prod.Id, &prod.Name)
		if err != nil {
			panic(err)
		}
		r.Products = append(r.Products, prod)
	}
}

func (r *DefaultRepository) AllProducts() []Product {
	if r.Products == nil {
		r.loadAllProducts()
	}
	return r.Products
}

func (r *DefaultRepository) RemoveProduct(p Product) error {
	writeMutex.Lock()
	defer writeMutex.Unlock()
	_, err := r.DB.Exec("DELETE FROM products WHERE id=$1", p.Id)
	if err != nil {
		return errors.New(err.Error())
	}
	go r.loadAllProducts()
	return nil
}

func (r *DefaultRepository) InitRepo(user, passwd, dbname string) error {
	dataSourceString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, passwd, dbname)
	db, err := sql.Open("postgres", dataSourceString)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	r.DB = db
	return err
}

func (r *DefaultRepository) Close() {
	err := r.DB.Close()
	log.Fatal(err)
}
