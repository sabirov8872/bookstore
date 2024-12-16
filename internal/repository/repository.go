package repository

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

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

	GetAllBooks(req types.GetAllBooksRequest) ([]*types.BookDB, error)
	GetBookByID(id int) (*types.BookDB, error)
	CreateBook(req types.CreateBookRequest) (int, error)
	UpdateBook(id int, req types.UpdateBookRequest) error
	DeleteBook(id int) (string, error)

	GetAllAuthors() ([]*types.AuthorDB, error)
	GetAuthorById(id int) (*types.AuthorDB, error)
	CreateAuthor(req types.CreateAuthorRequest) (int, error)
	UpdateAuthor(id int, req types.UpdateAuthorRequest) error
	DeleteAuthor(id int) error

	GetAllGenres() ([]*types.GenreDB, error)
	CreateGenre(req types.CreateGenreRequest) (int, error)
	UpdateGenre(id int, req types.UpdateGenreRequest) error
	DeleteGenre(id int) error

	GetFilename(id int) (string, error)
	UpdateFilename(id int, filename string) (string, error)
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

func (repo *Repository) CreateUser(req types.CreateUserRequest) (int, error) {
	var id int
	err := repo.DB.QueryRow(createUserQuery,
		req.Username,
		req.Password,
		req.Email,
		req.Phone,
		"user").
		Scan(&id)
	if err != nil {
		return 0, errors.New("bad request")
	}

	return id, nil
}

func (repo *Repository) GetUserByUsername(username string) (*types.GetUserByUsernameDB, error) {
	var resp types.GetUserByUsernameDB
	err := repo.DB.QueryRow(getUserByUsernameQuery, username).Scan(
		&resp.ID,
		&resp.Password,
		&resp.Role)
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
		err = rows.Scan(
			&u.ID,
			&u.Username,
			&u.Password,
			&u.Email,
			&u.Phone,
			&u.Role)
		if err != nil {
			return nil, err
		}

		resp = append(resp, &u)
	}

	return resp, nil
}

