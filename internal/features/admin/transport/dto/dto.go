package dto

import "time"

type UserResponseDTO struct {
	UserID    string    `json:"user_id"`
	Login     string    `json:"login"`
	FullName  string    `json:"full_name"`
	Position  string    `json:"position"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type GroupResponseDTO struct {
	GroupID     string    `json:"group_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MemberCount string    `json:"number_count,omitempty"` // не везде будем передавать, поэтому делаем необязательным
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type MembersResponseDTO struct {
	UserID   string    `json:"user_id"`
	Login    string    `json:"login"`
	FullName string    `json:"full_name"`
	Position string    `json:"position"`
	Role     string    `json:"role"`
	AddedAt  time.Time `json:"added_at"`
}

type PaginationResponseDTO struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type UserIDRequestDTO struct {
	UserID string `validate:"required,uuid4"`
}

type GroupIDRequestDTO struct {
	GroupID string `validate:"required,uuid4"`
}
