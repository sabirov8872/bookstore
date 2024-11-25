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
	httpSwagger "github.com/swaggo/http-swagger"
)

func Run(hand handler.IHandler, port, secretKey string) {
	router := mux.NewRouter()
	router.HandleFunc("/auth/sign-up", hand.CreateUser).Methods(http.MethodPost)
	router.HandleFunc("/auth/sign-in", hand.GetUserByUsername).Methods(http.MethodPost)

	router.HandleFunc("/users", AdminAuthorization(secretKey, hand.GetAllUsers)).Methods(http.MethodGet)
	router.HandleFunc("/users/{id}", AdminAuthorization(secretKey, hand.GetUserById)).Methods(http.MethodGet)
	router.HandleFunc("/users/{id}", AdminAuthorization(secretKey, hand.UpdateUserById)).Methods(http.MethodPut)
	router.HandleFunc("/users/{id}", AdminAuthorization(secretKey, hand.DeleteUser)).Methods(http.MethodDelete)
	router.HandleFunc("/users", UserAuthorization(secretKey, hand.UpdateUser)).Methods(http.MethodPut)

	router.HandleFunc("/books", hand.GetAllBooks).Methods(http.MethodGet)
	router.HandleFunc("/books/{id}", hand.GetBookById).Methods(http.MethodGet)
	router.HandleFunc("/books", AdminAuthorization(secretKey, hand.CreateBook)).Methods(http.MethodPost)
	router.HandleFunc("/books/{id}", AdminAuthorization(secretKey, hand.UpdateBook)).Methods(http.MethodPut)
	router.HandleFunc("/books/{id}", AdminAuthorization(secretKey, hand.DeleteBook)).Methods(http.MethodDelete)

	router.HandleFunc("/files/{id}", UserAuthorization(secretKey, hand.GetBookFile)).Methods(http.MethodGet)
	router.HandleFunc("/files/{id}", AdminAuthorization(secretKey, hand.UploadBookFile)).Methods(http.MethodPost)

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Fatal(http.ListenAndServe("localhost:"+port, router))
}

func UserAuthorization(secretKey string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		_, err := validateToken(authHeader, secretKey)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: err.Error()})
			return
		}

		handler(w, r)
	}
}

func AdminAuthorization(secretKey string, handler http.HandlerFunc) http.HandlerFunc {
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