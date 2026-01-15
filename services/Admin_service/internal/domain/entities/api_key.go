package entities

import "time"

type ApiKey struct {
	ID        string
	AppID     string
	KeyHash   string
	Scopes    []string
	CreatedAt time.Time
	RevokedAt *time.Time
}
