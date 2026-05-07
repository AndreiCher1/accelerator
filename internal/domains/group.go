package domains

import "time"

type Group struct {
	GroupID     string
	Name        string
	Description string 
	CreatedBy   string
	CreatedAt   time.Time
}
