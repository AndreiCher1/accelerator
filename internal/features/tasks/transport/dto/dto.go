package dto

import (
	"encoding/json"
	"time"
)

type RequestUploadDTO struct {
	TaskName         string          `json:"task_name" validate:"required"`
	Description      string          `json:"description"`
	MeetingDate      string          `json:"meeting_date"`
	SummaryPrompt    string          `json:"summary_prompt" validate:"required"`
	AdditionalPrompt json.RawMessage `json:"additional_prompt" validate:"required"`
	ASRModel         string          `json:"asr_model" validate:"required"`
	LLMModel         string          `json:"llm_model" validate:"required"`
	Tokens           string          `json:"tokens" validate:"required"`
}

type ResponseUploadDTO struct {
	TaskID      string    `json:"task_id"`
	Status      string    `json:"status"`
	FileName    string    `json:"original_filename"`
	FileType    string    `json:"file_type"`
	WaitSeconds int       `json:"estimated_wait_seconds"`
	CreatedAt   time.Time `json:"created_at"`
}
type ResponseDTO struct {
	TaskID      string    `json:"task_id"`
	Status      string    `json:"status"`
	FileName    string    `json:"original_filename"`
	FileType    string    `json:"file_type"`
	Duration    int       `json:"duration_seconds"`
	WaitSeconds int       `json:"estimated_wait_seconds"`
	CreatedAt   time.Time `json:"created_at"`
}

type GroupIDRequestDTO struct {
	GroupID string `validate:"required,uuid4"`
}