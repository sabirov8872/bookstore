package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/sabirov8872/bookstore/integral/handler"
	"github.com/sabirov8872/bookstore/integral/types"
)

func Run(handler *handler.Handler, port, secretKey string) {
	router := mux.NewRouter()
	router.HandleFunc("/signup", handler.CreateUser).Methods(http.MethodPost)
	router.HandleFunc("/login", handler.GetUserByUsername).Methods(http.MethodPost)

	router.HandleFunc("/user", AdminAuthMiddleware(secretKey, handler.GetAllUsers)).Methods(http.MethodGet)
	router.HandleFunc("/user/{id}", AdminAuthMiddleware(secretKey, handler.GetUserById)).Methods(http.MethodGet)
	router.HandleFunc("/user/{id}", AdminAuthMiddleware(secretKey, handler.UpdateUser)).Methods(http.MethodPut)
	router.HandleFunc("/user/{id}", AdminAuthMiddleware(secretKey, handler.DeleteUser)).Methods(http.MethodDelete)

	router.HandleFunc("/book", UserAuthMiddleware(secretKey, handler.GetAllBooks)).Methods(http.MethodGet)
	router.HandleFunc("/book/{id}", UserAuthMiddleware(secretKey, handler.GetBookById)).Methods(http.MethodGet)

	router.HandleFunc("/book", AdminAuthMiddleware(secretKey, handler.CreateBook)).Methods(http.MethodPost)
	router.HandleFunc("/book/{id}", AdminAuthMiddleware(secretKey, handler.UpdateBook)).Methods(http.MethodPut)
	router.HandleFunc("/book/{id}", AdminAuthMiddleware(secretKey, handler.DeleteBook)).Methods(http.MethodDelete)

	log.Fatal(http.ListenAndServe("localhost:"+port, router))
}

func UserAuthMiddleware(secretKey string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, "unauthorized")
			return
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
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: err.Error()})
			return
		}

		if !token.Valid {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: "invalid token"})
			return
		}

		handler(w, r)
	}
}

func AdminAuthMiddleware(secretKey string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, "unauthorized")
			return
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
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: err.Error()})
			return
		}

		if !token.Valid {
			writeJSON(w, http.StatusUnauthorized, types.ErrorResponse{Message: "invalid token"})
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

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Println(err)
	}
}
