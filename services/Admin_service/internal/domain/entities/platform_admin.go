package entities

import "time"

type PlatformAdmin struct {
	ID 				string 
	Email			string 
	PasswordHash	string 
	Status 			string 
	CreatedAt		time.Time 
}