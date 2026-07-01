package entities

import (
	"time"
)

type Payment struct {
	InvoiceID      string          `db:"invoice_id"`
	SubscriptionID string          `db:"subscription_id"`
	AppID          string          `db:"app_id"`
	Amount         int64 `db:"amount"`
	Currency       string          `db:"currency"`
	CustomerEmail  string          `db:"customer_email"`
	Status         string          `db:"status"`
	PaidAt         time.Time       `db:"paid_at"`
	CreateAt       time.Time       `db:"created_at"`
}
