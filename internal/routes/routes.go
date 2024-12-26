package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sabirov8872/bookstore/internal/handler"
	"github.com/sabirov8872/bookstore/internal/repository"
	"github.com/sabirov8872/bookstore/internal/types"
)

func Run(hand handler.IHandler, port int, repo repository.IRepository) {
	r := mux.NewRouter()
	r.HandleFunc("/signup", hand.CreateUser).Methods("POST")
	r.HandleFunc("/login", hand.GetSessionIdByUsername).Methods("POST")
	r.HandleFunc("/logout", hand.DeleteSessionId).Methods("POST")

	r.HandleFunc("/users", AdminAuth(repo, hand.GetAllUsers)).Methods("GET")
	r.HandleFunc("/users", UserAuth(repo, hand.UpdateUserBySessionId)).Methods("PUT")
	r.HandleFunc("/users/{id}", AdminAuth(repo, hand.GetUserById)).Methods("GET")
	r.HandleFunc("/users/{id}", AdminAuth(repo, hand.UpdateUserById)).Methods("PUT")
	r.HandleFunc("/users/{id}", AdminAuth(repo, hand.DeleteUser)).Methods("DELETE")

	r.HandleFunc("/books", hand.GetAllBooks).Methods("GET")
	r.HandleFunc("/books", AdminAuth(repo, hand.CreateBook)).Methods("POST")
	r.HandleFunc("/books/{id}", hand.GetBookById).Methods("GET")
	r.HandleFunc("/books/{id}", AdminAuth(repo, hand.UpdateBookById)).Methods("PUT")
	r.HandleFunc("/books/{id}", AdminAuth(repo, hand.DeleteBookById)).Methods("DELETE")

	r.HandleFunc("/authors", hand.GetAllAuthors).Methods("GET")
	r.HandleFunc("/authors", AdminAuth(repo, hand.CreateAuthor)).Methods("POST")
	r.HandleFunc("/authors/{id}", hand.GetAuthorById).Methods("GET")
	r.HandleFunc("/authors/{id}", AdminAuth(repo, hand.UpdateAuthorById)).Methods("PUT")
	r.HandleFunc("/authors/{id}", AdminAuth(repo, hand.DeleteAuthor)).Methods("DELETE")

	r.HandleFunc("/genres", hand.GetAllGenres).Methods("GET")
	r.HandleFunc("/genres", AdminAuth(repo, hand.CreateGenre)).Methods("POST")
	r.HandleFunc("/genres/{id}", AdminAuth(repo, hand.UpdateGenre)).Methods("PUT")
	r.HandleFunc("/genres/{id}", AdminAuth(repo, hand.DeleteGenre)).Methods("DELETE")

	r.HandleFunc("/files/{id}", hand.GetFileByBookId).Methods("GET")
	r.HandleFunc("/files/{id}", AdminAuth(repo, hand.UploadFileByBookId)).Methods("POST")

	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Cookie"}),
		handlers.AllowCredentials(),
	)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", port), cors(r)))
}

func UserAuth(repo repository.IRepository, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := checkRole(r, repo, []int{1, 2})
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: err.Error()})
			return
		}

		handler(w, r)
	}
}

func AdminAuth(repo repository.IRepository, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := checkRole(r, repo, []int{2})
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: err.Error()})
			return
		}

		handler(w, r)
	}
}

func checkRole(r *http.Request, repo repository.IRepository, roleIds []int) error {
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		return err
	}

	resRoleId, err := repo.GetUserRoleBySessionId(cookie.Value)
	if err != nil {
		return err
	}

	for _, roleId := range roleIds {
		if roleId == resRoleId {
			return nil
		}
	}

	return errors.New("unknown role")
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Println(err)
	}
}
