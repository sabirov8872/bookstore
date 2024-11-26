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
	Items []*User `json:"items"`
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
	BookName string `json:"bookName"`
	Genre    string `json:"genre"`
	Author   string `json:"author"`
	ISBN     string `json:"isbn"`
}

type UpdateBookRequest struct {
	BookName string `json:"bookName"`
	Genre    string `json:"genre"`
	Author   string `json:"author"`
	ISBN     string `json:"isbn"`
}

type CreateBookResponse struct {
	ID int `json:"bookId"`
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
