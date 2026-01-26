package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgres(dsn string) *sql.DB {
	db,err := sql.Open("postgres",dsn)
	if err != nil  {
		log.Fatalf("failed to open db: %v",err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v",err)
	}
	return db 
}