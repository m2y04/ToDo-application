package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db        *pgxpool.Pool
	logger    *log.Logger
	jwtSecret string
}

func New(db *pgxpool.Pool, logger *log.Logger, jwtSecret string) *Handler {
	return &Handler{
		db:        db,
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

func (h *Handler) NotImplemented(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "endpoint not implemented yet",
	})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Println("failed to write JSON response:", err)
	}
}
