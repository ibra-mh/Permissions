package models

type UserRole struct {
	ID        int     `json:"id"`
	Email     string  `json:"email"`
	RoleID    int     `json:"role_id"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	DeletedAt *string `json:"deleted_at"`
}