package postgres

import (
	"admin_service/internal/domain/entities"
	"database/sql"
)

type AppRepo struct {
	db *sql.DB
}

func NewAppRepo(db *sql.DB) *AppRepo {
	return &AppRepo{db}
}

func (r *AppRepo) Create(app *entities.App) error {
	_,err := r.db.Exec(`
		INSERT INTO apps(app_name,app_email,status,tier)
		VALUES($1,$2,$3,$4)
	`,app.AppName,app.AppEmail,app.Status,app.Tier)

	return err 
}

func (r *AppRepo) FindAll()([]*entities.App,error) {
	rows,err := r.db.Query(`
		SELECT app_id,app_name,app_email,status,tier, created_at, updated_at FROM apps ORDER BY created_at DESC
	`)
	if err != nil {
		return nil,err 
	}
	defer rows.Close()

	var apps []*entities.App
	for rows.Next() {
		var a entities.App
		err := rows.Scan(
			&a.AppID,
			&a.AppName,
			&a.AppEmail,
			&a.Status,
			&a.Tier,
			&a.CreatedAt,
			&a.UpdatedAt,
		)
		if err != nil {
			return nil,err 
		}
		apps = append(apps, &a)
	}
	return apps,nil 
}

func(r *AppRepo) UpdateStatus(appID, status string) error {
	_,err := r.db.Exec(`
		UPDATE apps
		SET status=$1
		WHERE app_id=$2
	`,status,appID)
	return err 
}

func (r *AppRepo)CreateFirst(user *entities.AppUser) error {

	_, err := r.db.Exec(`
		INSERT INTO app_users
		(app_id, email, password_hash, role, status)
		VALUES ($1,$2,$3,$4,$5)
	`, user.AppID, user.Email, user.PasswordHash, user.Role, user.Status)

	if err != nil {
		return err 
	}
	
	return nil 
}