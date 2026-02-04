package repository

import (
	"github.com/jackc/pgx"
)

type AdminRepo struct {
	db *pgx.Conn
}

func NewAdminDb(dsn string) *AdminRepo {
	cfg,err := pgx.ParseConnectionString(dsn)
	if err != nil {
		panic(err)
	}
	db,_ := pgx.Connect(cfg)
	return &AdminRepo{db: db}
}

func (r *AdminRepo) Concurrency(appID string)(int,error) {
	var limit int 
	err := r.db.QueryRow(
		"SELECT limit FROM apps WHERE id = $1",appID,
	).Scan(&limit)
	return limit,err 
}

// Need to correct logic here for better limit check