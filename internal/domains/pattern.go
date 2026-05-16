package domains

import (
	"encoding/json"
	"time"
)

type GroupWithPatterns struct {
	GroupID     string
    Name        string
    Description string
    Patterns    []Pattern
}

type Pattern struct {
	ID               string
	GroupID          string
	Name             string
	Description      string
	SummaryPrompt    string
	AdditionalPrompt json.RawMessage
	CreatedBy        string
	CreatedAt        time.Time
	ChangeFlag       bool // чтобы понимать, какой из шаблонов может редактировать админ
}
