package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/sabirov8872/bookstore/internal/minioClient"
	"github.com/sabirov8872/bookstore/internal/service"
	"github.com/sabirov8872/bookstore/internal/types"
	"golang.org/x/crypto/bcrypt"
)

const (
	getAllUsers        = "getAllUsers"
	getAllBooks        = "getAllBooks"
	userID             = "userID"
	bookID             = "bookID"
	getAllAuthors      = "getAllAuthors"
	getAllGenres       = "getAllGenres"
	getBooksByAuthorId = "getBooksByAuthorId"
	getBooksByGenreId  = "getBooksByGenreId"
)

type Handler struct {
	service     service.IService
	secretKey   string
	cache       *redis.Client
	minioClient *minioClient.MinioClient
}

type IHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	GetUserByUsername(w http.ResponseWriter, r *http.Request)

	GetAllUsers(w http.ResponseWriter, r *http.Request)
	GetUserById(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	UpdateUserById(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)

	GetAllBooks(w http.ResponseWriter, r *http.Request)
	GetBookById(w http.ResponseWriter, r *http.Request)
	CreateBook(w http.ResponseWriter, r *http.Request)
	UpdateBook(w http.ResponseWriter, r *http.Request)
	DeleteBook(w http.ResponseWriter, r *http.Request)

	GetAllAuthors(w http.ResponseWriter, r *http.Request)
	CreateAuthor(w http.ResponseWriter, r *http.Request)
	UpdateAuthor(w http.ResponseWriter, r *http.Request)
	DeleteAuthor(w http.ResponseWriter, r *http.Request)

	GetAllGenres(w http.ResponseWriter, r *http.Request)
	CreateGenre(w http.ResponseWriter, r *http.Request)
	UpdateGenre(w http.ResponseWriter, r *http.Request)
	DeleteGenre(w http.ResponseWriter, r *http.Request)

	UploadBookFile(w http.ResponseWriter, r *http.Request)
	GetBookFile(w http.ResponseWriter, r *http.Request)
}

func NewHandler(service service.IService, secretKey string,
	cache *redis.Client, client *minioClient.MinioClient) *Handler {
	return &Handler{
		service:     service,
		secretKey:   secretKey,
		cache:       cache,
		minioClient: client,
	}
}

// CreateUser
//
// @Summary        Registration
// @Description    For users
// @Tags           auth
// @Accept         json
// @Produce        json
// @Param          input body types.CreateUserRequest true "User data"
// @Success        200 {object} types.CreateUserResponse
// @Failure        400 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /sign-up [post]
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req types.CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	req.Password, err = hashingPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	resp, err := h.service.CreateUser(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getAllUsers).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetUserByUsername
//
// @Summary		   User verification
// @Description    For users and admins
// @Tags           auth
// @Accept         json
// @Produce        json
// @Param          input body types.GetUserByUserRequest true "User data"
// @Success        200 {object} types.GetUserByUserResponse
// @Failure		   400 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /sign-in [post]
func (h *Handler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	var req types.GetUserByUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	resp, err := h.service.GetUserByUsername(req.Username)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid username"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(resp.Password), []byte(req.Password))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid password"})
		return
	}

	token, err := createToken(resp.ID, resp.UserRole, h.secretKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, types.GetUserByUserResponse{
		UserID: resp.ID,
		Token:  token})
}

// GetAllUsers
//
// @Summary        Get all users
// @Description    For admins
// @Tags           users
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Success        200 {object} types.ListUserResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /users [get]
func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	data, err := h.cache.Get(r.Context(), getAllUsers).Result()
	if err == nil {
		var resp types.ListUserResponse
		err = json.Unmarshal([]byte(data), &resp)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, resp)
		return
	}

	res, err := h.service.GetAllUsers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	jsonData, err := json.Marshal(res)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Set(r.Context(), getAllUsers, string(jsonData), time.Minute*30).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// GetUserById
//
// @Summary        Get user by id
// @Description    For admins
// @Tags           users
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "User id"
// @Success        200 {object} types.User
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /users/{id} [get]
func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	data, err := h.cache.Get(r.Context(), userID+strconv.Itoa(id)).Result()
	if err == nil {
		var resp types.User
		err = json.Unmarshal([]byte(data), &resp)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, resp)
		return
	}

	res, err := h.service.GetUserById(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	jsonData, err := json.Marshal(res)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Set(r.Context(), userID+strconv.Itoa(id), string(jsonData), time.Minute*30).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// UpdateUser
//
// @Summary        Update user
// @Description    For users and admins
// @Tags           users
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          input body types.UpdateUserRequest true "User data"
// @Success        200
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /users [put]
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	req.Password, err = hashingPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	authHeader := r.Header.Get("Authorization")
	id, err := getUserIdFromToken(authHeader, h.secretKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.service.UpdateUser(id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getAllUsers).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), userID+strconv.Itoa(id)).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

