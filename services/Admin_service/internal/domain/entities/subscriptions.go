package entities

import "time"

type Subscription struct {
	AppID                string
	PlanID               string
	PlanName             string
	PlanLimit            int
	FreeLimit            int
	FreeUsage            int
	Status               string
	CurrentLimit         int
	StripeSubscriptionID string
	CurrentPeriodStart   time.Time
	CurrentPeriodEnd     time.Time
	CreatedAt            time.Time
}
