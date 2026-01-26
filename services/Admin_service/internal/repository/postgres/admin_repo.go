package postgres

import (
	"admin_service/internal/domain/entities"
	"database/sql"
)

type AdminRepo struct {
	db *sql.DB
}

func NewAdminRepo(db *sql.DB) *AdminRepo {
	return &AdminRepo{db}
}

func (r *AdminRepo) FindByEmail(email string) (*entities.PlatformAdmin,error) {
	row := r.db.QueryRow(`
		SELECT id,email,password_hash,status,created_at
		FROM platform_admins WHERE email = $1`,email)
	
	var padmin entities.PlatformAdmin
	err := row.Scan(&padmin.ID,&padmin.Email,&padmin.PasswordHash,&padmin.Status,&padmin.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil,nil 
		}
		return nil,err 
	}
	return &padmin,nil 
}

func (r *AdminRepo) Create(admin *entities.PlatformAdmin) error {
	_,err := r.db.Exec(`
		INSERT INTO platform_admins(email,password_hash,status)
		VALUES($1,$2,$3)
	`,admin.Email,admin.PasswordHash,admin.Status)

	return err 
}