func (repo *Repository) GetUserByID(id int) (*types.UserDB, error) {
	var resp types.UserDB
	err := repo.DB.QueryRow(getUserByIdQuery, id).Scan(
		&resp.ID,
		&resp.Username,
		&resp.Password,
		&resp.Email,
		&resp.Phone,
		&resp.Role)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (repo *Repository) UpdateUser(id int, req types.UpdateUserRequest) error {
	_, err := repo.DB.Query(updateUserQuery,
		req.Username,
		req.Password,
		req.Email,
		req.Phone,
		id)
	if err != nil {
		return errors.New("bad request")
	}

	return err
}

func (repo *Repository) UpdateUserById(id int, req types.UpdateUserByIdRequest) error {
	_, err := repo.DB.Query(updateUserByIdQuery,
		req.Username,
		req.Password,
		req.Email,
		req.Phone,
		req.Role,
		id)
	if err != nil {
		return errors.New("bad request")
	}

	return err
}

func (repo *Repository) DeleteUser(id int) error {
	_, err := repo.DB.Query(deleteUserQuery, id)
	return err
}

func (repo *Repository) GetAllBooks(req types.GetAllBooksRequest) ([]*types.BookDB, error) {
	query := getAllBooksQuery

	if req.Filter == "author_id" {
		_, err := strconv.Atoi(req.ID)
		if err != nil {
			return nil, errors.New("bad author id")
		}

		query += "\nWHERE b.author_id = " + req.ID
	} else if req.Filter == "genre_id" {
		_, err := strconv.Atoi(req.ID)
		if err != nil {
			return nil, errors.New("bad genre id")
		}

		query += "\nWHERE b.genre_id = " + req.ID
	}

	checkSortBy := false
	if req.SortBy == "title" {
		query += "\nORDER BY b.title"
		checkSortBy = true
	} else if req.SortBy == "created_at" {
		query += "\nORDER BY b.created_at"
		checkSortBy = true
	} else if req.SortBy == "updated_at" {
		query += "\nORDER BY b.updated_at"
		checkSortBy = true
	}

	if checkSortBy && req.OrderBy == "desc" {
		query += " DESC"
	} else if checkSortBy && req.OrderBy == "asc" {
		query += " ASC"
	}

	rows, err := repo.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resp []*types.BookDB
	for rows.Next() {
		var b types.BookDB
		err = rows.Scan(
			&b.ID,
			&b.Title,
			&b.Author.ID,
			&b.Author.Name,
			&b.Genre.ID,
			&b.Genre.Name,
			&b.ISBN,
			&b.Filename,
			&b.Description,
			&b.CreatedAt,
			&b.UpdatedAt)
		if err != nil {
			return nil, err
		}

		resp = append(resp, &b)
	}

	return resp, nil
}

func (repo *Repository) GetBookByID(id int) (*types.BookDB, error) {
	var res types.BookDB
	err := repo.DB.QueryRow(getBookByIdQuery, id).Scan(
		&res.ID,
		&res.Title,
		&res.Author.ID,
		&res.Author.Name,
		&res.Genre.ID,
		&res.Genre.Name,
		&res.ISBN,
		&res.Filename,
		&res.Description,
		&res.CreatedAt,
		&res.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repo *Repository) CreateBook(req types.CreateBookRequest) (int, error) {
	var id int
	err := repo.DB.QueryRow(createBookQuery,
		req.AuthorId,
		req.GenreId,
		req.Title,
		req.ISBN,
		"",
		req.Description,
		time.Now(),
		time.Now()).
		Scan(&id)
	if err != nil {
		return 0, errors.New("bad request")
	}

	return id, nil
}

func (repo *Repository) UpdateBook(id int, req types.UpdateBookRequest) error {
	_, err := repo.DB.Query(updateBookQuery,
		req.AuthorId,
		req.GenreId,
		req.Title,
		req.ISBN,
		req.Description,
		time.Now(),
		id)
	if err != nil {
		return errors.New("bad request")
	}

	return err
}

func (repo *Repository) DeleteBook(id int) (string, error) {
	var filename string
	err := repo.DB.QueryRow(getFilenameQuery, id).Scan(&filename)
	if err != nil {
		return "", err
	}

	_, err = repo.DB.Query(deleteBookQuery, id)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (repo *Repository) UpdateFilename(id int, filename string) (string, error) {
	var oldFilename string
	err := repo.DB.QueryRow(getFilenameQuery, id).Scan(&oldFilename)
	if err != nil {
		return "", err
	}

	_, err = repo.DB.Query(updateFilenameQuery, filename, id)
	if err != nil {
		return "", err
	}

	return oldFilename, nil
}

func (repo *Repository) GetFilename(id int) (string, error) {
	var filename string
	err := repo.DB.QueryRow(getFilenameQuery, id).Scan(&filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (repo *Repository) GetAllAuthors() ([]*types.AuthorDB, error) {
	rows, err := repo.DB.Query(getAllAuthorsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []*types.AuthorDB
	for rows.Next() {
		var author types.AuthorDB
		err = rows.Scan(&author.ID, &author.Name)
		if err != nil {
			return nil, err
		}

		authors = append(authors, &author)
	}

	return authors, nil
}

func (repo *Repository) GetAuthorById(id int) (*types.AuthorDB, error) {
	var res types.AuthorDB
	err := repo.DB.QueryRow(`select id, name from authors where id = $1`, id).Scan(
		&res.ID,
		&res.Name)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repo *Repository) CreateAuthor(req types.CreateAuthorRequest) (int, error) {
	var id int
	err := repo.DB.QueryRow(createAuthorQuery, req.Name).Scan(&id)
	if err != nil {
		return 0, errors.New("bad request")
	}

	return id, nil
}

func (repo *Repository) UpdateAuthor(id int, req types.UpdateAuthorRequest) error {
	_, err := repo.DB.Query(updateAuthorQuery, req.Name, id)
	if err != nil {
		return errors.New("bad request")
	}

	return nil
}

func (repo *Repository) DeleteAuthor(id int) error {
	row, err := repo.DB.Query(`select id from books where author_id = $1`, id)
	if err != nil {
		return err
	}

	var ids []int
	for row.Next() {
		var i int
		err = row.Scan(&i)
		if err != nil {
			return err
		}

		ids = append(ids, i)

		if len(ids) > 0 {
			break
		}
	}

	if len(ids) == 0 {
		_, err = repo.DB.Query(deleteAuthorQuery, id)
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("cannot delete author")
}

func (repo *Repository) GetAllGenres() ([]*types.GenreDB, error) {
	rows, err := repo.DB.Query(getAllGenresQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []*types.GenreDB
	for rows.Next() {
		var genre types.GenreDB
		err = rows.Scan(&genre.ID, &genre.Name)
		if err != nil {
			return nil, err
		}

		genres = append(genres, &genre)
	}

	return genres, nil
}

func (repo *Repository) CreateGenre(req types.CreateGenreRequest) (int, error) {
	var id int
	err := repo.DB.QueryRow(createGenreQuery, req.Name).Scan(&id)
	if err != nil {
		return 0, errors.New("bad request")
	}

	return id, nil
}

func (repo *Repository) UpdateGenre(id int, req types.UpdateGenreRequest) error {
	_, err := repo.DB.Query(updateGenreQuery, req.Name, id)
	if err != nil {
		return errors.New("bad request")
	}

	return nil
}

func (repo *Repository) DeleteGenre(id int) error {
	rows, err := repo.DB.Query(`select id from books where genre_id = $1`, id)
	if err != nil {
		return err
	}

	var ids []int
	for rows.Next() {
		var i int
		err = rows.Scan(&i)
		if err != nil {
			return err
		}

		ids = append(ids, i)

		if len(ids) > 0 {
			break
		}
	}

	if len(ids) == 0 {
		_, err = repo.DB.Query(deleteGenreQuery, id)
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("cannot delete genre")
}
