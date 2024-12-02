package service

import (
	"github.com/sabirov8872/bookstore/internal/repository"
	"github.com/sabirov8872/bookstore/internal/types"
)

type Service struct {
	repo repository.IRepository
}

type IService interface {
	CreateUser(req types.CreateUserRequest) (*types.CreateUserResponse, error)
	GetUserByUsername(username string) (*types.GetUserByUsernameDB, error)

	GetAllUsers() (*types.ListUserResponse, error)
	GetUserById(id int) (*types.User, error)
	UpdateUser(id int, req types.UpdateUserRequest) error
	UpdateUserById(id int, userRole types.UpdateUserByIdRequest) error
	DeleteUser(id int) error

	GetAllBooks() (*types.ListBookResponse, error)
	GetBookById(id int) (*types.Book, error)
	CreateBook(req types.CreateBookRequest) (*types.CreateBookResponse, error)
	UpdateBook(id int, req types.UpdateBookRequest) error
	DeleteBook(id int) (string, error)

	GetAllAuthors() (*types.ListAuthorResponse, error)
	CreateAuthor(req types.CreateAuthorRequest) (*types.CreateAuthorResponse, error)
	UpdateAuthor(id int, req types.UpdateAuthorRequest) error
	DeleteAuthor(id int) error

	GetGenres() (*types.ListGenreResponse, error)
	CreateGenre(req types.CreateGenreRequest) (*types.CreateGenreResponse, error)
	UpdateGenre(id int, req types.UpdateGenreRequest) error
	DeleteGenre(id int) error

	UpdateFilename(id int, filename string) (string, error)
	GetFilename(id int) (string, error)
}

func NewService(repo repository.IRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(req types.CreateUserRequest) (*types.CreateUserResponse, error) {
	id, err := s.repo.CreateUser(req)
	if err != nil {
		return nil, err
	}

	return &types.CreateUserResponse{
		ID: id,
	}, nil
}

func (s *Service) GetUserByUsername(username string) (*types.GetUserByUsernameDB, error) {
	return s.repo.GetUserByUsername(username)
}

func (s *Service) GetAllUsers() (*types.ListUserResponse, error) {
	res, err := s.repo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	resp := make([]*types.User, len(res))
	for i, v := range res {
		resp[i] = &types.User{
			ID:       v.ID,
			Username: v.Username,
			Password: v.Password,
			Email:    v.Email,
			Phone:    v.Phone,
			UserRole: v.UserRole,
		}
	}

	return &types.ListUserResponse{
		UsersCount: len(resp),
		Items:      resp,
	}, nil
}

func (s *Service) GetUserById(id int) (*types.User, error) {
	res, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	return &types.User{
		ID:       res.ID,
		Username: res.Username,
		Password: res.Password,
		Email:    res.Email,
		Phone:    res.Phone,
		UserRole: res.UserRole,
	}, nil
}

func (s *Service) UpdateUser(id int, req types.UpdateUserRequest) error {
	return s.repo.UpdateUser(id, req)
}

func (s *Service) UpdateUserById(id int, req types.UpdateUserByIdRequest) error {
	return s.repo.UpdateUserById(id, req)
}

func (s *Service) DeleteUser(id int) error {
	return s.repo.DeleteUser(id)
}

func (s *Service) GetAllBooks() (*types.ListBookResponse, error) {
	res, err := s.repo.GetAllBooks()
	if err != nil {
		return nil, err
	}

	resp := make([]*types.Book, len(res))
	for i, v := range res {
		resp[i] = &types.Book{
			ID:   v.ID,
			Name: v.Name,
			Author: types.Author{
				ID:   v.Author.ID,
				Name: v.Author.Name,
			},
			Genre: types.Genre{
				ID:   v.Genre.ID,
				Name: v.Genre.Name,
			},
			ISBN:        v.ISBN,
			Filename:    v.Filename,
			Description: v.Description,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		}
	}

	return &types.ListBookResponse{
		BooksCount: len(resp),
		Items:      resp,
	}, nil
}

func (s *Service) GetBookById(id int) (*types.Book, error) {
	res, err := s.repo.GetBookByID(id)
	if err != nil {
		return nil, err
	}

	return &types.Book{
		ID:   res.ID,
		Name: res.Name,
		Author: types.Author{
			ID:   res.Author.ID,
			Name: res.Author.Name,
		},
		Genre: types.Genre{
			ID:   res.Genre.ID,
			Name: res.Genre.Name,
		},
		ISBN:        res.ISBN,
		Filename:    res.Filename,
		Description: res.Description,
		CreatedAt:   res.CreatedAt,
		UpdatedAt:   res.UpdatedAt,
	}, nil
}

func (s *Service) CreateBook(req types.CreateBookRequest) (*types.CreateBookResponse, error) {
	id, err := s.repo.CreateBook(req)
	if err != nil {
		return nil, err
	}

	return &types.CreateBookResponse{
		ID: id,
	}, nil
}

func (s *Service) UpdateBook(id int, req types.UpdateBookRequest) error {
	return s.repo.UpdateBook(id, req)
}

func (s *Service) DeleteBook(id int) (string, error) {
	return s.repo.DeleteBook(id)
}

func (s *Service) GetAllAuthors() (*types.ListAuthorResponse, error) {
	authors, err := s.repo.GetAllAuthors()
	if err != nil {
		return nil, err
	}

	resp := make([]*types.Author, len(authors))
	for i, author := range authors {
		resp[i] = &types.Author{
			ID:   author.ID,
			Name: author.Name,
		}
	}

	return &types.ListAuthorResponse{
		AuthorsCount: len(resp),
		Items:        resp,
	}, nil
}

func (s *Service) GetGenres() (*types.ListGenreResponse, error) {
	genres, err := s.repo.GetGenres()
	if err != nil {
		return nil, err
	}

	resp := make([]*types.Genre, len(genres))
	for i, genre := range genres {
		resp[i] = &types.Genre{
			ID:   genre.ID,
			Name: genre.Name,
		}
	}

	return &types.ListGenreResponse{
		GenresCount: len(resp),
		Items:       resp,
	}, nil
}

func (s *Service) UpdateFilename(id int, filename string) (string, error) {
	return s.repo.UpdateFilename(id, filename)
}

func (s *Service) GetFilename(id int) (string, error) {
	return s.repo.GetFilename(id)
}

func (s *Service) CreateAuthor(req types.CreateAuthorRequest) (*types.CreateAuthorResponse, error) {
	id, err := s.repo.CreateAuthor(req)
	if err != nil {
		return nil, err
	}

	return &types.CreateAuthorResponse{
		ID: id,
	}, nil
}

func (s *Service) UpdateAuthor(id int, req types.UpdateAuthorRequest) error {
	return s.repo.UpdateAuthor(id, req)
}

func (s *Service) DeleteAuthor(id int) error {
	return s.repo.DeleteAuthor(id)
}

func (s *Service) CreateGenre(req types.CreateGenreRequest) (*types.CreateGenreResponse, error) {
	res, err := s.repo.CreateGenre(req)
	if err != nil {
		return nil, err
	}

	return &types.CreateGenreResponse{
		ID: res,
	}, nil
}

func (s *Service) UpdateGenre(id int, req types.UpdateGenreRequest) error {
	return s.repo.UpdateGenre(id, req)
}

func (s *Service) DeleteGenre(id int) error {
	return s.repo.DeleteGenre(id)
}
