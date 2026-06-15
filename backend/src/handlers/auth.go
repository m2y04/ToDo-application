package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"todo-backend/src/auth"
	"todo-backend/src/middleware"
	"todo-backend/src/models"

	"github.com/jackc/pgx/v5/pgconn"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var input models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	input.Username = strings.TrimSpace(input.Username)
	if len(input.Username) < 3 {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username must be at least 3 characters"})
		return
	}

	if len(input.Password) < 8 {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		h.logger.Println("password hashing failed:", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	var user models.UserResponse
	query := `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, username
	`

	err = h.db.QueryRow(r.Context(), query, input.Username, passwordHash).Scan(&user.ID, &user.Username)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			h.writeJSON(w, http.StatusConflict, map[string]string{"error": "username already exists"})
			return
		}

		h.logger.Println("user insert failed:", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	token, err := auth.SignToken(h.jwtSecret, user.ID, user.Username)
	if err != nil {
		h.logger.Println("token creation failed:", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create token"})
		return
	}

	h.writeJSON(w, http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var input models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	input.Username = strings.TrimSpace(input.Username)

	var user models.UserResponse
	var passwordHash string
	query := `
		SELECT id, username, password_hash
		FROM users
		WHERE username = $1
	`

	err := h.db.QueryRow(r.Context(), query, input.Username).Scan(&user.ID, &user.Username, &passwordHash)
	if err != nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
		return
	}

	if err := auth.CheckPassword(input.Password, passwordHash); err != nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
		return
	}

	token, err := auth.SignToken(h.jwtSecret, user.ID, user.Username)
	if err != nil {
		h.logger.Println("token creation failed:", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create token"})
		return
	}

	h.writeJSON(w, http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	user, err := middleware.UserFromRequest(h.jwtSecret, r)
	if err != nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]models.UserResponse{
		"user": user,
	})
}
