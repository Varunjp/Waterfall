package repository

import (
	"github.com/jackc/pgx"
)

type AdminRepo struct {
	db *pgx.Conn
}

func NewAdminDb(dsn string) *AdminRepo {
	cfg, err := pgx.ParseConnectionString(dsn)
	if err != nil {
		panic(err)
	}
	db, _ := pgx.Connect(cfg)
	return &AdminRepo{db: db}
}
