package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresDB DBConfig
	MongoDB    DBConfig
}

type DBConfig struct {
	User     string
	Password string
	Name     string
	Host     string
	Port     string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	return &Config{
		PostgresDB: DBConfig{
			User:     getEnv("PG_USER", ""),
			Password: getEnv("PG_PASS", ""),
			Name:     getEnv("PG_NAME", ""),
			Host:     getEnv("PG_HOST", ""),
			Port:     getEnv("PG_PORT", ""),
		},
		MongoDB: DBConfig{
			User:     getEnv("MONGO_USER", ""),
			Password: getEnv("MONGO_PASS", ""),
			Name:     getEnv("MONGO_NAME", ""),
			Host:     getEnv("MONGO_HOST", ""),
			Port:     getEnv("MONGO_PORT", ""),
		},
	}, nil
}

func getEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
