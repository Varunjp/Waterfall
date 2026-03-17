package entities

import "time"

type Subscription struct {
	AppID                string
	PlanID               string
	Status               string
	StripeSubscriptionID string
	CurrentPeriodStart   time.Time
	CurrentPeriodEnd     time.Time 
	CreatedAt 			 time.Time 
}