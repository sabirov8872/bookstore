package repository

import _ "embed"

var (

	//go:embed queries/get_user_by_username.sql
	getUserByUsernameQuery string

	//go:embed queries/get_all_users.sql
	getAllUsersQuery string

	//go:embed queries/get_user_by_id.sql
	getUserByIdQuery string

	//go:embed queries/create_user.sql
	createUserQuery string

	//go:embed queries/update_user.sql
	updateUserQuery string

	//go:embed queries/delete_user.sql
	deleteUserQuery string

	//go:embed queries/get_all_books.sql
	getAllBooksQuery string

	//go:embed queries/get_book_by_id.sql
	getBookByIdQuery string

	//go:embed queries/create_book.sql
	createBookQuery string

	//go:embed queries/update_book.sql
	updateBookQuery string

	//go:embed queries/delete_book.sql
	deleteBookQuery string

	//go:embed queries/update_user_by_id.sql
	updateUserByIdQuery string
)
