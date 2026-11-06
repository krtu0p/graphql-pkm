package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"graphql-pkm/config"
	"graphql-pkm/internal/ai"
	"graphql-pkm/internal/database"
	"graphql-pkm/internal/gql/generated"
	"graphql-pkm/internal/gql/resolvers"
	"graphql-pkm/internal/service"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var (
		db              database.Database
		embeddingsCache database.EmbeddingsCache
		closeDB         func() error
		dbDesc          string
	)

	switch cfg.DBDriver {
	case "sqlite", "":
		sqliteDB, err := database.NewSQLiteDB(cfg.DBPath)
		if err != nil {
			log.Fatalf("Failed to open SQLite database: %v", err)
		}
		db = sqliteDB
		embeddingsCache = database.NewSQLiteEmbeddingsCache(sqliteDB, cfg.EmbedModel)
		closeDB = sqliteDB.Close
		dbDesc = fmt.Sprintf("SQLite (%s)", cfg.DBPath)
	case "mysql":
		mysqlDB, err := database.NewMySQLDB(cfg.DatabaseUrl)
		if err != nil {
			log.Fatalf("Failed to connect to MySQL: %v", err)
		}
		db = mysqlDB
		embeddingsCache = database.NewMySQLEmbeddingsCache(mysqlDB, cfg.EmbedModel)
		closeDB = mysqlDB.Close
		dbDesc = "MySQL (persistent storage)"
		log.Println("✓ MySQL database connected successfully")
	default:
		log.Fatalf("Unsupported DB_DRIVER %q (expected \"sqlite\" or \"mysql\")", cfg.DBDriver)
	}
	defer closeDB()

	aiClient := ai.NewClientWithConfig(ai.ClientConfig{
		Enabled:      cfg.AIEnabled,
		BaseURL:      cfg.AIBaseURL,
		APIKey:       cfg.AIAPIKey,
		Model:        cfg.AIModel,
		EmbedBaseURL: cfg.EmbedBaseURL,
		EmbedAPIKey:  cfg.EmbedAPIKey,
		EmbedModel:   cfg.EmbedModel,
	})

	noteService := service.NewNoteService(db, aiClient)
	aiService := service.NewAIService(aiClient, embeddingsCache, noteService)
	aiService.SetSemanticThreshold(cfg.SemanticThreshold)
	searchService := service.NewSearchService(noteService, aiService)
	resolver := resolvers.NewResolver(noteService, searchService)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	http.Handle("/query", srv)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("🚀 GraphQL server running on http://%s", addr)
	log.Printf("🎮 Playground at http://%s", addr)

	if aiClient.IsEnabled() {
		log.Printf("🤖 AI features: ENABLED (chat: %s @ %s, embed: %s)", cfg.AIModel, cfg.AIBaseURL, cfg.EmbedModel)
	} else {
		log.Println("⚠️  AI features: DISABLED")
	}

	log.Printf("💾 Database: %s", dbDesc)
	log.Println("=" + strings.Repeat("=", 50))

	log.Fatal(http.ListenAndServe(addr, nil))
}
