package repository

import (
	"database/sql"

	"github.com/sabirov8872/bookstore/integral/types"
)

type Repository struct {
	DB *sql.DB
}

type IRepository interface {
	CreateUser(req types.CreateUserRequest) (int64, error)
	GetUserByUsername(username string) (*types.GetUserByUserDB, error)

	GetAllUsers() (resp []*types.UserDB, err error)
	GetUserByID(id string) (*types.UserDB, error)
	UpdateUser(id string, req types.UpdateUserRequest) error
	DeleteUser(id string) error

	GetAllBooks() ([]*types.BookDB, error)
	GetBookByID(id string) (*types.BookDB, error)
	CreateBook(req types.CreateBookRequest) (int64, error)
	UpdateBook(id string, req types.UpdateBookRequest) error
	DeleteBook(id string) error
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

func (repo *Repository) CreateUser(req types.CreateUserRequest) (int64, error) {
	var id int64
	err := repo.DB.QueryRow(createUserQuery, req.Username, req.Password, req.Email, req.Phone, "user").Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *Repository) GetUserByUsername(username string) (*types.GetUserByUserDB, error) {
	var resp types.GetUserByUserDB
	err := repo.DB.QueryRow(getUserByUsernameQuery, username).Scan(&resp.ID, &resp.Password, &resp.UserRole)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (repo *Repository) GetAllUsers() (resp []*types.UserDB, err error) {
	rows, err := repo.DB.Query(getAllUsersQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u types.UserDB
		err = rows.Scan(&u.ID, &u.Username, &u.Password, &u.Email, &u.Phone, &u.UserRole)
		if err != nil {
			return nil, err
		}

		resp = append(resp, &u)
	}

	return resp, nil
}

func (repo *Repository) GetUserByID(id string) (*types.UserDB, error) {
	var resp types.UserDB
	err := repo.DB.QueryRow(getUserByIdQuery, id).Scan(&resp.ID, &resp.Username, &resp.Password, &resp.Email, &resp.Phone, &resp.UserRole)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (repo *Repository) UpdateUser(id string, req types.UpdateUserRequest) error {
	_, err := repo.DB.Query(updateUserQuery, req.Username, req.Password, req.Email, req.Phone, req.UserRole, id)
	return err
}

func (repo *Repository) DeleteUser(id string) error {
	_, err := repo.DB.Query(deleteUserQuery, id)
	return err
}

func (repo *Repository) GetAllBooks() ([]*types.BookDB, error) {
	rows, err := repo.DB.Query(getAllBooksQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resp []*types.BookDB
	for rows.Next() {
		var b types.BookDB
		err = rows.Scan(&b.ID, &b.Name, &b.Genre, &b.Author)
		if err != nil {
			return nil, err
		}

		resp = append(resp, &b)
	}

	return resp, nil
}

func (repo *Repository) GetBookByID(id string) (*types.BookDB, error) {
	var b types.BookDB
	err := repo.DB.QueryRow(getBookByIdQuery, id).Scan(&b.ID, &b.Name, &b.Genre, &b.Author)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func (repo *Repository) CreateBook(req types.CreateBookRequest) (int64, error) {
	var id int64
	err := repo.DB.QueryRow(createBookQuery, req.Name, req.Genre, req.Author).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *Repository) UpdateBook(id string, req types.UpdateBookRequest) error {
	_, err := repo.DB.Query(updateBookQuery, req.Name, req.Genre, req.Author, id)
	return err
}

func (repo *Repository) DeleteBook(id string) error {
	_, err := repo.DB.Query(deleteBookQuery, id)
	return err
}
