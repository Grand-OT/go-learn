package config

import "os"

type Config struct {
	Port      string
	StaticDir string
}

func Load() Config {
	cfg := Config{
		Port:      getEnv("HTTP_PORT", "80"),
		StaticDir: getEnv("STATIC_DIR", "static"),
	}
	return cfg
}

func getEnv(key, def string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return def
}
