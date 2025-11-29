package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	Environment string
	ApiKey      string
	ApiUrl      string
	DefaultModel string
	DatabaseUrl string
	HfToken string
}

func LoadConfig() (*Config, error) {
	portStr := getEnv("Port", "8080")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:        port,
		Environment: getEnv("ENVIRONMENT", "development"),
		ApiKey:      getEnv("apiKey", ""),
		ApiUrl:      getEnv("apiUrl", "https://api.groq.com/openai/v1"),
		DefaultModel: getEnv("defaultModel", "llama-3.1-8b-instant"),
		DatabaseUrl: getEnv("dbUrl", "root:password@tcp(localhost:3306)/pkm?parseTime=true"),
		HfToken: getEnv("hfToken", ""),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}