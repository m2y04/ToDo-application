package middleware

import (
	"errors"
	"net/http"
	"strings"

	"todo-backend/src/auth"
	"todo-backend/src/models"
)

func UserFromRequest(secret string, r *http.Request) (models.UserResponse, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return models.UserResponse{}, errors.New("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return models.UserResponse{}, errors.New("invalid authorization header")
	}

	claims, err := auth.ValidateToken(secret, parts[1])
	if err != nil {
		return models.UserResponse{}, err
	}

	return models.UserResponse{
		ID:       claims.UserID,
		Username: claims.Username,
	}, nil
}
