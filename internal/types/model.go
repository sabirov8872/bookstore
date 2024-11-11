package types

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	UserRole string `json:"userRole"`
}

type GetUserByUserResponse struct {
	UserID int64  `json:"userId"`
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
	UserRole string `json:"userRole"`
}

type ListUserResponse struct {
	Items []*User `json:"items"`
}

type CreateUserResponse struct {
	ID int64 `json:"Id"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type Book struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Genre  string `json:"genre"`
	Author string `json:"author"`
}

type ListBookResponse struct {
	Items []*Book `json:"items"`
}

type CreateBookRequest struct {
	Name   string `json:"name"`
	Genre  string `json:"genre"`
	Author string `json:"author"`
}

type UpdateBookRequest struct {
	Name   string `json:"name"`
	Genre  string `json:"genre"`
	Author string `json:"author"`
}

type CreateBookResponse struct {
	ID int64 `json:"bookId"`
}
