package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sabirov8872/bookstore/internal/service"
	"github.com/sabirov8872/bookstore/internal/types"
)

type Handler struct {
	service service.IService
}

type IHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	GetSessionIdByUsername(w http.ResponseWriter, r *http.Request)
	DeleteSessionId(w http.ResponseWriter, r *http.Request)

	GetAllUsers(w http.ResponseWriter, r *http.Request)
	GetUserById(w http.ResponseWriter, r *http.Request)
	UpdateUserBySessionId(w http.ResponseWriter, r *http.Request)
	UpdateUserById(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)

	GetAllBooks(w http.ResponseWriter, r *http.Request)
	GetBookById(w http.ResponseWriter, r *http.Request)
	CreateBook(w http.ResponseWriter, r *http.Request)
	UpdateBookById(w http.ResponseWriter, r *http.Request)
	DeleteBookById(w http.ResponseWriter, r *http.Request)

	GetAllAuthors(w http.ResponseWriter, r *http.Request)
	GetAuthorById(w http.ResponseWriter, r *http.Request)
	CreateAuthor(w http.ResponseWriter, r *http.Request)
	UpdateAuthorById(w http.ResponseWriter, r *http.Request)
	DeleteAuthor(w http.ResponseWriter, r *http.Request)

	GetAllGenres(w http.ResponseWriter, r *http.Request)
	CreateGenre(w http.ResponseWriter, r *http.Request)
	UpdateGenre(w http.ResponseWriter, r *http.Request)
	DeleteGenre(w http.ResponseWriter, r *http.Request)

	UploadFileByBookId(w http.ResponseWriter, r *http.Request)
	GetFileByBookId(w http.ResponseWriter, r *http.Request)
}

func NewHandler(service service.IService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req types.CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	resp, err := h.service.CreateUser(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetSessionIdByUsername(w http.ResponseWriter, r *http.Request) {
	var req types.GetSessionIdByUsernameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	res, err := h.service.GetSessionIdByUsername(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: "invalid username"})
		return
	}

	cookie := &http.Cookie{
		Name:     "sessionId",
		Value:    res.SessionId,
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	}

	http.SetCookie(w, cookie)
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) DeleteSessionId(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("sessionId")

	err := h.service.DeleteSessionId(cookie.Value)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: "invalid sessionId"})
		return
	}
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.GetAllUsers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) UpdateUserBySessionId(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	cookie, _ := r.Cookie("sessionId")

	err = h.service.UpdateUserBySessionId(req, cookie.Value)
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

	res, err := h.service.GetUserById(id)
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

	err = h.service.UpdateUserById(id, req)
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

	res, err := h.service.GetBookById(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) UpdateBookById(w http.ResponseWriter, r *http.Request) {
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
}

func (h *Handler) DeleteBookById(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.service.DeleteBook(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

func (h *Handler) GetAllAuthors(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.GetAllAuthors()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
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

func (h *Handler) UpdateAuthorById(w http.ResponseWriter, r *http.Request) {
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
}

func (h *Handler) GetAllGenres(w http.ResponseWriter, r *http.Request) {
	listGenres, err := h.service.GetAllGenres()
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
}

func (h *Handler) UploadFileByBookId(w http.ResponseWriter, r *http.Request) {
	var err error
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid book id"})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: err.Error()})
		return
	}
	defer file.Close()

	err = h.service.UploadFileByBookId(types.UploadFileByBookIdRequest{
		ID:         id,
		File:       file,
		FileHeader: fileHeader,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}
}

func (h *Handler) GetFileByBookId(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{Message: "invalid book id"})
		return
	}

	req, err := h.service.GetFileByBookId(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{Message: err.Error()})
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+req.Filename)
	http.ServeContent(w, r, req.Filename, time.Now(), req.File)
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println(err)
	}
}
