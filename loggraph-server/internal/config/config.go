package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
	ServerPort  string
	S3          S3Config
}

type S3Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
}

func Load() *Config {
	return &Config{
		DatabaseURL: envOr("DATABASE_URL", "postgres://loggraph:loggraph@localhost:5432/loggraph?sslmode=disable"),
		ServerPort:  envOr("SERVER_PORT", "8080"),
		S3: S3Config{
			Endpoint:  envOr("S3_ENDPOINT", "http://localhost:9000"),
			Bucket:    envOr("S3_BUCKET", "loggraph"),
			AccessKey: envOr("S3_ACCESS_KEY", "minioadmin"),
			SecretKey: envOr("S3_SECRET_KEY", "minioadmin"),
			Region:    envOr("S3_REGION", "us-east-1"),
		},
	}
}

func (c *Config) Addr() string {
	return fmt.Sprintf(":%s", c.ServerPort)
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
