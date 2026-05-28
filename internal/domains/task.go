package domains

import (
	"encoding/json"
	"time"
)

type Task struct {
	TaskID  string
	UserID  string
	GroupID string

	TaskName    string
	Description string
	MeetingDate string

	PatternID string

	FilePath string
	FileName string
	Duration int

	Status string

	ResultJson json.RawMessage

	CurrentInputKey  string
    CurrentOutputKey string

	StageEnteredAt time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	StartedAt      time.Time
	CompletedAt    time.Time

	ChangeFlag bool // может ли изменять callerID данную задачу
}

type TaskCheck struct {
	Status                     string
	IsProcess                  bool
	InTheQueueBefore           int
	ApproximateLeadTimeProcess int
}

// содержит prompt и additional_prompt, полученные по задаче.
type TaskPatternPrompts struct {
    Prompt           string
    AdditionalPrompt json.RawMessage
}


// создаем конечный автомат жизненного цикла задачи
type TaskStatus string

const (
	StatusProcessingUpload     TaskStatus = "processing_upload" // начальный статус, до полной загрузки в s3
	StatusPendingDenoise       TaskStatus = "pending_denoise"
	StatusProcessingDenoise    TaskStatus = "processing_denoise"
	StatusPendingDiarize       TaskStatus = "pending_diarize"
	StatusProcessingDiarize    TaskStatus = "processing_diarize"
	StatusPendingTranscribe    TaskStatus = "pending_transcribe"
	StatusProcessingTranscribe TaskStatus = "processing_transcribe"
	StatusPendingSummarize     TaskStatus = "pending_summarize"
	StatusProcessingSummarize  TaskStatus = "processing_summarize"
	StatusDone                 TaskStatus = "done"

	StatusErrorUpload     TaskStatus = "error_upload"
	StatusErrorDenoise    TaskStatus = "error_denoise"
	StatusErrorTranscribe TaskStatus = "error_transcribe"
	StatusErrorDiarize    TaskStatus = "error_diarize"
	StatusErrorSummarize  TaskStatus = "error_summarize"
)

// возвращает следующий статус в конвейере (или "" для конечных)
func (s TaskStatus) NextStatus() TaskStatus {
	switch s {
	case StatusProcessingUpload:
		return StatusPendingDenoise
	case StatusPendingDenoise:
		return StatusProcessingDenoise
	case StatusProcessingDenoise:
		return StatusPendingDiarize
	case StatusPendingDiarize:
		return StatusProcessingDiarize
	case StatusProcessingDiarize:
		return StatusPendingTranscribe
	case StatusPendingTranscribe:
		return StatusProcessingTranscribe
	case StatusProcessingTranscribe:
		return StatusPendingSummarize
	case StatusPendingSummarize:
		return StatusProcessingSummarize
	case StatusProcessingSummarize:
		return StatusDone
	default:
		return ""
	}
}

// возвращает true, если статус означает активную обработку
func (s TaskStatus) IsProcessing() bool {
	return s == StatusProcessingUpload ||
		s == StatusProcessingDenoise ||
		s == StatusProcessingTranscribe ||
		s == StatusProcessingDiarize ||
		s == StatusProcessingSummarize
}

// возвращает true, если статус означает ожидание в очереди
func (s TaskStatus) IsPending() bool {
	return s == StatusPendingDenoise ||
		s == StatusPendingTranscribe ||
		s == StatusPendingDiarize ||
		s == StatusPendingSummarize
}
