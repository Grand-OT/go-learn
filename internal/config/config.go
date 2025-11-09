package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RepoType   string
	Port       string
	StaticDir  string
	DBHost     string
	DBPort     uint16
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("error during env file loading: %s", err)
	}

	cfg := Config{
		RepoType:  getEnv("REPO_TYPE", "postgres"),
		Port:      getEnv("HTTP_PORT", "80"),
		StaticDir: getEnv("STATIC_DIR", "static"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "app"),
		DBPassword: getEnv("DB_PASSWORD", "app"),
		DBName:     getEnv("DB_NAME", "app"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	dbPortStr := getEnv("DB_PORT", "5432")
	if p, err := strconv.Atoi(dbPortStr); err == nil && p > 0 && p < 65536 {
		cfg.DBPort = uint16(p)
	} else {
		// fallback по умолчанию
		cfg.DBPort = 5432
	}

	return cfg
}

func getEnv(key, def string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return def
}
