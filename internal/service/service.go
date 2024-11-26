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
	DeleteBook(id int) error

	GetFilename(id int) (string, error)
	UpdateFilename(id int, filename string) error

	GetAuthors() (*types.ListAuthors, error)
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
		Items: resp,
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
			ID:       v.ID,
			BookName: v.BookName,
			Author:   v.Author,
			Genre:    v.Genre,
			ISBN:     v.ISBN,
			Filename: v.Filename,
		}
	}

	return &types.ListBookResponse{
		Items: resp,
	}, nil
}

func (s *Service) GetBookById(id int) (*types.Book, error) {
	res, err := s.repo.GetBookByID(id)
	if err != nil {
		return nil, err
	}

	return &types.Book{
		ID:       res.ID,
		BookName: res.BookName,
		Author:   res.Author,
		Genre:    res.Genre,
		ISBN:     res.ISBN,
		Filename: res.Filename,
	}, nil
}

func (s *Service) CreateBook(req types.CreateBookRequest) (*types.CreateBookResponse, error) {
	res, err := s.repo.CreateBook(req)
	if err != nil {
		return nil, err
	}

	return &types.CreateBookResponse{
		ID: res,
	}, nil
}

func (s *Service) UpdateBook(id int, req types.UpdateBookRequest) error {
	return s.repo.UpdateBook(id, req)
}

func (s *Service) DeleteBook(id int) error {
	return s.repo.DeleteBook(id)
}

func (s *Service) GetFilename(id int) (string, error) {
	return s.repo.GetFilename(id)
}

func (s *Service) UpdateFilename(id int, filename string) error {
	return s.repo.UpdateFilename(id, filename)
}

func (s *Service) GetAuthors() (*types.ListAuthors, error) {
	authors, totalAuthors, err := s.repo.GetAuthors()
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

	return &types.ListAuthors{
		TotalAuthors: totalAuthors,
		Authors:      resp,
	}, nil
}
