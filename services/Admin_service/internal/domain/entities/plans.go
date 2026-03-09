package entities

import "time"

type Plan struct {
	PlanID          string
	Name            string
	MonthlyJobLimit int
	Price           float64
	CreatedAt       time.Time
}