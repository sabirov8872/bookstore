package types

import "time"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	UserRole string `json:"userRole"`
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
	UserRole string `json:"userRole"`
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

type Book struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
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
	Name        string `json:"name"`
	ISBN        string `json:"isbn"`
	Description string `json:"description"`
}

type CreateBookResponse struct {
	ID int `json:"bookId"`
}

type UpdateBookRequest struct {
	AuthorId    int    `json:"authorId"`
	GenreId     int    `json:"genreId"`
	Name        string `json:"name"`
	ISBN        string `json:"isbn"`
	Description string `json:"description"`
}

type Author struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ListAuthorResponse struct {
	AuthorsCount int       `json:"authorsCount"`
	Items        []*Author `json:"items"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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

type UpdateAuthorRequest struct {
	Name string `json:"name"`
}

type CreateGenreResponse struct {
	ID int `json:"genreId"`
}

type CreateGenreRequest struct {
	Name string `json:"name"`
}

type UpdateGenreRequest struct {
	Name string `json:"name"`
}
