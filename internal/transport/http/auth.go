package http

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

func (h *Handler) JWTAuth(next http.HandlerFunc) http.HandlerFunc {
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

		if isTokenInvalid(r.Context(), h.logger, authHeaderParts[1]) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func isTokenInvalid(ctx context.Context, logger *slog.Logger, t string) bool {
	signingKey := []byte(getSigningKey(ctx, logger))

	token, err := jwt.Parse(t, func(_ *jwt.Token) (any, error) {
		return signingKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to parse token", slog.Any("error", err))
		return true
	}

	return !token.Valid
}

func getSigningKey(ctx context.Context, logger *slog.Logger) string {
	key := os.Getenv("JWT_SIGNING_KEY")
	if key == "" {
		logger.WarnContext(ctx, "JWT_SIGNING_KEY environment variable is not set, using default key")
		key = "default_secret" // Fallback key (should be avoided in production)
	}
	return key
}
