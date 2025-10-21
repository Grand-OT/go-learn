package todo

import "time"

const StatusPending = "pending"

type Todo struct {
	ID          int64
	Title       string
	Description *string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TodoDTO struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func ToDTO(t Todo) TodoDTO {
	return TodoDTO{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
}
