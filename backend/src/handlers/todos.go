package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"todo-backend/src/middleware"
	"todo-backend/src/models"
)

func (h *Handler) Todos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listTodos(w, r)
	case http.MethodPost:
		h.createTodo(w, r)
	default:
		h.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Handler) TodoByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		h.updateTodo(w, r)
	case http.MethodDelete:
		h.deleteTodo(w, r)
	default:
		h.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Handler) listTodos(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.UserFromRequest(h.jwtSecret, r)
	if err != nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	query := `
		SELECT id, user_id, title, completed, created_at, updated_at
		FROM todos
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := h.db.Query(r.Context(), query, user.ID)
	if err != nil {
		h.logger.Println("todo list query failed:", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load todos"})
		return
	}
	defer rows.Close()

	todos := []models.TodoResponse{}

	for rows.Next() {
		var todo models.TodoResponse
		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			h.logger.Println("todo row scan failed:", err)
			h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load todos"})
			return
		}

		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		h.logger.Println("todo rows failed:", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load todos"})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string][]models.TodoResponse{
		"todos": todos,
	})
}

func (h *Handler) createTodo(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.UserFromRequest(h.jwtSecret, r)
	if err != nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input models.TodoRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	var todo models.TodoResponse
	query := `
		INSERT INTO todos (user_id, title)
		VALUES ($1, $2)
		RETURNING id, user_id, title, completed, created_at, updated_at
	`

	err = h.db.QueryRow(r.Context(), query, user.ID, input.Title).
		Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		h.logger.Println("todo insert failed:", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create todo"})
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]models.TodoResponse{
		"todo": todo,
	})
}

func (h *Handler) updateTodo(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.UserFromRequest(h.jwtSecret, r)
	if err != nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	todoID, err := todoIDFromPath(r.URL.Path)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid todo id"})
		return
	}

	var input models.TodoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	var todo models.TodoResponse
	query := `
		UPDATE todos
		SET title = $1, completed = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, title, completed, created_at, updated_at
	`

	err = h.db.QueryRow(r.Context(), query, input.Title, input.Completed, todoID, user.ID).
		Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		h.writeJSON(w, http.StatusNotFound, map[string]string{"error": "todo not found"})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]models.TodoResponse{
		"todo": todo,
	})
}

func (h *Handler) deleteTodo(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.UserFromRequest(h.jwtSecret, r)
	if err != nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	todoID, err := todoIDFromPath(r.URL.Path)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid todo id"})
		return
	}

	query := `
		DELETE FROM todos
		WHERE id = $1 AND user_id = $2
	`

	result, err := h.db.Exec(r.Context(), query, todoID, user.ID)
	if err != nil {
		h.logger.Println("todo delete failed:", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete todo"})
		return
	}

	if result.RowsAffected() == 0 {
		h.writeJSON(w, http.StatusNotFound, map[string]string{"error": "todo not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func todoIDFromPath(path string) (int, error) {
	idText := strings.TrimPrefix(path, "/todos/")
	idText = strings.Trim(idText, "/")

	return strconv.Atoi(idText)
}
