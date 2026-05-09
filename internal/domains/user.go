package domains

import "time"

type User struct {
	ID                string
	Login             string
	FullName          string
	Position          string
	Role              string
	TemporaryPassword bool // нужно только для входа, взаимодействуем только там
	CreatedAt         time.Time
}
