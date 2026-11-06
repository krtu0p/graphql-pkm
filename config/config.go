package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Host         string
	Port         int
	Environment  string
	DBDriver     string
	DBPath       string
	DatabaseUrl  string
	ApiKey       string
	ApiUrl       string
	DefaultModel string
	JinaToken    string
}

func LoadConfig() (*Config, error) {
	portStr := getEnvAny([]string{"PORT", "Port"}, "8080")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	return &Config{
		Host:         getEnv("HOST", "127.0.0.1"),
		Port:         port,
		Environment:  getEnv("ENVIRONMENT", "development"),
		DBDriver:     strings.ToLower(getEnv("DB_DRIVER", "sqlite")),
		DBPath:       getEnv("DB_PATH", "./data/pkm.db"),
		DatabaseUrl:  getEnvAny([]string{"DATABASE_URL", "dbUrl"}, "root:password@tcp(localhost:3306)/pkm?parseTime=true"),
		ApiKey:       getEnv("apiKey", ""),
		ApiUrl:       getEnv("apiUrl", "https://api.groq.com/openai/v1"),
		DefaultModel: getEnv("defaultModel", "llama-3.1-8b-instant"),
		JinaToken:    getEnv("jinaToken", ""),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAny(keys []string, defaultValue string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return defaultValue
}
