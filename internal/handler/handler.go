package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

	PutObject(w http.ResponseWriter, r *http.Request)
	GetObject(w http.ResponseWriter, r *http.Request)
}

func NewHandler(service service.IService, secretKey string, cache *cache.Cache, client *minio_client.MinioClient) *Handler {
	return &Handler{
		service:     service,
		secretKey:   secretKey,
		cache:       cache,
		minioClient: client,
	}
}

// CreateUser
//
// @Summary        Create user
// @Description    Create a new user
// @Tags           auth
// @Accept         json
// @Produce        json
// @Param          input body types.CreateUserRequest true "User data"
// @Success        200 {object} types.CreateUserResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /auth/sign-up [post]
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req types.CreateUserRequest
	json.NewDecoder(r.Body).Decode(&req)

	hashedPassword, err := hashingPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
	req.Password = hashedPassword

	resp, err := h.service.CreateUser(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
	h.cache.Delete(getAllUsers)
}

// GetUserByUsername
//
// @Summary        Check user
// @Description    check username and password
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
	json.NewDecoder(r.Body).Decode(&req)

	resp, err := h.service.GetUserByUsername(req.Username)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, "invalid username")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(resp.Password), []byte(req.Password))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid password"})
		return
	}

	var token string
	token, err = createToken(resp.ID, resp.UserRole, h.secretKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, types.GetUserByUserResponse{UserID: resp.ID, Token: token})
}

// GetAllUsers
//
// @Summary        Get all users
// @Description    get all users
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
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, res)
	h.cache.Set(getAllUsers, res, 30*time.Minute)
}

// GetUserById
//
// @Summary        Get user by id
// @Description    get user by id
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
// @Description    update user
// @Tags           users
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          input body types.UpdateUserRequest true "User data"
// @Success        200
// @Success        204
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /users [put]
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateUserRequest
	json.NewDecoder(r.Body).Decode(&req)

	hashedPassword, err := hashingPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
	req.Password = hashedPassword

	authHeader := r.Header.Get("Authorization")
	id, err := getUserIdFromToken(authHeader, h.secretKey)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: err.Error()})
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
// @Description    update user by id
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
	json.NewDecoder(r.Body).Decode(&req)

	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	err = h.service.UpdateUserById(id, req)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
	}

	h.cache.Delete(getAllUsers)
	h.cache.Delete(userID + strconv.Itoa(id))
}

// DeleteUser
//
// @Summary        Delete user by id
// @Description    delete user by id
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
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
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
// @Description    get all books
// @Tags           books
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Success        200 {object} types.ListBookResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /books [get]
func (h *Handler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	if data, ok := h.cache.Get(getAllBooks); ok {
		writeJSON(w, http.StatusOK, data)
		return
	}

	res, err := h.service.GetAllBooks()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, res)
	h.cache.Set(getAllBooks, res, 30*time.Minute)
}

// GetBookById
//
// @Summary        Get book by id
// @Description    get book by id
// @Tags           books
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "Book ID"
// @Success        200 {object} types.Book
// @Success        204
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Router         /books/{id} [get]
func (h *Handler) GetBookById(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	if data, ok := h.cache.Get(bookID + strconv.Itoa(id)); ok {
		writeJSON(w, http.StatusOK, data)
		return
	}

	res, err := h.service.GetBookById(id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSON(w, http.StatusOK, res)
	h.cache.Set(bookID+strconv.Itoa(id), res, 30*time.Minute)
}

// CreateBook
//
// @Summary        Create book
// @Description    create book
// @Tags           books
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          input body types.CreateBookRequest true "Book data"
// @Success        200 {object} types.CreateBookResponse
// @Failure        401 {object} types.ErrorResponse
// @Failure        500 {object} types.ErrorResponse
// @Router         /books [post]
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req types.CreateBookRequest
	json.NewDecoder(r.Body).Decode(&req)

	res, err := h.service.CreateBook(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, res)
	h.cache.Delete(getAllBooks)
}

// UpdateBook
//
// @Summary        Update book by id
// @Description    update book by id
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
// @Router         /books/{id} [put]
func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateBookRequest
	json.NewDecoder(r.Body).Decode(&req)

	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	err = h.service.UpdateBook(id, req)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.cache.Delete(getAllBooks)
	h.cache.Delete(bookID + strconv.Itoa(id))
}

// DeleteBook
//
// @Summary        Delete book by id
// @Description    delete book by id
// @Tags           books
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "User ID"
// @Success        200
// @Success        204
// @Failure        400 {object} types.ErrorResponse
// @Failure        401 {object} types.ErrorResponse
// @Router         /books/{id} [delete]
func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	err = h.service.DeleteBook(id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.cache.Delete(getAllBooks)
	h.cache.Delete(bookID + strconv.Itoa(id))
}

// PutObject
// @Summary File upload
// @Description Loads a file using multipart/form-data and saves it to MinIO
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File (JPG, PNG, etc.)"
// @Success 200
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /files [post]
func (h *Handler) PutObject(w http.ResponseWriter, r *http.Request) {
	file, fileHeader, err := r.FormFile("file")
	fmt.Println(fileHeader.Filename)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}
	defer file.Close()

	err = h.minioClient.PutObjectToTheMinio(r.Context(), fileHeader.Filename, file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

// GetObject
//
// @Summary Get an object from the Minio
// @Description This endpoint retrieves an object from the Minio bucket by its name.
// @Tags files
// @Accept  json
// @Produce  octet-stream
// @Param objectName path string true "Name of the object to retrieve"
// @Success 200 {file} file "The requested object file"
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /files/{objectName} [get]
func (h *Handler) GetObject(w http.ResponseWriter, r *http.Request) {
	objectName := mux.Vars(r)["objectName"]

	decodedFilename, err := url.PathUnescape(objectName)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid object name"})
		return
	}

	object, err := h.minioClient.GetObjectFromMinio(r.Context(), decodedFilename)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
	defer object.Close()

	//w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+decodedFilename)
	//if _, err = io.Copy(w, object); err != nil {
	//	writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
	//}

	http.ServeContent(w, r, decodedFilename, time.Now(), object)
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
