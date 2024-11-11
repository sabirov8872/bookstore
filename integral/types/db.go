package types

type UserDB struct {
	ID       int64  `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Email    string `db:"email"`
	Phone    string `db:"phone"`
	UserRole string `db:"userRole"`
}

type GetUserByUserDB struct {
	ID       int64  `db:"id"`
	Password string `db:"password"`
	UserRole string `db:"userRole"`
}

type BookDB struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Genre  string `db:"genre"`
	Author string `db:"author"`
}
