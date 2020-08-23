package repo

import (
	"errors"
	"log"
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

	Product struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}
)

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
