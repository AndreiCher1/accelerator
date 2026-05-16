package dto

import (
	"encoding/json"
	"time"
)

type CreatePatternRequestDTO struct {
	Name             string          `json:"name" validate:"required"`
	Description      string          `json:"description" validate:"required"`
	SummaryPrompt    string          `json:"summary_prompt" validate:"required"`
	AdditionalPrompt json.RawMessage `json:"additional_prompt"` // будет храниться напрямую в json, если там nil, то в json будет "null"
	GroupID          string          `json:"group_id" validate:"required"`
}

type EditPatternRequestDTO struct {
	Name             *string          `json:"name" validate:"omitempty"`
	Description      *string          `json:"description" validate:"omitempty"`
	SummaryPrompt    *string          `json:"summary_prompt" validate:"omitempty"`
	AdditionalPrompt *json.RawMessage `json:"additional_prompt"`
}

type PatternIDDTO struct {
	PatternID string `json:"pattern_id" validate:"required,uuid4"`
}

type PatternResponseDTO struct {
	ID               string          `json:"pattern_id"`
	GroupID          string          `json:"group_id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	SummaryPrompt    string          `json:"summary_prompt"`
	AdditionalPrompt json.RawMessage `json:"additional_prompt"`
	CreatedAt        time.Time       `json:"created_at"`
	ChangeFlag       bool            `json:"change_flag"`
}

type GroupPatternsResponseDTO struct {
	GlobalPatterns []PatternResponseDTO `json:"global_patterns"`
	GroupPatterns  []PatternResponseDTO `json:"group_patterns"`
}

type GlobalPatternsResponseDTO struct {
	GlobalPatterns []PatternResponseDTO `json:"global_patterns"`
}

type GroupWithPatternsResponse struct {
	GroupID     string               `json:"group_id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Patterns    []PatternResponseDTO `json:"patterns"`
}

type AllGroupWithPatternsResponse struct {
	Groups []GroupWithPatternsResponse `json:"groups"`
}
