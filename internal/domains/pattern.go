package domains

import "time"

type Pattern struct {
	ID               string
	Name             string
	Description      string
	SummaryPrompt    string
	AdditionalPrompt string
	CreatedAt        time.Time
}
