package dto

import "time"

// ---------- Пользователи ----------

type UserResponseDTO struct {
	UserID    string    `json:"user_id"`
	Login     string    `json:"login"`
	FullName  string    `json:"full_name"`
	Position  string    `json:"position"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ---------- Группы ----------

type GroupResponseDTO struct {
	GroupID     string    `json:"group_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MemberCount int       `json:"member_count"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	CanEdit     bool      `json:"can_edit"`
	CanDelete   bool      `json:"can_delete"`
}

type CreateGroupRequestDTO struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description" validate:"required"`
	OwnerID     *string `json:"owner_id" validate:"omitempty"` // опционально: id админа-владельца
}

type EditGroupRequestDTO struct {
	Name        *string `json:"name" validate:"omitempty"` // если передаётся, не может быть пустым
	Description *string `json:"description" validate:"omitempty"`
	OwnerID     *string `json:"owner_id" validate:"omitempty,uuid4"`
}

// Информация о группе + список участников
type GetMembersGroupResponseDTO struct {
	GroupID     string               `json:"group_id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Members     []MembersResponseDTO `json:"members"`
	OwnerID     string               `json:"owner_id"`
	CreatedAt   time.Time            `json:"created_at"`
	CanEdit     bool                 `json:"can_edit"`
	CanDelete   bool                 `json:"can_delete"`
}

// Участник группы
type MembersResponseDTO struct {
	UserID   string    `json:"user_id"`
	Login    string    `json:"login"`
	FullName string    `json:"full_name"`
	Position string    `json:"position"`
	Role     string    `json:"role"`
	AddedAt  time.Time `json:"added_at"`
}

type GetGroupsResponseDTO struct {
	Groups []GroupResponseDTO `json:"groups"`
}

// ---------- Общие ----------

type UserIDRequestDTO struct {
	UserID string `validate:"required,uuid4"`
}

type GroupIDRequestDTO struct {
	GroupID string `validate:"required,uuid4"`
}

type PaginationResponseDTO struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}
