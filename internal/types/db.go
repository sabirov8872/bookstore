package types

import "time"

type UserDB struct {
	ID       int    `postgres:"id"`
	Username string `postgres:"username"`
	Password string `postgres:"password"`
	Email    string `postgres:"email"`
	Phone    string `postgres:"phone"`
	Role     string `postgres:"role"`
}

type GetUserByUsernameDB struct {
	ID       int    `postgres:"id"`
	Password string `postgres:"password"`
	Role     string `postgres:"role"`
}

type BookDB struct {
	ID          int       `postgres:"id"`
	Title       string    `postgres:"title"`
	Author      AuthorDB  `postgres:"author"`
	Genre       GenreDB   `postgres:"genre"`
	ISBN        string    `postgres:"isbn"`
	Filename    string    `postgres:"filename"`
	Description string    `postgres:"description"`
	CreatedAt   time.Time `postgres:"createdAt"`
	UpdatedAt   time.Time `postgres:"updatedAt"`
}

type AuthorDB struct {
	ID   int    `postgres:"id"`
	Name string `postgres:"name"`
}

type GenreDB struct {
	ID   int    `postgres:"id"`
	Name string `postgres:"name"`
}
