package entities

import "time"

type InvoiceData struct {
	UserID    string 
	UserName  string
	UserEmail string

	PlanID 		string 
	PlanName   	string
	PlanAmount 	float64

	InvoiceNumber string
	TotalPaid     float64
	CreatedDate   time.Time
	NextPayment   time.Time
}