package types

import "time"

type UserDB struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Email    string `db:"email"`
	Phone    string `db:"phone"`
	UserRole string `db:"userrole"`
}

type GetUserByUsernameDB struct {
	ID       int    `db:"id"`
	Password string `db:"password"`
	UserRole string `db:"userrole"`
}

type BookDB struct {
	ID          int       `db:"id"`
	Name        string    `db:"name"`
	Author      AuthorDB  `db:"author"`
	Genre       GenreDB   `db:"genre"`
	ISBN        string    `db:"isbn"`
	Filename    string    `db:"filename"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"createdAt"`
	UpdatedAt   time.Time `db:"updatedAt"`
}

type AuthorDB struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type GenreDB struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}
