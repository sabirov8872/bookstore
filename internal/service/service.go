package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sabirov8872/bookstore/internal/repository"
	"github.com/sabirov8872/bookstore/internal/types"
	"github.com/sabirov8872/bookstore/pkg/minio"
	"github.com/sabirov8872/bookstore/pkg/redis"
	"golang.org/x/crypto/bcrypt"
)

const (
	allUsers   = "allUsers"
	userID     = "userID"
	bookID     = "bookId"
	allAuthors = "allAuthors"
	authorID   = "authorID"
	allGenres  = "allGenres"
)

type Service struct {
	repo      repository.IRepository
	redis     redis.IClient
	minio     minio.IClient
	secretKey string
}

type IService interface {
	CreateUser(req types.CreateUserRequest) (*types.CreateUserResponse, error)
	GetUserByUsername(req types.GetUserByUserRequest) (*types.GetUserByUserResponse, error)

	GetAllUsers() (*types.ListUserResponse, error)
	GetUserById(id int) (*types.User, error)
	UpdateUser(req types.UpdateUserRequest, authHeader string) error
	UpdateUserById(id int, userRole types.UpdateUserByIdRequest) error
	DeleteUser(id int) error

	GetAllBooks(req types.GetAllBooksRequest) (*types.ListBookResponse, error)
	GetBookById(id int) (*types.Book, error)
	CreateBook(req types.CreateBookRequest) (*types.CreateBookResponse, error)
	UpdateBook(id int, req types.UpdateBookRequest) error
	DeleteBook(id int) error

	GetAllAuthors() (*types.ListAuthorResponse, error)
	GetAuthorById(id int) (*types.Author, error)
	CreateAuthor(req types.CreateAuthorRequest) (*types.CreateAuthorResponse, error)
	UpdateAuthor(id int, req types.UpdateAuthorRequest) error
	DeleteAuthor(id int) error

	GetAllGenres() (*types.ListGenreResponse, error)
	CreateGenre(req types.CreateGenreRequest) (*types.CreateGenreResponse, error)
	UpdateGenre(id int, req types.UpdateGenreRequest) error
	DeleteGenre(id int) error

	UploadFileByBookId(req types.UploadFileByBookIdRequest) error
	GetFileByBookId(id int) (res *types.GetFileByBookIdResponse, err error)
}

func NewService(repo repository.IRepository, redis redis.IClient, minio minio.IClient, secretKey string) *Service {
	return &Service{
		repo:      repo,
		redis:     redis,
		minio:     minio,
		secretKey: secretKey,
	}
}

func (s *Service) CreateUser(req types.CreateUserRequest) (*types.CreateUserResponse, error) {
	var err error
	req.Password, err = hashingPassword(req.Password)
	if err != nil {
		return nil, err
	}

	id, err := s.repo.CreateUser(req)
	if err != nil {
		return nil, err
	}

	err = s.redis.Del(context.Background(), []string{allUsers})
	if err != nil {
		return nil, err
	}

	return &types.CreateUserResponse{
		ID: id,
	}, nil
}

func (s *Service) GetUserByUsername(req types.GetUserByUserRequest) (*types.GetUserByUserResponse, error) {
	res, err := s.repo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(res.Password), []byte(req.Password))
	if err != nil {
		return nil, err
	}

	token, err := createToken(res.ID, res.Role, s.secretKey)
	if err != nil {
		return nil, err
	}

	return &types.GetUserByUserResponse{
		UserID: res.ID,
		Token:  token,
	}, nil
}

