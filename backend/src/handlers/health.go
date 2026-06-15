package handlers

import (
	"net/http"

	appdb "todo-backend/src/db"
)

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	dbStatus := "ok"
	statusCode := http.StatusOK

	if err := appdb.Ping(h.db); err != nil {
		h.logger.Println("database health check failed:", err)
		dbStatus = "down"
		statusCode = http.StatusServiceUnavailable
	}

	h.writeJSON(w, statusCode, map[string]string{
		"status":   "ok",
		"database": dbStatus,
	})
}
