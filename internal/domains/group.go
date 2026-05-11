package domains

import (
	"time"
)

type Group struct {
	GroupID     string
	Name        string
	Description string
	MemberCount int
	CreatedBy   string
	OwnerID     string
	CreatedAt   time.Time
	CanEdit     bool
	CanDelete   bool
}
