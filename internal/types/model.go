package types

import (
	"github.com/minio/minio-go/v7"
	"mime/multipart"
	"time"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

type GetUserByUserResponse struct {
	UserID int    `json:"userId"`
	Token  string `json:"token"`
}

type GetUserByUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type UpdateUserByIdRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

type ListUserResponse struct {
	UsersCount int     `json:"usersCount"`
	Items      []*User `json:"items"`
}

type CreateUserResponse struct {
	ID int `json:"userId"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type Author struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Book struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Author      Author    `json:"author"`
	Genre       Genre     `json:"genre"`
	ISBN        string    `json:"isbn"`
	Filename    string    `json:"filename"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ListBookResponse struct {
	BooksCount int     `json:"booksCount"`
	Items      []*Book `json:"items"`
}

type CreateBookRequest struct {
	AuthorId    int    `json:"authorId"`
	GenreId     int    `json:"genreId"`
	Title       string `json:"title"`
	ISBN        string `json:"isbn"`
	Description string `json:"description"`
}

type CreateBookResponse struct {
	ID int `json:"bookId"`
}

type UpdateBookRequest struct {
	AuthorId    int    `json:"authorId"`
	GenreId     int    `json:"genreId"`
	Title       string `json:"title"`
	ISBN        string `json:"isbn"`
	Description string `json:"description"`
}

type ListAuthorResponse struct {
	AuthorsCount int       `json:"authorsCount"`
	Items        []*Author `json:"items"`
}

type ListGenreResponse struct {
	GenresCount int      `json:"genresCount"`
	Items       []*Genre `json:"items"`
}

type CreateAuthorRequest struct {
	Name string `json:"name"`
}

type CreateAuthorResponse struct {
	ID int `json:"authorId"`
}

type CreateGenreResponse struct {
	ID int `json:"genreId"`
}

type UpdateAuthorRequest struct {
	Name string `json:"name"`
}

type CreateGenreRequest struct {
	Name string `json:"name"`
}

type UpdateGenreRequest struct {
	Name string `json:"name"`
}

type GetAllBooksRequest struct {
	Filter  string
	ID      string
	SortBy  string
	OrderBy string
}

type UploadFileByBookIdRequest struct {
	ID         int
	FileHeader *multipart.FileHeader
	File       multipart.File
}

type GetFileByBookIdResponse struct {
	Filename string
	File     *minio.Object
}
