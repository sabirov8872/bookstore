package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/sabirov8872/bookstore/internal/handler"
	"github.com/sabirov8872/bookstore/internal/types"
)

func Run(hand handler.IHandler, port, secretKey string) {
	r := mux.NewRouter()
	r.HandleFunc("/sign-up", hand.CreateUser).Methods("POST")
	r.HandleFunc("/sign-in", hand.GetUserByUsername).Methods("POST")

	r.HandleFunc("/users", AdminAuth(secretKey, hand.GetAllUsers)).Methods("GET")
	r.HandleFunc("/users", UserAuth(secretKey, hand.UpdateUser)).Methods("PUT")
	r.HandleFunc("/users/{id}", AdminAuth(secretKey, hand.GetUserById)).Methods("GET")
	r.HandleFunc("/users/{id}", AdminAuth(secretKey, hand.UpdateUserById)).Methods("PUT")
	r.HandleFunc("/users/{id}", AdminAuth(secretKey, hand.DeleteUser)).Methods("DELETE")

	r.HandleFunc("/books", hand.GetAllBooks).Methods("GET")
	r.HandleFunc("/books", AdminAuth(secretKey, hand.CreateBook)).Methods("POST")
	r.HandleFunc("/books/{id}", hand.GetBookById).Methods("GET")
	r.HandleFunc("/books/{id}", AdminAuth(secretKey, hand.UpdateBook)).Methods("PUT")
	r.HandleFunc("/books/{id}", AdminAuth(secretKey, hand.DeleteBook)).Methods("DELETE")

	r.HandleFunc("/authors", hand.GetAllAuthors).Methods("GET")
	r.HandleFunc("/authors", AdminAuth(secretKey, hand.CreateAuthor)).Methods("POST")
	r.HandleFunc("/authors/{id}", hand.GetAuthorById).Methods("GET")
	r.HandleFunc("/authors/{id}", AdminAuth(secretKey, hand.UpdateAuthor)).Methods("PUT")
	r.HandleFunc("/authors/{id}", AdminAuth(secretKey, hand.DeleteAuthor)).Methods("DELETE")

	r.HandleFunc("/genres", hand.GetAllGenres).Methods("GET")
	r.HandleFunc("/genres", AdminAuth(secretKey, hand.CreateGenre)).Methods("POST")
	r.HandleFunc("/genres/{id}", AdminAuth(secretKey, hand.UpdateGenre)).Methods("PUT")
	r.HandleFunc("/genres/{id}", AdminAuth(secretKey, hand.DeleteGenre)).Methods("DELETE")

	r.HandleFunc("/files/{id}", hand.GetBookFile).Methods("GET")
	r.HandleFunc("/files/{id}", AdminAuth(secretKey, hand.UploadBookFile)).Methods("POST")

	log.Fatal(http.ListenAndServe("localhost:"+port, r))
}

func UserAuth(secretKey string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token, err := validateToken(authHeader, secretKey)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: err.Error()})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: "error reading token"})
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: "error reading role"})
			return
		}

		if role != "admin" && role != "user" {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: "invalid role"})
			return
		}

		handler(w, r)
	}
}

func AdminAuth(secretKey string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token, err := validateToken(authHeader, secretKey)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: err.Error()})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: "error reading token"})
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: "error reading role"})
			return
		}

		if role != "admin" {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: "invalid role"})
			return
		}

		handler(w, r)
	}
}

func validateToken(authHeader, secretKey string) (*jwt.Token, error) {
	if authHeader == "" {
		return nil, errors.New("empty authorization header")
	}

	if strings.HasPrefix(authHeader, "Bearer ") {
		authHeader = authHeader[len("Bearer "):]
	}

	token, err := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Println(err)
	}
}
