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
		ApiUrl:      getEnv("apiUrl", "https://openrouter.ai/api/v1"),
		DefaultModel: getEnv("defaultModel", "deepseek/deepseek-coder:33b-instruct"),
		DatabaseUrl: getEnv("dbUrl", "root:password@tcp(localhost:3306)/pkm?parseTime=true"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}