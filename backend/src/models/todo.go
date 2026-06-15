package models

import "time"

type TodoRequest struct {
	Title string `json:"title"`
}

type TodoUpdateRequest struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type TodoResponse struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
