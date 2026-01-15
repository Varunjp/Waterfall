package postgres

import (
	"admin_service/internal/domain/entities"
	"database/sql"
)

type EmailRepo struct {
	db *sql.DB
}

func NewEmailRepo(db *sql.DB) *EmailRepo {
	return &EmailRepo{db}
}

func (r *EmailRepo) Upsert(e *entities.EmailConfig) error {
	_, err := r.db.Exec(`
		INSERT INTO email_configs (app_id, provider, from_email)
		VALUES ($1,$2,$3)
		ON CONFLICT (app_id)
		DO UPDATE SET provider=$3, from_email=$4
	`, e.AppID, e.Provider, e.FromEmail)
	return err
}
