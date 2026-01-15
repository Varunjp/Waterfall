package entities

import "time"

type AppUser struct {
	ID           string
	AppID        string
	Email        string
	PasswordHash string
	Role         string
	Status       string
	CreatedAt    time.Time
}