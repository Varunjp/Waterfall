package postgres

import (
	"admin_service/internal/domain/entities"
	"database/sql"
)

type AppUserRepo struct {
	db *sql.DB
}

func NewAppUserRepo(db *sql.DB) *AppUserRepo {
	return &AppUserRepo{db}
}

func (r *AppUserRepo) Create(u *entities.AppUser) error {
	_, err := r.db.Exec(`
		INSERT INTO app_users
		(app_id, email, password_hash, role, status)
		VALUES ($1,$2,$3,$4,$5)
	`, u.AppID, u.Email, u.PasswordHash, u.Role, u.Status)

	return err
}

func (r *AppUserRepo) FindByApp(appID string) ([]*entities.AppUser, error) {
	rows, err := r.db.Query(`
		SELECT id, app_id, email, role, status, created_at
		FROM app_users
		WHERE app_id=$1
	`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entities.AppUser
	for rows.Next() {
		var u entities.AppUser
		err := rows.Scan(
			&u.ID,
			&u.AppID,
			&u.Email,
			&u.Role,
			&u.Status,
			&u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

func (r *AppUserRepo) FindByEmail(email string) (*entities.AppUser,error) {
	row := r.db.QueryRow(
		`SELECT id,app_id,email,password_hash,role,status,created_at
		FROM app_users WHERE email = $1`,email,
	)

	var appUser entities.AppUser
	err := row.Scan(&appUser.ID,&appUser.AppID,&appUser.Email,&appUser.PasswordHash,&appUser.Role,&appUser.Status,&appUser.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil,err 
		}
		return nil,err 
	}
	return &appUser,nil 
}
