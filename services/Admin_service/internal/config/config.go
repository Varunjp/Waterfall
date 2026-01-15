package config

type Config struct {
	GrpcPort string
	DBUrl    string
	JWTKey   string
}

func Load() *Config {
	return &Config{
		GrpcPort: GetEnv("GRPC_PORT", "50051"),
		DBUrl: "postgres://" +
			GetEnv("DB_USER", "") + ":" +
			GetEnv("DB_PASSWORD", "") + "@" +
			GetEnv("DB_HOST", "") + ":" +
			GetEnv("DB_PORT", "") + "/" +
			GetEnv("DB_NAME", "") + "?sslmode=disable",
		JWTKey: GetEnv("JWT_SECRET", ""),
	}
}