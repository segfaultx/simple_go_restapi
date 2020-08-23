package repo

import "errors"

type (
	UserRepository interface {
		AddUser(u User) error
		GetByUsername(username string) (User, error)
	}

	User struct {
		Id       int    `json:"id"`
		Username string `json:"username"`
		Password string `json:"password"`
		Role     Role `json:"role"`
	}

	Role string
)

const (
	ADMIN Role = "ADMIN"
	USER  Role = "USER"
)

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
	_, err := repo.DB.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)", u.Username, u.Password, u.Role)
	if err != nil {
		return err
	}
	go repo.loadAllUsers()
	return nil
}

func (repo *DefaultRepository) GetByUsername(username string) (User, error) {
	if repo.Users == nil {
		repo.loadAllUsers()
	}
	for _, usr := range repo.Users {
		if usr.Username == username {
			return usr, nil
		}
	}
	return User{}, errors.New("user not found")
}
