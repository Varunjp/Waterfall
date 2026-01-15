package postgres

import (
	"admin_service/internal/domain/entities"
	"database/sql"
	"encoding/json"
)

type AuditRepo struct {
	db *sql.DB
}

func NewAuditRepo(db *sql.DB) *AuditRepo {
	return &AuditRepo{db}
}

func (r *AuditRepo) Log(a *entities.AuditLog) error {
	meta, _ := json.Marshal(a.Metadata)
	_, err := r.db.Exec(`
		INSERT INTO audit_logs
		(actor_type, actor_id, action, resource, metadata)
		VALUES ($1,$2,$3,$4,$5)
	`, a.ActorType, a.ActorID, a.Action, a.Resource, meta)
	return err
}
