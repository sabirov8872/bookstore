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
	GetUserByUsername(username string) (*types.GetUserByUserDB, error)

	GetAllUsers() (*types.ListUserResponse, error)
	GetUserById(id string) (*types.User, error)
	UpdateUser(id string, req types.UpdateUserRequest) error
	DeleteUser(id string) error

	GetAllBooks() (*types.ListBookResponse, error)
	GetBookById(id string) (*types.Book, error)
	CreateBook(req types.CreateBookRequest) (*types.CreateBookResponse, error)
	UpdateBook(id string, req types.UpdateBookRequest) error
	DeleteBook(id string) error
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

func (s *Service) GetUserByUsername(username string) (*types.GetUserByUserDB, error) {
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

func (s *Service) GetUserById(id string) (*types.User, error) {
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

func (s *Service) UpdateUser(id string, req types.UpdateUserRequest) error {
	return s.repo.UpdateUser(id, req)
}

func (s *Service) DeleteUser(id string) error {
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
			ID:     v.ID,
			Name:   v.Name,
			Genre:  v.Genre,
			Author: v.Author,
		}
	}

	return &types.ListBookResponse{
		Items: resp,
	}, nil
}

func (s *Service) GetBookById(id string) (*types.Book, error) {
	res, err := s.repo.GetBookByID(id)
	if err != nil {
		return nil, err
	}

	return &types.Book{
		ID:     res.ID,
		Name:   res.Name,
		Genre:  res.Genre,
		Author: res.Author,
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

func (s *Service) UpdateBook(id string, req types.UpdateBookRequest) error {
	return s.repo.UpdateBook(id, req)
}

func (s *Service) DeleteBook(id string) error {
	return s.repo.DeleteBook(id)
}
