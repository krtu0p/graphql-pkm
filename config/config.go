package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Host              string
	Port              int
	Environment       string
	DBDriver          string
	DBPath            string
	DatabaseUrl       string
	AIEnabled         bool
	AIProvider        string
	AIBaseURL         string
	AIAPIKey          string
	AIModel           string
	EmbedBaseURL      string
	EmbedAPIKey       string
	EmbedModel        string
	SemanticThreshold float64

	// Legacy fields kept for compatibility
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

	aiBaseURL := getEnvAny([]string{"AI_BASE_URL", "apiUrl"}, "http://localhost:11434/v1")
	aiAPIKey := getEnvAny([]string{"AI_API_KEY", "apiKey"}, "")
	aiModel := getEnvAny([]string{"AI_MODEL", "defaultModel"}, "llama3.2")
	embedBaseURL := getEnv("EMBED_BASE_URL", aiBaseURL)
	embedAPIKey := getEnvAny([]string{"EMBED_API_KEY", "jinaToken"}, aiAPIKey)
	embedModel := getEnv("EMBED_MODEL", "nomic-embed-text")

	aiEnabled := aiBaseURL != ""
	if rawEnabled := strings.TrimSpace(os.Getenv("AI_ENABLED")); rawEnabled != "" {
		parsed, err := strconv.ParseBool(rawEnabled)
		if err != nil {
			return nil, err
		}
		aiEnabled = parsed && aiBaseURL != ""
	}

	semanticThreshold := 0.6
	if rawThresh := strings.TrimSpace(os.Getenv("SEMANTIC_THRESHOLD")); rawThresh != "" {
		parsed, err := strconv.ParseFloat(rawThresh, 64)
		if err != nil {
			return nil, err
		}
		semanticThreshold = parsed
	}

	return &Config{
		Host:              getEnv("HOST", "127.0.0.1"),
		Port:              port,
		Environment:       getEnv("ENVIRONMENT", "development"),
		DBDriver:          strings.ToLower(getEnv("DB_DRIVER", "sqlite")),
		DBPath:            getEnv("DB_PATH", "./data/pkm.db"),
		DatabaseUrl:       getEnvAny([]string{"DATABASE_URL", "dbUrl"}, "root:password@tcp(localhost:3306)/pkm?parseTime=true"),
		AIEnabled:         aiEnabled,
		AIProvider:        strings.ToLower(strings.TrimSpace(getEnv("AI_PROVIDER", "ollama"))),
		AIBaseURL:         aiBaseURL,
		AIAPIKey:          aiAPIKey,
		AIModel:           aiModel,
		EmbedBaseURL:      embedBaseURL,
		EmbedAPIKey:       embedAPIKey,
		EmbedModel:        embedModel,
		SemanticThreshold: semanticThreshold,
		ApiKey:            aiAPIKey,
		ApiUrl:            aiBaseURL,
		DefaultModel:      aiModel,
		JinaToken:         embedAPIKey,
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
