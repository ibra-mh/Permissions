package models

type Role struct {
	ID          int     `json:"id"`
	RoleKey     string  `json:"role_key"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at"`
}