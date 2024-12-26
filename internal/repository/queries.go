package repository

import _ "embed"

var (
	//users
	//go:embed queries/create_user.sql
	createUserQuery string

	//go:embed queries/getSessionIdByUsername.sql
	getSessionIdByUsernameQuery string

	//go:embed queries/get_all_users.sql
	getAllUsersQuery string

	//go:embed queries/get_user_by_id.sql
	getUserByIdQuery string

	//go:embed queries/updateUserBySessionId.sql
	updateUserBySessionIdQuery string

	//go:embed queries/delete_user.sql
	deleteUserQuery string

	//go:embed queries/update_user_by_id.sql
	updateUserByIdQuery string

	//books
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

	//authors
	//go:embed queries/get_all_authors.sql
	getAllAuthorsQuery string

	//go:embed queries/create_author.sql
	createAuthorQuery string

	//go:embed queries/update_author.sql
	updateAuthorQuery string

	//go:embed queries/delete_author.sql
	deleteAuthorQuery string

	//genres
	//go:embed queries/get_all_genres.sql
	getAllGenresQuery string

	//go:embed queries/create_genre.sql
	createGenreQuery string

	//go:embed queries/update_genre.sql
	updateGenreQuery string

	//go:embed queries/delete_genre.sql
	deleteGenreQuery string

	//files
	//go:embed queries/get_filename.sql
	getFilenameQuery string

	//go:embed queries/update_filename.sql
	updateFilenameQuery string
)
