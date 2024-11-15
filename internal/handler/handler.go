package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/sabirov8872/bookstore/internal/cache"
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
	service   service.IService
	secretKey string
	cache     *cache.Cache
}

type IHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	GetUserByUsername(w http.ResponseWriter, r *http.Request)

	GetAllUsers(w http.ResponseWriter, r *http.Request)
	GetUserById(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)

	GetAllBooks(w http.ResponseWriter, r *http.Request)
	GetBookById(w http.ResponseWriter, r *http.Request)
	CreateBook(w http.ResponseWriter, r *http.Request)
	UpdateBook(w http.ResponseWriter, r *http.Request)
	DeleteBook(w http.ResponseWriter, r *http.Request)
}

func NewHandler(service service.IService, secretKey string, cache *cache.Cache) *Handler {
	return &Handler{
		service:   service,
		secretKey: secretKey,
		cache:     cache,
	}
}

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

func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	id := getID(r)

	if data, ok := h.cache.Get(userID + id); ok {
		writeJSON(w, http.StatusOK, data)
		return
	}

	res, err := h.service.GetUserById(id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSON(w, http.StatusOK, res)
	h.cache.Set(userID+id, res, 30*time.Minute)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateUserRequest
	json.NewDecoder(r.Body).Decode(&req)

	hashedPassword, err := hashingPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
	req.Password = hashedPassword

	id := getID(r)
	err = h.service.UpdateUser(id, req)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.cache.Delete(getAllUsers)
	h.cache.Delete(userID + id)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := getID(r)
	err := h.service.DeleteUser(id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.cache.Delete(getAllUsers)
	h.cache.Delete(userID + id)
}

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

func (h *Handler) GetBookById(w http.ResponseWriter, r *http.Request) {
	id := getID(r)

	if data, ok := h.cache.Get(bookID + id); ok {
		writeJSON(w, http.StatusOK, data)
		return
	}

	res, err := h.service.GetBookById(id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSON(w, http.StatusOK, res)
	h.cache.Set(bookID+id, res, 30*time.Minute)
}

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

func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateBookRequest
	json.NewDecoder(r.Body).Decode(&req)

	id := getID(r)
	err := h.service.UpdateBook(id, req)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.cache.Delete(getAllBooks)
	h.cache.Delete(bookID + id)
}

func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id := getID(r)
	err := h.service.DeleteBook(id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.cache.Delete(getAllBooks)
	h.cache.Delete(bookID + id)
}

func getID(r *http.Request) string {
	return mux.Vars(r)["id"]
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Println(err)
	}
}

func createToken(id int64, userRole, secretKey string) (string, error) {
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