// UpdateUserById
//
// @Summary        Update user by id
// @Description    For admins
// @Tags           users
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "User ID"
// @Param          input body types.UpdateUserByIdRequest true "User data"
// @Success        200
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /users/{id} [put]
func (h *Handler) UpdateUserById(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateUserByIdRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	req.Password, err = hashingPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.service.UpdateUserById(id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getAllUsers).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), userID+strconv.Itoa(id)).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

// DeleteUser
//
// @Summary        Delete user by id
// @Description    For admins
// @Tags           users
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "User ID"
// @Success        200
// @Success        204
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /users/{id} [delete]
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	err = h.service.DeleteUser(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getAllUsers).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), userID+strconv.Itoa(id)).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

// GetAllBooks
//
// @Summary        Get all books
// @Description    For admins, users and guests
// @Tags           books
// @Accept         json
// @Produce        json
// @Success        200 {object} types.ListBookResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /books [get]
func (h *Handler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	data, err := h.cache.Get(r.Context(), getAllBooks).Result()
	if err == nil {
		var resp types.ListBookResponse
		err = json.Unmarshal([]byte(data), &resp)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, resp)
		return
	}

	res, err := h.service.GetAllBooks()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	jsonData, err := json.Marshal(res)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Set(r.Context(), getAllBooks, string(jsonData), 30*time.Minute).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// GetBookById
//
// @Summary        Get book by id
// @Description    For admins, users and guests
// @Tags           books
// @Accept         json
// @Produce        json
// @Param          id path int true "Book id"
// @Success        200 {object} types.Book
// @Failure        400 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /books/{id} [get]
func (h *Handler) GetBookById(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	data, err := h.cache.Get(r.Context(), bookID+strconv.Itoa(id)).Result()
	if err == nil {
		var resp types.Book
		err = json.Unmarshal([]byte(data), &resp)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, resp)
		return
	}

	res, err := h.service.GetBookById(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	jsonData, err := json.Marshal(res)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Set(r.Context(), bookID+strconv.Itoa(id), string(jsonData), 30*time.Minute).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// CreateBook
//
// @Summary        Create a new book
// @Description    For admins
// @Tags           books
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          input body types.CreateBookRequest true "Book data"
// @Success        200 {object} types.CreateBookResponse
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /books [post]
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req types.CreateBookRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	res, err := h.service.CreateBook(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getAllBooks).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getBooksByAuthorId).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getBooksByGenreId).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// UpdateBook
//
// @Summary        Update book by id
// @Description    For admins
// @Tags           books
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "Book ID"
// @Param          input body types.UpdateBookRequest true "Book data"
// @Success        200
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /books/{id} [put]
func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateBookRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	err = h.service.UpdateBook(id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getAllBooks).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getBooksByAuthorId).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getBooksByGenreId).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), bookID+strconv.Itoa(id)).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

// DeleteBook
//
// @Summary        Delete book by id
// @Description    For admins
// @Tags           books
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "Book id"
// @Success        200
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /books/{id} [delete]
func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	filename, err := h.service.DeleteBook(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	if filename != "no file" {
		err = h.minioClient.DeleteBookFile(r.Context(), filename)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
			return
		}
	}

	err = h.cache.Del(r.Context(), getAllBooks).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getBooksByAuthorId).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getBooksByGenreId).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), bookID+strconv.Itoa(id)).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

// GetAllAuthors
//
// @Summary        Get all authors
// @Description    For admins, users and guests
// @Tags           authors
// @Accept         json
// @Produce        json
// @Success        200 {object} types.ListAuthorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /authors [get]
func (h *Handler) GetAllAuthors(w http.ResponseWriter, r *http.Request) {
	data, err := h.cache.Get(r.Context(), getAllAuthors).Result()
	if err == nil {
		var resp types.ListAuthorResponse
		err = json.Unmarshal([]byte(data), &resp)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, resp)
		return
	}

	listAuthors, err := h.service.GetAllAuthors()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	jsonData, err := json.Marshal(listAuthors)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Set(r.Context(), getAllAuthors, string(jsonData), 30*time.Minute).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, listAuthors)
}

