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
	getAllUsers   = "getAllUsers"
	userID        = "userID"
	bookID        = "bookID"
	getAllAuthors = "getAllAuthors"
	getAllGenres  = "getAllGenres"
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
	GetAuthorById(w http.ResponseWriter, r *http.Request)
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

	token, err := createToken(resp.ID, resp.Role, h.secretKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, types.GetUserByUserResponse{
		UserID: resp.ID,
		Token:  token})
}

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

func (h *Handler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	var req types.GetAllBooksRequest
	req.Filter = r.URL.Query().Get("filter")
	req.ID = r.URL.Query().Get("id")
	req.SortBy = r.URL.Query().Get("sort_by")
	req.OrderBy = r.URL.Query().Get("order_by")

	res, err := h.service.GetAllBooks(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

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

	writeJSON(w, http.StatusOK, res)
}

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

	err = h.cache.Del(r.Context(), bookID+strconv.Itoa(id)).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

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

	if filename != "" {
		err = h.minioClient.DeleteFile(r.Context(), filename)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
			return
		}
	}

	err = h.cache.Del(r.Context(), bookID+strconv.Itoa(id)).Err()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

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

func (h *Handler) GetAuthorById(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid user id"})
		return
	}

	res, err := h.service.GetAuthorById(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

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

	err = h.minioClient.PutFile(r.Context(), fileHeader.Filename, file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	if oldFilename != "no file" {
		err = h.minioClient.DeleteFile(r.Context(), oldFilename)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
			return
		}
	}
}

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

	file, err := h.minioClient.GetFile(r.Context(), filename)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeContent(w, r, filename, time.Now(), file)
}

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

	listGenres, err := h.service.GetAllGenres()
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
