package entities

import "time"

type Subscriber struct {
	AppID     string
	AppName   string
	PlanID    string
	PlanName  string
	Status    string
	StartDate time.Time
	EndDate   time.Time
}