// CreateAuthor
//
// @Summary        Create a new author
// @Description    For admins
// @Tags           authors
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          input body types.CreateAuthorRequest true "Author data"
// @Success        200 {object} types.CreateAuthorResponse
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /authors [post]
func (h *Handler) CreateAuthor(w http.ResponseWriter, r *http.Request) {
	var req types.CreateAuthorRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	res, err := h.service.CreateAuthor(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getAllAuthors).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// UpdateAuthor
//
// @Summary        Update author by id
// @Description    For admins
// @Tags           authors
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "Author ID"
// @Param          input body types.UpdateAuthorRequest true "Author data"
// @Success        200
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /authors/{id} [put]
func (h *Handler) UpdateAuthor(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateAuthorRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	err = h.service.UpdateAuthor(id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getAllAuthors).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

// DeleteAuthor
//
// @Summary        Delete author by id
// @Description    For admins
// @Tags           authors
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "Author ID"
// @Success        200
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /authors/{id} [delete]
func (h *Handler) DeleteAuthor(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	err = h.service.DeleteAuthor(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getAllAuthors).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

// UploadBookFile
//
// @Summary       Upload book file by book id
// @Description   For admins
// @Tags          files
// @Accept        multipart/form-data
// @Produce       json
// @Security      ApiKeyAuth
// @Param         id path int true "Book id"
// @Param         file formData file true "Book pdf file"
// @Success       200
// @Failure       400 {object} types.ErrorResponse
// @Failure       401 {object} types.ErrorResponse
// @Failure       500 {object} types.ErrorResponse
// @Router        /files/{id} [post]
func (h *Handler) UploadBookFile(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}
	defer file.Close()

	oldFilename, err := h.service.UpdateFilename(id, fileHeader.Filename)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.minioClient.PutBookFile(r.Context(), fileHeader.Filename, file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	if oldFilename != "no file" {
		err = h.minioClient.DeleteBookFile(r.Context(), oldFilename)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
			return
		}
	}
}

// GetBookFile
//
// @Summary           Get book file by book id
// @Description       For admins, users and guests
// @Tags              files
// @Accept            json
// @Produce           application/pdf
// @Param id          path int true "Book id"
// @Success           200 {file} file "Book file"
// @Failure           400 {object} types.ErrorResponse
// @Failure           500 {object} types.ErrorResponse
// @Router            /files/{id} [get]
func (h *Handler) GetBookFile(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	filename, err := h.service.GetFilename(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	file, err := h.minioClient.GetBookFile(r.Context(), filename)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeContent(w, r, filename, time.Now(), file)
}

// GetAllGenres
//
// @Summary        Get all genres
// @Description    For admins, users and guests
// @Tags           genres
// @Accept         json
// @Produce        json
// @Success        200 {object} types.ListGenreResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /genres [get]
func (h *Handler) GetAllGenres(w http.ResponseWriter, r *http.Request) {
	data, err := h.cache.Get(r.Context(), getAllGenres).Result()
	if err == nil {
		var resp types.ListGenreResponse
		err = json.Unmarshal([]byte(data), &resp)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, resp)
		return
	}

	listGenres, err := h.service.GetGenres()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	jsonData, err := json.Marshal(listGenres)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Set(r.Context(), getAllGenres, string(jsonData), 30*time.Minute).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, listGenres)
}

// CreateGenre
//
// @Summary        Create a new genre
// @Description    For admins
// @Tags           genres
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          input body types.CreateGenreRequest true "Genre data"
// @Success        200 {object} types.CreateGenreResponse
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /genres [post]
func (h *Handler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	var req types.CreateGenreRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	res, err := h.service.CreateGenre(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getAllGenres).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// UpdateGenre
//
// @Summary        Update genre by id
// @Description    For admins
// @Tags           genres
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "Genre ID"
// @Param          input body types.UpdateGenreRequest true "Genre data"
// @Success        200
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /genres/{id} [put]
func (h *Handler) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateGenreRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	err = h.service.UpdateGenre(id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getAllGenres).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

// DeleteGenre
//
// @Summary        Delete genre by id
// @Description    For admins
// @Tags           genres
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "Genre ID"
// @Success        200
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /genres/{id} [delete]
func (h *Handler) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	err = h.service.DeleteGenre(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.cache.Del(r.Context(), getAllGenres).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

func getID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return 0, err
	}

	return id, nil
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Println(err)
	}
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
