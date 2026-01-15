package entities

import "time"

type AuditLog struct {
	ID        string
	ActorType string
	ActorID   string
	Action    string
	Resource  string
	Metadata  map[string]interface{}
	CreatedAt time.Time
}