func (s *Service) GetAllUsers() (*types.ListUserResponse, error) {
	if data, err := s.redis.Get(context.Background(), allUsers); err == nil {
		res := &types.ListUserResponse{}
		err = json.Unmarshal([]byte(data), res)
		if err != nil {
			return nil, err
		}

		return res, nil
	}

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
			Role:     v.Role,
		}
	}

	data := &types.ListUserResponse{
		UsersCount: len(resp),
		Items:      resp,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = s.redis.Set(context.Background(), allUsers, jsonData, time.Minute*30)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Service) GetUserById(id int) (*types.User, error) {
	if data, err := s.redis.Get(context.Background(), userID+strconv.Itoa(id)); err == nil {
		res := &types.User{}
		err = json.Unmarshal([]byte(data), res)
		if err != nil {
			return nil, err
		}

		return res, nil
	}
	res, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	data := &types.User{
		ID:       res.ID,
		Username: res.Username,
		Password: res.Password,
		Email:    res.Email,
		Phone:    res.Phone,
		Role:     res.Role,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = s.redis.Set(context.Background(), userID+strconv.Itoa(id), jsonData, time.Minute*30)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Service) UpdateUser(req types.UpdateUserRequest, authHeader string) error {
	id, err := getUserIdFromToken(authHeader, s.secretKey)
	if err != nil {
		return err
	}

	err = s.repo.UpdateUser(id, req)
	if err != nil {
		return err
	}

	err = s.redis.Del(context.Background(), []string{allUsers, userID + strconv.Itoa(id)})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateUserById(id int, req types.UpdateUserByIdRequest) error {
	err := s.repo.UpdateUserById(id, req)
	if err != nil {
		return err
	}

	err = s.redis.Del(context.Background(), []string{allUsers, userID + strconv.Itoa(id)})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteUser(id int) error {
	err := s.repo.DeleteUser(id)
	if err != nil {
		return err
	}

	err = s.redis.Del(context.Background(), []string{allUsers, userID + strconv.Itoa(id)})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetAllBooks(req types.GetAllBooksRequest) (*types.ListBookResponse, error) {
	res, err := s.repo.GetAllBooks(req)
	if err != nil {
		return nil, err
	}

	resp := make([]*types.Book, len(res))
	for i, v := range res {
		resp[i] = &types.Book{
			ID:    v.ID,
			Title: v.Title,
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
	if data, err := s.redis.Get(context.Background(), bookID+strconv.Itoa(id)); err == nil {
		res := &types.Book{}
		err = json.Unmarshal([]byte(data), res)
		if err != nil {
			return nil, err
		}

		return res, nil
	}

	res, err := s.repo.GetBookByID(id)
	if err != nil {
		return nil, err
	}

	resp := &types.Book{
		ID:    res.ID,
		Title: res.Title,
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
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	err = s.redis.Set(context.Background(), bookID+strconv.Itoa(id), jsonData, time.Minute*30)
	if err != nil {
		return nil, err
	}

	return resp, nil
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
	err := s.repo.UpdateBook(id, req)
	if err != nil {
		return err
	}

	err = s.redis.Del(context.Background(), []string{bookID + strconv.Itoa(id)})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteBook(id int) error {
	filename, err := s.repo.DeleteBook(id)
	if err != nil {
		return err
	}

	err = s.minio.DeleteFile(context.Background(), filename)
	if err != nil {
		return err
	}

	err = s.redis.Del(context.Background(), []string{bookID + strconv.Itoa(id)})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetAllAuthors() (*types.ListAuthorResponse, error) {
	if data, err := s.redis.Get(context.Background(), allAuthors); err == nil {
		res := &types.ListAuthorResponse{}
		err = json.Unmarshal([]byte(data), res)
		if err != nil {
			return nil, err
		}

		return res, nil
	}

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

	data := &types.ListAuthorResponse{
		AuthorsCount: len(resp),
		Items:        resp,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = s.redis.Set(context.Background(), allAuthors, jsonData, time.Minute*30)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Service) GetAuthorById(id int) (*types.Author, error) {
	if data, err := s.redis.Get(context.Background(), authorID+strconv.Itoa(id)); err == nil {
		res := &types.Author{}
		err = json.Unmarshal([]byte(data), res)
		if err != nil {
			return nil, err
		}

		return res, nil
	}

	res, err := s.repo.GetAuthorById(id)
	if err != nil {
		return nil, err
	}

	data := &types.Author{
		ID:   res.ID,
		Name: res.Name,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = s.redis.Set(context.Background(), authorID+strconv.Itoa(id), jsonData, time.Minute*30)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Service) CreateAuthor(req types.CreateAuthorRequest) (*types.CreateAuthorResponse, error) {
	id, err := s.repo.CreateAuthor(req)
	if err != nil {
		return nil, err
	}

	err = s.redis.Del(context.Background(), []string{allAuthors})
	if err != nil {
		return nil, err
	}

	return &types.CreateAuthorResponse{
		ID: id,
	}, nil
}

func (s *Service) UpdateAuthor(id int, req types.UpdateAuthorRequest) error {
	err := s.repo.UpdateAuthor(id, req)
	if err != nil {
		return err
	}

	err = s.redis.Del(context.Background(), []string{allAuthors, authorID + strconv.Itoa(id)})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteAuthor(id int) error {
	err := s.repo.DeleteAuthor(id)
	if err != nil {
		return err
	}

	err = s.redis.Del(context.Background(), []string{allAuthors, authorID + strconv.Itoa(id)})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UploadFileByBookId(req types.UploadFileByBookIdRequest) error {
	oldFilename, err := s.repo.UploadFileByBookId(req.ID, req.FileHeader.Filename)
	if err != nil {
		return err
	}

	if oldFilename != "" {
		err = s.minio.DeleteFile(context.Background(), oldFilename)
		if err != nil {
			return err
		}
	}

	err = s.minio.PutFile(context.Background(), req.FileHeader.Filename, req.File)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetFileByBookId(id int) (res *types.GetFileByBookIdResponse, err error) {
	filename, err := s.repo.GetFileByBookId(id)
	if err != nil {
		return nil, err
	}

	file, err := s.minio.GetFile(context.Background(), filename)
	if err != nil {
		return nil, err
	}

	return &types.GetFileByBookIdResponse{
		Filename: filename,
		File:     file,
	}, nil
}

func (s *Service) GetAllGenres() (*types.ListGenreResponse, error) {
	if data, err := s.redis.Get(context.Background(), allGenres); err == nil {
		res := &types.ListGenreResponse{}
		err = json.Unmarshal([]byte(data), res)
		if err != nil {
			return nil, err
		}

		return res, nil
	}

	genres, err := s.repo.GetAllGenres()
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

	data := &types.ListGenreResponse{
		GenresCount: len(resp),
		Items:       resp,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = s.redis.Set(context.Background(), allGenres, jsonData, time.Minute*30)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Service) CreateGenre(req types.CreateGenreRequest) (*types.CreateGenreResponse, error) {
	res, err := s.repo.CreateGenre(req)
	if err != nil {
		return nil, err
	}

	err = s.redis.Del(context.Background(), []string{allGenres})
	if err != nil {
		return nil, err
	}

	return &types.CreateGenreResponse{
		ID: res,
	}, nil
}

func (s *Service) UpdateGenre(id int, req types.UpdateGenreRequest) error {
	err := s.repo.UpdateGenre(id, req)
	if err != nil {
		return err
	}

	err = s.redis.Del(context.Background(), []string{allGenres})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteGenre(id int) error {
	err := s.repo.DeleteGenre(id)
	if err != nil {
		return err
	}

	err = s.redis.Del(context.Background(), []string{allGenres})
	if err != nil {
		return err
	}

	return nil
}

func createToken(id int, userRole, secretKey string) (string, error) {
	claims := &jwt.MapClaims{
		"id":   id,
		"role": userRole,
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secretKey))
}

func hashingPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func getUserIdFromToken(authHeader, secretKey string) (int, error) {
	if strings.HasPrefix(authHeader, "Bearer ") {
		authHeader = authHeader[len("Bearer "):]
	}

	token, _ := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	userId, ok := claims["id"].(float64)
	if !ok {
		return 0, errors.New("invalid user id")
	}

	return int(userId), nil
}
