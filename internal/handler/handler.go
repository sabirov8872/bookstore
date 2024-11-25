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
	"github.com/sabirov8872/bookstore/internal/cache"
	"github.com/sabirov8872/bookstore/internal/minio_client"
	"github.com/sabirov8872/bookstore/internal/service"
	"github.com/sabirov8872/bookstore/internal/types"
	"golang.org/x/crypto/bcrypt"
)

const (
	getAllUsers = "getAllUsers"
	getAllBooks = "getAllBooks"
	userID      = "userID"
	bookID      = "bookID"
)

type Handler struct {
	service     service.IService
	secretKey   string
	cache       *cache.Cache
	minioClient *minio_client.MinioClient
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

	UploadBookFile(w http.ResponseWriter, r *http.Request)
	GetBookFile(w http.ResponseWriter, r *http.Request)
}

func NewHandler(service service.IService, secretKey string,
	cache *cache.Cache, client *minio_client.MinioClient) *Handler {
	return &Handler{
		service:     service,
		secretKey:   secretKey,
		cache:       cache,
		minioClient: client,
	}
}

// CreateUser
//
// @Summary        Create a new user
// @Description    A new user will be created. Hashed before saving to database.
// @Tags           auth
// @Accept         json
// @Produce        json
// @Param          input body types.CreateUserRequest true "User data"
// @Success        200 {object} types.CreateUserResponse
// @Failure        400 {object} types.CreateUserResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /auth/sign-up [post]
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req types.CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	req.Password, err = hashingPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	resp, err := h.service.CreateUser(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
	h.cache.Delete(getAllUsers)
}

// GetUserByUsername
//
// @Summary		   User verification
// @Description    User username and password are checked
// @Tags           auth
// @Accept         json
// @Produce        json
// @Param          input body types.GetUserByUserRequest true "User data"
// @Success        200 {object} types.GetUserByUserResponse
// @Failure		   400 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /auth/sign-in [post]
func (h *Handler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	var req types.GetUserByUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	resp, err := h.service.GetUserByUsername(req.Username)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: "invalid username"})
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(resp.Password),
		[]byte(req.Password))

	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: "invalid password"})
		return
	}

	token, err := createToken(resp.ID, resp.UserRole, h.secretKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, types.GetUserByUserResponse{
		UserID: resp.ID,
		Token:  token})
}

// GetAllUsers
//
// @Summary        Get all users
// @Description    All user data is retrieved from the database.
// @Tags           users
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Success        200 {object} types.ListUserResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /users [get]
func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if data, ok := h.cache.Get(getAllUsers); ok {
		writeJSON(w, http.StatusOK, data)
		return
	}

	res, err := h.service.GetAllUsers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, res)
	h.cache.Set(getAllUsers, res, 30*time.Minute)
}

// GetUserById
//
// @Summary        Get user by id
// @Description    User information is obtained by id.
// @Tags           users
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "User id"
// @Success        200 {object} types.User
// @Success        204
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Router         /users/{id} [get]
func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: "invalid user id"})
		return
	}

	if data, ok := h.cache.Get(userID + strconv.Itoa(id)); ok {
		writeJSON(w, http.StatusOK, data)
		return
	}

	res, err := h.service.GetUserById(id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSON(w, http.StatusOK, res)
	h.cache.Set(userID+strconv.Itoa(id), res, 30*time.Minute)
}

// UpdateUser
//
// @Summary        Update user
// @Description    User information is updated using the provided information.
// @Tags           users
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          input body types.UpdateUserRequest true "User data"
// @Success        200
// @Success        204
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /users [put]
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	req.Password, err = hashingPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	authHeader := r.Header.Get("Authorization")
	id, err := getUserIdFromToken(authHeader, h.secretKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.service.UpdateUser(id, req)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.cache.Delete(getAllUsers)
	h.cache.Delete(userID + strconv.Itoa(id))
}

// UpdateUserById
//
// @Summary        Update user by id
// @Description    The user will be updated using the given data.
// @Tags           users
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "User ID"
// @Param          input body types.UpdateUserByIdRequest true "User data"
// @Success        200
// @Success	       204
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Router         /users/{id} [put]
func (h *Handler) UpdateUserById(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateUserByIdRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: "invalid user id"})
		return
	}

	req.Password, err = hashingPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.service.UpdateUserById(id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.cache.Delete(getAllUsers)
	h.cache.Delete(userID + strconv.Itoa(id))
}

