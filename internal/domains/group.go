package domains

import "time"

type Group struct {
	GroupID     string
	Name        string
	Description string
	MemberCount int
	CreatedBy   string
	CreatedAt   time.Time
}
