package entities

import "time"

type App struct {
	AppID 		string 
	AppName		string 
	AppEmail	string 
	Status 		string 
	Tier 		string 
	PlanID 		string 
	CreatedAt   time.Time 
	UpdatedAt 	time.Time 
}

type AppDetails struct {
	AppID 		string 
	AppName		string 
	AppEmail	string 
	Status 		string 
	PlanName 	string 
	EndDate		time.Time
}