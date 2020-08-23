package repo

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"sync"
)

type (
	DefaultRepository struct {
		Products []Product
		Users    []User
		DB       *sql.DB
	}
)

var readMutex = &sync.Mutex{}
var writeMutex = &sync.Mutex{}

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
