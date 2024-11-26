package types

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
	ID       int    `db:"id"`
	BookName string `db:"bookname"`
	Author   string `db:"author"`
	Genre    string `db:"genre"`
	ISBN     string `db:"isbn"`
	Filename string `db:"filename"`
}

type AuthorDB struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type GenreDB struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}
