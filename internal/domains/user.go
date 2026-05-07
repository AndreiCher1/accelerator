package domains

import "time"

type User struct {
	ID            string
	Login         string
	FullName      string
	Position      string
	Role          string
	Password_hash string
	CreatedAt     time.Time
}
