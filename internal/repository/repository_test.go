package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/sabirov8872/bookstore/internal/types"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainer struct {
	container testcontainers.Container
	ctx       context.Context
}

func newTestContainer(t *testing.T) *TestContainer {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	return &TestContainer{
		container: container,
		ctx:       ctx,
	}
}

func (c *TestContainer) getDB(t *testing.T) *sql.DB {
	Host, err := c.container.Host(c.ctx)
	require.NoError(t, err)
	Port, err := c.container.MappedPort(c.ctx, "5432")
	require.NoError(t, err)

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		"testuser", "testpass", Host, Port.Port(), "testdb")

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	m, err := migrate.New(
		"file:///Users/Pro/go/src/bookstore/migrations",
		connStr,
	)
	require.NoError(t, err)

	err = m.Migrate(2)
	require.NoError(t, err)

	return db
}

func (c *TestContainer) terminate(t *testing.T) {
	require.NoError(t, c.container.Terminate(c.ctx))
}

func TestRepository_CreateUser(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()

	tests := map[string]struct {
		req types.CreateUserRequest
		res int
		err error
	}{
		"case 01: success": {
			req: types.CreateUserRequest{
				Username: "testuser",
				Password: "testpass",
			},
			res: 1,
			err: nil,
		},
		"case 02: fail": {
			req: types.CreateUserRequest{
				Username: "testuser",
				Password: "testpass",
			},
			res: 0,
			err: errors.New("bad request"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := NewRepository(db)
			res, err := repo.CreateUser(tt.req)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestRepository_GetUserByUsername(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateUser(types.CreateUserRequest{
		Username: "foo",
		Password: "bar"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		req string
		res *types.GetUserByUsernameDB
		err error
	}{
		"case 01: bad request": {
			req: "",
			res: nil,
			err: sql.ErrNoRows,
		},
		"case 02: success": {
			req: "foo",
			res: &types.GetUserByUsernameDB{
				ID:       1,
				Password: "bar",
				Role:     "user",
			},
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := repo.GetUserByUsername(tt.req)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestRepository_GetAllUsers(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateUser(types.CreateUserRequest{
		Username: "foo",
		Password: "bar",
	})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		res []*types.UserDB
		err error
	}{
		"case 01: success": {
			res: []*types.UserDB{
				{
					ID:       1,
					Username: "foo",
					Password: "bar",
					Role:     "user",
				},
			},
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := repo.GetAllUsers()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestRepository_GetUserByID(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateUser(types.CreateUserRequest{
		Username: "foo",
		Password: "bar",
	})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		id  int
		res *types.UserDB
		err error
	}{
		"case 01: fail": {
			id:  0,
			res: nil,
			err: sql.ErrNoRows,
		},
		"case 02: success": {
			id: 1,
			res: &types.UserDB{
				ID:       1,
				Username: "foo",
				Password: "bar",
				Role:     "user",
			},
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := repo.GetUserByID(tt.id)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestRepository_UpdateUser(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateUser(types.CreateUserRequest{
		Username: "foo",
		Password: "bar",
		Email:    "foo@bar.com",
		Phone:    "000-00-00",
	})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateUser(types.CreateUserRequest{
		Username: "john",
		Password: "doe",
		Email:    "john@doe.com",
		Phone:    "111-11-11",
	})
	require.NoError(t, err)
	require.Equal(t, id, 2)

	tests := map[string]struct {
		id  int
		req types.UpdateUserRequest
		err error
	}{
		"case 01: fail": {
			id: 1,
			req: types.UpdateUserRequest{
				Username: "john",
				Password: "doe",
			},
			err: errors.New("bad request"),
		},
		"case 02: success": {
			id: 1,
			req: types.UpdateUserRequest{
				Username: "foo",
				Password: "bar",
			},
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err = repo.UpdateUser(tt.id, tt.req)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestRepository_UpdateUserById(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateUser(types.CreateUserRequest{
		Username: "foo",
		Password: "bar",
		Email:    "foo@bar.com",
		Phone:    "000-00-00",
	})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateUser(types.CreateUserRequest{
		Username: "john",
		Password: "doe",
		Email:    "john@doe.com",
		Phone:    "111-11-11",
	})
	require.NoError(t, err)
	require.Equal(t, id, 2)

	tests := map[string]struct {
		id  int
		req types.UpdateUserByIdRequest
		err error
	}{
		"case 01: fail": {
			id: 1,
			req: types.UpdateUserByIdRequest{
				Username: "john",
				Password: "doe",
				Email:    "john@doe.com",
				Phone:    "111-11-11",
				Role:     "admin",
			},
			err: errors.New("bad request"),
		},
		"case 02: success": {
			id: 2,
			req: types.UpdateUserByIdRequest{
				Username: "john",
				Password: "doe",
				Email:    "john@doe.com",
				Phone:    "111-11-11",
				Role:     "admin",
			},
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err = repo.UpdateUserById(tt.id, tt.req)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestRepository_DeleteUser(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateUser(types.CreateUserRequest{
		Username: "foo",
		Password: "bar",
		Email:    "foo@bar.com",
		Phone:    "000-00-00",
	})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		id  int
		err error
	}{
		"case 01: success": {
			id:  1,
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err = repo.DeleteUser(tt.id)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestRepository_GetAllBooks(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateAuthor(types.CreateAuthorRequest{Name: "Jane"})
	require.NoError(t, err)
	require.Equal(t, id, 2)

	id, err = repo.CreateGenre(types.CreateGenreRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateGenre(types.CreateGenreRequest{Name: "bar"})
	require.NoError(t, err)
	require.Equal(t, id, 2)

	id, err = repo.CreateBook(types.CreateBookRequest{
		AuthorId: 1,
		GenreId:  1,
		Title:    "foo",
	})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	var createdAt, updatedAt time.Time
	err = db.QueryRow(`select created_at, updated_at from books where id = 1`).Scan(&createdAt, &updatedAt)
	require.NoError(t, err)

	id, err = repo.CreateBook(types.CreateBookRequest{
		AuthorId: 2,
		GenreId:  2,
		Title:    "bar",
	})
	require.NoError(t, err)
	require.Equal(t, id, 2)

	var createdAt2, updatedAt2 time.Time
	err = db.QueryRow(`select created_at, updated_at from books where id = 2`).Scan(&createdAt2, &updatedAt2)
	require.NoError(t, err)

	tests := map[string]struct {
		req types.GetAllBooksRequest
		res []*types.BookDB
		err error
	}{
		"case 01: fail author_id": {
			req: types.GetAllBooksRequest{
				Filter: "author_id",
				ID:     "1a",
			},
			res: nil,
			err: errors.New("bad author id"),
		},
		"case 02: success author_id": {
			req: types.GetAllBooksRequest{
				Filter: "author_id",
				ID:     "1",
			},
			res: []*types.BookDB{
				{
					ID: 1,
					Author: types.AuthorDB{
						ID:   1,
						Name: "John",
					},
					Genre: types.GenreDB{
						ID:   1,
						Name: "foo",
					},
					Title:     "foo",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				},
			},
			err: nil,
		},
		"case 03: fail genre_id": {
			req: types.GetAllBooksRequest{
				Filter: "genre_id",
				ID:     "1a",
			},
			res: nil,
			err: errors.New("bad genre id"),
		},
		"case 04: success genre_id": {
			req: types.GetAllBooksRequest{
				Filter: "genre_id",
				ID:     "2",
			},
			res: []*types.BookDB{
				{
					ID: 2,
					Author: types.AuthorDB{
						ID:   2,
						Name: "Jane",
					},
					Genre: types.GenreDB{
						ID:   2,
						Name: "bar",
					},
					Title:     "bar",
					CreatedAt: createdAt2,
					UpdatedAt: updatedAt2,
				},
			},
			err: nil,
		},
		"case 05: success updated_at asc": {
			req: types.GetAllBooksRequest{
				SortBy:  "updated_at",
				OrderBy: "asc",
			},
			res: []*types.BookDB{
				{
					ID: 1,
					Author: types.AuthorDB{
						ID:   1,
						Name: "John",
					},
					Genre: types.GenreDB{
						ID:   1,
						Name: "foo",
					},
					Title:     "foo",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				},
				{
					ID: 2,
					Author: types.AuthorDB{
						ID:   2,
						Name: "Jane",
					},
					Genre: types.GenreDB{
						ID:   2,
						Name: "bar",
					},
					Title:     "bar",
					CreatedAt: createdAt2,
					UpdatedAt: updatedAt2,
				},
			},
			err: nil,
		},
		"case 06: success updated_at desc": {
			req: types.GetAllBooksRequest{
				SortBy:  "updated_at",
				OrderBy: "desc",
			},
			res: []*types.BookDB{
				{
					ID: 2,
					Author: types.AuthorDB{
						ID:   2,
						Name: "Jane",
					},
					Genre: types.GenreDB{
						ID:   2,
						Name: "bar",
					},
					Title:     "bar",
					CreatedAt: createdAt2,
					UpdatedAt: updatedAt2,
				},
				{
					ID: 1,
					Author: types.AuthorDB{
						ID:   1,
						Name: "John",
					},
					Genre: types.GenreDB{
						ID:   1,
						Name: "foo",
					},
					Title:     "foo",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				},
			},
			err: nil,
		},
		"case 07: success title asc": {
			req: types.GetAllBooksRequest{
				SortBy:  "title",
				OrderBy: "asc",
			},
			res: []*types.BookDB{
				{
					ID: 2,
					Author: types.AuthorDB{
						ID:   2,
						Name: "Jane",
					},
					Genre: types.GenreDB{
						ID:   2,
						Name: "bar",
					},
					Title:     "bar",
					CreatedAt: createdAt2,
					UpdatedAt: updatedAt2,
				},
				{
					ID: 1,
					Author: types.AuthorDB{
						ID:   1,
						Name: "John",
					},
					Genre: types.GenreDB{
						ID:   1,
						Name: "foo",
					},
					Title:     "foo",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				},
			},
			err: nil,
		},
		"case 08: success title desc": {
			req: types.GetAllBooksRequest{
				SortBy:  "title",
				OrderBy: "desc",
			},
			res: []*types.BookDB{
				{
					ID: 1,
					Author: types.AuthorDB{
						ID:   1,
						Name: "John",
					},
					Genre: types.GenreDB{
						ID:   1,
						Name: "foo",
					},
					Title:     "foo",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				},
				{
					ID: 2,
					Author: types.AuthorDB{
						ID:   2,
						Name: "Jane",
					},
					Genre: types.GenreDB{
						ID:   2,
						Name: "bar",
					},
					Title:     "bar",
					CreatedAt: createdAt2,
					UpdatedAt: updatedAt2,
				},
			},
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := repo.GetAllBooks(tt.req)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestRepository_GetBookByID(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateGenre(types.CreateGenreRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateBook(types.CreateBookRequest{
		AuthorId: 1,
		GenreId:  1})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	var createdAt, updatedAt time.Time
	err = db.QueryRow(`select created_at,updated_at from books where id = 1`).Scan(&createdAt, &updatedAt)
	require.NoError(t, err)

	tests := map[string]struct {
		id  int
		res *types.BookDB
		err error
	}{
		"case 01: fail": {
			id:  0,
			res: nil,
			err: sql.ErrNoRows,
		},
		"case 02: success": {
			id: 1,
			res: &types.BookDB{
				ID: 1,
				Author: types.AuthorDB{
					ID:   1,
					Name: "John",
				},
				Genre: types.GenreDB{
					ID:   1,
					Name: "foo",
				},
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := repo.GetBookByID(tt.id)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestRepository_CreateBook(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateGenre(types.CreateGenreRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		req types.CreateBookRequest
		id  int
		err error
	}{
		"case 01: success": {
			req: types.CreateBookRequest{
				AuthorId: 1,
				GenreId:  1,
			},
			id:  1,
			err: nil,
		},
		"case 02: bad request": {
			req: types.CreateBookRequest{
				AuthorId: 0,
				GenreId:  0,
			},
			id:  0,
			err: errors.New("bad request"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			id, err = repo.CreateBook(tt.req)
			require.Equal(t, tt.id, id)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestRepository_UpdateBook(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateGenre(types.CreateGenreRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateBook(types.CreateBookRequest{
		AuthorId: 1,
		GenreId:  1})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		id  int
		req types.UpdateBookRequest
		err error
	}{
		"case 01: bad request": {
			id: 1,
			req: types.UpdateBookRequest{
				AuthorId: 0,
				GenreId:  0,
			},
			err: errors.New("bad request"),
		},
		"case 02: success": {
			id: 1,
			req: types.UpdateBookRequest{
				AuthorId: 1,
				GenreId:  1,
				Title:    "test",
			},
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err = repo.UpdateBook(tt.id, tt.req)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestRepository_DeleteBook(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateGenre(types.CreateGenreRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateBook(types.CreateBookRequest{
		AuthorId: 1,
		GenreId:  1,
	})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		id       int
		filename string
		err      error
	}{
		"case 01: bad request": {
			id:       0,
			filename: "",
			err:      sql.ErrNoRows,
		},
		"case 02: success": {
			id:       1,
			filename: "",
			err:      nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			filename, err := repo.DeleteBook(tt.id)
			require.Equal(t, tt.filename, filename)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestRepository_UpdateFilename(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateGenre(types.CreateGenreRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateBook(types.CreateBookRequest{
		AuthorId: 1,
		GenreId:  1,
	})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		id          int
		filename    string
		oldFilename string
		err         error
	}{
		"case 01: bad request": {
			id:          0,
			filename:    "foo",
			oldFilename: "",
			err:         sql.ErrNoRows,
		},
		"case 02: success": {
			id:          1,
			filename:    "foo",
			oldFilename: "",
			err:         nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			oldFilename, err := repo.UpdateFilename(tt.id, tt.filename)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.oldFilename, oldFilename)
		})
	}
}

func TestRepository_GetFilename(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateGenre(types.CreateGenreRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateBook(types.CreateBookRequest{
		AuthorId: 1,
		GenreId:  1,
	})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		id       int
		filename string
		err      error
	}{
		"case 01: bad request": {
			id:       0,
			filename: "",
			err:      sql.ErrNoRows,
		},
		"case 02: success": {
			id:       1,
			filename: "",
			err:      nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			filename, err := repo.GetFilename(tt.id)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.filename, filename)
		})
	}
}

func TestRepository_GetAllAuthors(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		res []*types.AuthorDB
		err error
	}{
		"case 01: success": {
			res: []*types.AuthorDB{
				{
					ID:   1,
					Name: "John",
				},
			},
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := repo.GetAllAuthors()
			require.Equal(t, tt.res, res)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestRepository_GetAuthorById(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		id  int
		res *types.AuthorDB
		err error
	}{
		"case 01: bad request": {
			id:  0,
			res: nil,
			err: sql.ErrNoRows,
		},
		"case 02: success": {
			id: 1,
			res: &types.AuthorDB{
				ID:   1,
				Name: "John",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := repo.GetAuthorById(tt.id)
			require.Equal(t, tt.res, res)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestRepository_CreateAuthor(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		req types.CreateAuthorRequest
		id  int
		err error
	}{
		"case 01: success": {
			req: types.CreateAuthorRequest{
				Name: "Jane",
			},
			id:  2,
			err: nil,
		},
		"case 02: bad request": {
			req: types.CreateAuthorRequest{
				Name: "John",
			},
			id:  0,
			err: errors.New("bad request"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			id, err = repo.CreateAuthor(tt.req)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.id, id)
		})
	}
}

func TestRepository_UpdateAuthor(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateAuthor(types.CreateAuthorRequest{Name: "Jane"})
	require.NoError(t, err)
	require.Equal(t, id, 2)

	tests := map[string]struct {
		id  int
		req types.UpdateAuthorRequest
		err error
	}{
		"case 01: success": {
			id: 1,
			req: types.UpdateAuthorRequest{
				Name: "foo",
			},
			err: nil,
		},
		"case 02: bad request": {
			id: 1,
			req: types.UpdateAuthorRequest{
				Name: "Jane",
			},
			err: errors.New("bad request"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err = repo.UpdateAuthor(tt.id, tt.req)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestRepository_DeleteAuthor(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateAuthor(types.CreateAuthorRequest{Name: "Jane"})
	require.NoError(t, err)
	require.Equal(t, id, 2)

	id, err = repo.CreateGenre(types.CreateGenreRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateBook(types.CreateBookRequest{
		AuthorId: 1,
		GenreId:  1,
	})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		id  int
		err error
	}{
		"case 01: bad request": {
			id:  1,
			err: errors.New("cannot delete author"),
		},
		"case 02: success": {
			id:  2,
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err = repo.DeleteAuthor(tt.id)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestRepository_GetAllGenres(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateGenre(types.CreateGenreRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		res []*types.GenreDB
		err error
	}{
		"case 01: success": {
			res: []*types.GenreDB{
				{
					ID:   1,
					Name: "foo",
				},
			},
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := repo.GetAllGenres()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestRepository_CreateGenre(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateGenre(types.CreateGenreRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		req types.CreateGenreRequest
		id  int
		err error
	}{
		"case 01: success": {
			req: types.CreateGenreRequest{
				Name: "bar",
			},
			id:  2,
			err: nil,
		},
		"case 02: bad request": {
			req: types.CreateGenreRequest{
				Name: "foo",
			},
			id:  0,
			err: errors.New("bad request"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			id, err = repo.CreateGenre(tt.req)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.id, id)
		})
	}
}

func TestRepository_UpdateGenre(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateGenre(types.CreateGenreRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, id, 1)
	id, err = repo.CreateGenre(types.CreateGenreRequest{Name: "bar"})
	require.NoError(t, err)
	require.Equal(t, id, 2)

	tests := map[string]struct {
		id  int
		req types.UpdateGenreRequest
		err error
	}{
		"case 01: success": {
			id: 1,
			req: types.UpdateGenreRequest{
				Name: "Doe",
			},
			err: nil,
		},
		"case 02: bad request": {
			id: 1,
			req: types.UpdateGenreRequest{
				Name: "bar",
			},
			err: errors.New("bad request"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err = repo.UpdateGenre(tt.id, tt.req)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestRepository_DeleteGenre(t *testing.T) {
	container := newTestContainer(t)
	defer container.terminate(t)
	db := container.getDB(t)
	defer db.Close()
	repo := NewRepository(db)

	id, err := repo.CreateAuthor(types.CreateAuthorRequest{Name: "John"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateGenre(types.CreateGenreRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	id, err = repo.CreateGenre(types.CreateGenreRequest{Name: "bar"})
	require.NoError(t, err)
	require.Equal(t, id, 2)

	id, err = repo.CreateBook(types.CreateBookRequest{
		AuthorId: 1,
		GenreId:  2})
	require.NoError(t, err)
	require.Equal(t, id, 1)

	tests := map[string]struct {
		id  int
		err error
	}{
		"case 01: success": {
			id:  1,
			err: nil,
		},
		"case 02: fail": {
			id:  2,
			err: errors.New("cannot delete genre"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err = repo.DeleteGenre(tt.id)
			require.Equal(t, tt.err, err)
		})
	}
}
