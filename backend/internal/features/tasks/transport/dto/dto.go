package dto

import (
	"encoding/json"
	"time"
)

type RequestUploadDTO struct {
	TaskName    string `json:"task_name" validate:"required"`
	Description string `json:"description" validate:"required"`
	MeetingDate string `json:"meeting_date" validate:"required"`
	PatternID   string `json:"pattern_id" validate:"required"`
}

type ResponseUploadDTO struct {
	TaskID      string `json:"task_id"`
	Status      string `json:"status"`
	TaskName    string `json:"task_name"`
	Description string `json:"description"`
	MeetingDate string `json:"meeting_date"`
	PatternID   string `json:"pattern_id"`

	FileName  string    `json:"original_filename"`
	FileType  string    `json:"file_type"`
	CreatedAt time.Time `json:"created_at"`

	ChangeFlag bool `json:"change_flag"`
}

type ResponseTaskDTO struct {
	TaskID  string `json:"task_id"`
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`

	TaskName    string `json:"task_name"`
	Description string `json:"description"`
	MeetingDate string `json:"meeting_date"`
	PatternID   string `json:"pattern_id"`

	Status     string          `json:"status"`
	ResultJson json.RawMessage `json:"result"`

	FileName string `json:"original_filename"`
	Duration int    `json:"duration_seconds"`

	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`

	ChangeFlag bool `json:"change_flag"`
}

type ResponseCheckTaskDTO struct {
	Status                     string `json:"status"`
	IsProcess                  bool   `json:"is_process"`
	InTheQueueBefore           int    `json:"in_the_queue_before"`
	ApproximateLeadTimeProcess int    `json:"approximate_lead_time_process"`
}

type GroupIDRequestDTO struct {
	GroupID string `validate:"required,uuid4"`
}

type TaskIDRequestDTO struct {
	TaskID string `validate:"required,uuid4"`
}

type AllTasksResponseDTO struct {
	Tasks      []ResponseTaskDTO     `json:"tasks"`
	Pagination PaginationResponseDTO `json:"pagination"`
}

type PaginationRequestDTO struct {
	Page  string `validate:"required,number"`
	Limit string `validate:"required,number"`
}

type PaginationResponseDTO struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type EditTaskRequestDTO struct {
	TaskName    *string `json:"task_name" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
	MeetingDate *string `json:"meeting_date" validate:"omitempty"`
}
