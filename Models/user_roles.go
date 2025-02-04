package models

import "time"

type UserRole struct {
	ID        int        `json:"id"`
	Email     string     `json:"email"`
	RoleID    int        `json:"role_id"`
	RoleKey   string     `json:"role_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
