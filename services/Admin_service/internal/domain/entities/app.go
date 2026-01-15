package entities

import "time"

type App struct {
	AppID 		string 
	AppName		string 
	AppEmail	string 
	Status 		string 
	Tier 		string 
	CreatedAt   time.Time 
	UpdatedAt 	time.Time 
}