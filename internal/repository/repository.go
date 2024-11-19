package repository

import (
	"database/sql"

	"github.com/sabirov8872/bookstore/internal/types"
)

type Repository struct {
	DB *sql.DB
}

type IRepository interface {
	CreateUser(req types.CreateUserRequest) (int, error)
	GetUserByUsername(username string) (*types.GetUserByUsernameDB, error)

	GetAllUsers() (resp []*types.UserDB, err error)
	GetUserByID(id int) (*types.UserDB, error)
	UpdateUser(id int, req types.UpdateUserRequest) error
	UpdateUserById(id int, req types.UpdateUserByIdRequest) error
	DeleteUser(id int) error

	GetAllBooks() ([]*types.BookDB, error)
	GetBookByID(id int) (*types.BookDB, error)
	CreateBook(req types.CreateBookRequest) (int, error)
	UpdateBook(id int, req types.UpdateBookRequest) error
	DeleteBook(id int) error
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

func (repo *Repository) CreateUser(req types.CreateUserRequest) (int, error) {
	var id int
	err := repo.DB.QueryRow(createUserQuery, req.Username, req.Password, req.Email, req.Phone, "user").Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *Repository) GetUserByUsername(username string) (*types.GetUserByUsernameDB, error) {
	var resp types.GetUserByUsernameDB
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

func (repo *Repository) GetUserByID(id int) (*types.UserDB, error) {
	var resp types.UserDB
	err := repo.DB.QueryRow(getUserByIdQuery, id).Scan(&resp.ID, &resp.Username, &resp.Password, &resp.Email, &resp.Phone, &resp.UserRole)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (repo *Repository) UpdateUser(id int, req types.UpdateUserRequest) error {
	_, err := repo.DB.Query(updateUserQuery, req.Username, req.Password, req.Email, req.Phone, id)
	return err
}

func (repo *Repository) UpdateUserById(id int, req types.UpdateUserByIdRequest) error {
	_, err := repo.DB.Query(updateUserByIdQuery, req.Username, req.Password, req.Email, req.Phone, req.UserRole, id)
	return err
}

func (repo *Repository) DeleteUser(id int) error {
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

func (repo *Repository) GetBookByID(id int) (*types.BookDB, error) {
	var b types.BookDB
	err := repo.DB.QueryRow(getBookByIdQuery, id).Scan(&b.ID, &b.Name, &b.Genre, &b.Author)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func (repo *Repository) CreateBook(req types.CreateBookRequest) (int, error) {
	var id int
	err := repo.DB.QueryRow(createBookQuery, req.Name, req.Genre, req.Author).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *Repository) UpdateBook(id int, req types.UpdateBookRequest) error {
	_, err := repo.DB.Query(updateBookQuery, req.Name, req.Genre, req.Author, id)
	return err
}

func (repo *Repository) DeleteBook(id int) error {
	_, err := repo.DB.Query(deleteBookQuery, id)
	return err
}
