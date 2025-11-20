package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	Environment string
	apiKey      string
	apiUrl      string
	defaultModel string
}

func LoadConfig() (*Config, error) {
	portStr := getEnv("PORT", "8080")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:        port,
		Environment: getEnv("ENVIRONMENT", "development"),
		apiKey:      getEnv("API_KEY", ""),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
