package postgres

import (
	"admin_service/internal/domain/entities"
	"database/sql"
)

type ApiKeyRepo struct {
	db *sql.DB
}

func NewApiKeyRepo(db *sql.DB) *ApiKeyRepo {
	return &ApiKeyRepo{db}
}

func (r *ApiKeyRepo) Create(k *entities.ApiKey) error {
	_, err := r.db.Exec(`
		INSERT INTO api_keys (app_id, key_hash, scopes)
		VALUES ($1,$2,$3)
	`, k.AppID, k.KeyHash, k.Scopes)
	return err
}

func (r *ApiKeyRepo) Revoke(id string) error {
	_, err := r.db.Exec(`
		UPDATE api_keys SET revoked_at=now() WHERE id=$1
	`, id)
	return err
}