// DeleteUser
//
// @Summary        Delete user by id
// @Description    User data will be deleted using the given information.
// @Tags           users
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "User ID"
// @Success        200
// @Success        204
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Router         /users/{id} [delete]
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: "invalid user id"})
		return
	}

	err = h.service.DeleteUser(id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.cache.Delete(getAllUsers)
	h.cache.Delete(userID + strconv.Itoa(id))
}

// GetAllBooks
//
// @Summary        Get all books
// @Description    All book data is retrieved from the database.
// @Tags           books
// @Accept         json
// @Produce        json
// @Success        200 {object} types.ListBookResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /books [get]
func (h *Handler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	if data, ok := h.cache.Get(getAllBooks); ok {
		writeJSON(w, http.StatusOK, data)
		return
	}

	res, err := h.service.GetAllBooks()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
	h.cache.Set(getAllBooks, res, 30*time.Minute)
}

// GetBookById
//
// @Summary        Get book by id
// @Description    User information is obtained by id and data.
// @Tags           books
// @Accept         json
// @Produce        json
// @Param          id path int true "Book id"
// @Success        200 {object} types.Book "The requested book data"
// @Success        204
// @Failure        400 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /books/{id} [get]
func (h *Handler) GetBookById(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: "invalid user id"})
		return
	}

	if data, ok := h.cache.Get(bookID + strconv.Itoa(id)); ok {
		writeJSON(w, http.StatusOK, data)
		return
	}

	res, err := h.service.GetBookById(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
	h.cache.Set(bookID+strconv.Itoa(id), res, 30*time.Minute)
}

// CreateBook
//
// @Summary        Create a new book
// @Description    A new book is created using the given information.
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
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	res, err := h.service.CreateBook(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
	h.cache.Delete(getAllBooks)
}

// UpdateBook
//
// @Summary        Update book by id
// @Description    The book will be updated using the given id and data.
// @Tags           books
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "Book ID"
// @Param          input body types.UpdateBookRequest true "Book data"
// @Success        200
// @Success        204
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /books/{id} [put]
func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateBookRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: "invalid user id"})
		return
	}

	err = h.service.UpdateBook(id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	h.cache.Delete(getAllBooks)
	h.cache.Delete(bookID + strconv.Itoa(id))
}

// DeleteBook
//
// @Summary        Delete book by id
// @Description    The book will be deleted using the given data.
// @Tags           books
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "Book id"
// @Success        200
// @Success        204
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /books/{id} [delete]
func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	filename, err := h.service.GetFilename(id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = h.service.DeleteBook(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	if filename != "no file" {
		err = h.minioClient.DeleteBookFile(r.Context(), filename)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, err)
			return
		}
	}

	h.cache.Delete(getAllBooks)
	h.cache.Delete(bookID + strconv.Itoa(id))
}

// UploadBookFile
//
// @Summary Upload a new book file
// @Description Upload a file for the specified book ID, replacing the existing file if any.
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Book id"
// @Param file formData file true "Book file"
// @Success 200
// @Success 204
// @Failure 400 {object} types.ErrorResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /files/{id} [post]
func (h *Handler) UploadBookFile(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: "invalid user id"})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: err.Error()})
		return
	}
	defer file.Close()

	filename, err := h.service.GetFilename(id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if filename != "no file" {
		err = h.minioClient.DeleteBookFile(r.Context(), filename)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError,
				types.ErrorResponse{Message: err.Error()})
			return
		}
	}

	err = h.minioClient.PutBookFile(r.Context(), fileHeader.Filename, file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.service.UpdateFilename(id, fileHeader.Filename)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}
}

// GetBookFile
//
// @Summary           Get book file
// @Description       This endpoint retrieves an object from the Minio bucket by its name.
// @Tags              files
// @Accept            json
// @Produce           application/pdf
// @Param id          path int true "Book id"
// @Success           200 {file} file "Book file"
// @Success           204
// @Failure           400 {object} types.ErrorResponse
// @Failure           401 {object} types.ErrorResponse
// @Failure           500 {object} types.ErrorResponse
// @Router            /files/{id} [get]
func (h *Handler) GetBookFile(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			types.ErrorResponse{Message: "invalid user id"})
		return
	}

	filename, err := h.service.GetFilename(id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	file, err := h.minioClient.GetBookFile(r.Context(), filename)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			types.ErrorResponse{Message: err.Error()})
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeContent(w, r, filename, time.Now(), file)
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
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func getUserIdFromToken(authHeader, secretKey string) (int, error) {
	if strings.HasPrefix(authHeader, "Bearer ") {
		authHeader = authHeader[len("Bearer "):]
	}

	token, _ := jwt.Parse(
		authHeader,
		func(token *jwt.Token) (interface{}, error) {
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
