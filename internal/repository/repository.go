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
	GetFilename(id int) (string, error)
	UpdateFilename(id int, filename string) error
	GetAuthors() ([]*types.AuthorDB, int, error)
	GetGenres() ([]*types.GenreDB, int, error)
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
		return 0, err
	}

	return id, nil
}

func (repo *Repository) GetUserByUsername(username string) (*types.GetUserByUsernameDB, error) {
	var resp types.GetUserByUsernameDB
	err := repo.DB.QueryRow(getUserByUsernameQuery, username).Scan(
		&resp.ID,
		&resp.Password,
		&resp.UserRole)
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
			&u.UserRole)
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
		&resp.UserRole)
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
	return err
}

func (repo *Repository) UpdateUserById(id int, req types.UpdateUserByIdRequest) error {
	_, err := repo.DB.Query(updateUserByIdQuery,
		req.Username,
		req.Password,
		req.Email,
		req.Phone,
		req.UserRole,
		id)
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
		err = rows.Scan(
			&b.ID,
			&b.BookName,
			&b.Author,
			&b.Genre,
			&b.ISBN,
			&b.Filename)
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
		&res.BookName,
		&res.Author,
		&res.Genre,
		&res.ISBN,
		&res.Filename)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repo *Repository) CreateBook(req types.CreateBookRequest) (int, error) {
	var authorId int
	err := repo.DB.QueryRow(
		"select id from authors where author = $1", req.Author).Scan(&authorId)
	if err != nil {
		err = repo.DB.QueryRow(
			"INSERT INTO authors (author) VALUES ($1) RETURNING id",
			req.Author).Scan(&authorId)
		if err != nil {
			return 0, err
		}
	}

	var genreId int
	err = repo.DB.QueryRow(
		"select id from genres where genre = $1", req.Genre).Scan(&genreId)
	if err != nil {
		err = repo.DB.QueryRow(
			"insert into genres (genre) values ($1) RETURNING id",
			req.Genre).Scan(&genreId)
		if err != nil {
			return 0, err
		}
	}

	var bookId int
	err = repo.DB.QueryRow(createBookQuery,
		authorId,
		genreId,
		req.BookName,
		req.ISBN,
		"no file").
		Scan(&bookId)
	if err != nil {
		return 0, err
	}

	return bookId, nil
}

func (repo *Repository) UpdateBook(id int, req types.UpdateBookRequest) error {
	var authorId, genreId int
	err := repo.DB.QueryRow(
		"select author_id, genre_id from books where id = $1",
		id).Scan(&authorId, &genreId)
	if err != nil {
		return err
	}

	rowsByAuthorId, err := repo.DB.Query(
		"select id from books where author_id = $1", authorId)
	if err != nil {
		return err
	}
	defer rowsByAuthorId.Close()

	var bookIdsByAuthorId []int
	for rowsByAuthorId.Next() {
		var i int
		err = rowsByAuthorId.Scan(&i)
		if err != nil {
			return err
		}

		bookIdsByAuthorId = append(bookIdsByAuthorId, i)

		if len(bookIdsByAuthorId) == 2 {
			break
		}
	}

	if len(bookIdsByAuthorId) < 2 {
		_, err = repo.DB.Query(
			"UPDATE authors SET author = $1 WHERE id = $2",
			req.Author, authorId)
		if err != nil {
			return err
		}
	} else {
		err = repo.DB.QueryRow(
			"insert into authors (author) values ($1) RETURNING id",
			req.Author).Scan(&authorId)
		if err != nil {
			return err
		}
	}

	rowsByGenreId, err := repo.DB.Query(
		"select id from books where genre_id = $1", genreId)
	if err != nil {
		return err
	}
	defer rowsByGenreId.Close()

	var bookIdsByGenreId []int
	for rowsByGenreId.Next() {
		var i int
		err = rowsByGenreId.Scan(&i)
		if err != nil {
			return err
		}

		bookIdsByGenreId = append(bookIdsByGenreId, i)

		if len(bookIdsByGenreId) == 2 {
			break
		}
	}

	if len(bookIdsByGenreId) < 2 {
		_, err = repo.DB.Query(
			"UPDATE genres SET genre = $1 WHERE id = $2",
			req.Genre, genreId)
		if err != nil {
			return err
		}
	} else {
		err = repo.DB.QueryRow(
			"insert into genres (genre) values ($1) RETURNING id",
			req.Genre).Scan(&genreId)
		if err != nil {
			return err
		}
	}

	_, err = repo.DB.Query(updateBookQuery,
		authorId,
		genreId,
		req.BookName,
		req.ISBN,
		id)

	return err
}

func (repo *Repository) DeleteBook(id int) error {
	var authorId, genreId int
	err := repo.DB.QueryRow(
		"select author_id, genre_id from books where id = $1",
		id).Scan(&authorId, &genreId)
	if err != nil {
		return err
	}

	rowsByAuthorId, err := repo.DB.Query(
		"select id from books where author_id = $1", authorId)
	if err != nil {
		return err
	}
	defer rowsByAuthorId.Close()

	var bookIdsByAuthorId []int
	for rowsByAuthorId.Next() {
		var i int
		err = rowsByAuthorId.Scan(&i)
		if err != nil {
			return err
		}

		bookIdsByAuthorId = append(bookIdsByAuthorId, i)

		if len(bookIdsByAuthorId) == 2 {
			break
		}
	}

	rowsByGenreId, err := repo.DB.Query(
		"select id from books where genre_id = $1", genreId)
	if err != nil {
		return err
	}
	defer rowsByGenreId.Close()

	var bookIdsByGenreId []int
	for rowsByGenreId.Next() {
		var i int
		err = rowsByGenreId.Scan(&i)
		if err != nil {
			return err
		}

		bookIdsByGenreId = append(bookIdsByGenreId, i)

		if len(bookIdsByGenreId) == 2 {
			break
		}
	}

	_, err = repo.DB.Query(deleteBookQuery, id)
	if err != nil {
		return err
	}

	if len(bookIdsByAuthorId) < 2 {
		_, err = repo.DB.Query(
			"delete from authors where id = $1", authorId)
		if err != nil {
			return err
		}
	}

	if len(bookIdsByGenreId) < 2 {
		_, err = repo.DB.Query(
			"delete from genres where id = $1", genreId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *Repository) UpdateFilename(id int, filename string) error {
	_, err := repo.DB.Query(updateFilenameQuery, filename, id)
	return err
}

func (repo *Repository) GetFilename(id int) (string, error) {
	var filename string
	err := repo.DB.QueryRow(getFilenameQuery, id).Scan(&filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (repo *Repository) GetAuthors() ([]*types.AuthorDB, int, error) {
	rows, err := repo.DB.Query(getAuthorsQuery)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var authors []*types.AuthorDB
	totalAuthors := 0
	for rows.Next() {
		var author types.AuthorDB
		err = rows.Scan(&author.ID, &author.Name)
		if err != nil {
			return nil, 0, err
		}

		authors = append(authors, &author)
		totalAuthors += 1
	}

	return authors, totalAuthors, nil
}

func (repo *Repository) GetGenres() ([]*types.GenreDB, int, error) {
	rows, err := repo.DB.Query(getGenresQuery)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var genres []*types.GenreDB
	totalGenres := 0
	for rows.Next() {
		var genre types.GenreDB
		err = rows.Scan(&genre.ID, &genre.Name)
		if err != nil {
			return nil, 0, err
		}

		genres = append(genres, &genre)
		totalGenres += 1
	}

	return genres, totalGenres, nil
}
