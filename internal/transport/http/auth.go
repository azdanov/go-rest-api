package http

import (
	"log"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

const signingKey = "secret"

func JWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		authHeaderParts := strings.Split(token, " ")
		if len(authHeaderParts) != 2 || authHeaderParts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		if isTokenInvalid(authHeaderParts[1]) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func isTokenInvalid(t string) bool {
	signingKey := []byte(signingKey)

	token, err := jwt.Parse(t, func(_ *jwt.Token) (any, error) {
		return signingKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		log.Printf("Error parsing token: %v", err)
		return true
	}

	return !token.Valid
}
