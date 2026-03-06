package config

import (
	"admin_service/internal/domain/entities"
	"admin_service/internal/infrastructure/security"
	repo "admin_service/internal/repository/interfaces"
	"strconv"
	"strings"
)

type Config struct {
	GrpcPort string
	DBUrl    string
	JWTKey   string
	RedisAddr string 
	RedisPassword string 
	RedisDB  int 
	SmtpHost string 
	SmtpPort string 
	SmtpUser string 
	Smtppass string 
}

func Load() *Config {
	redisDB, _ := strconv.Atoi(GetEnv("REDIS_DB", "0"))
	return &Config{
		GrpcPort: GetEnv("GRPC_PORT", "50051"),
		DBUrl: "postgres://" +
			GetEnv("DB_USER", "") + ":" +
			GetEnv("DB_PASSWORD", "") + "@" +
			GetEnv("DB_HOST", "") + ":" +
			GetEnv("DB_PORT", "") + "/" +
			GetEnv("DB_NAME", "") + "?sslmode=disable",
		JWTKey: GetEnv("JWT_SECRET", ""),
		RedisAddr: GetEnv("REDIS_ADDR",""),
		RedisPassword: GetEnv("REDIS_PASSWORD",""),
		RedisDB: redisDB,
		SmtpHost: GetEnv("SMTP_HOST",""),
		SmtpPort: GetEnv("SMTP_PORT",""),
		SmtpUser: GetEnv("SMTP_USER",""),
		Smtppass: GetEnv("SMTP_PASS",""),
	}
}

func BootstrapAdmin(db repo.AdminRepository) error {

	email := GetEnv("PLATFORM_ADMIN_EMAIL","")
	pass := GetEnv("PLATFORM_ADMIN_PASSWORD","")

	email = strings.ToLower(strings.TrimSpace(email))

	existing,err := db.FindByEmail(email)
	if err != nil {
		return err 
	}

	if existing != nil {
		return nil 
	}

	hash,err := security.Hash(pass)
	if err != nil {
		return err 
	}

	admin := &entities.PlatformAdmin{
		Email: email,
		PasswordHash: hash,
		Status: "active",
	}

	return db.Create(admin)
}