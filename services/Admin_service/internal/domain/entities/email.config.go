package entities

import "time"

type EmailConfig struct {
	ID        string
	AppID     string
	Provider  string
	FromEmail string
	CreatedAt time.Time
}
