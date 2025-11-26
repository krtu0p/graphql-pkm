package main

import (
	"log"
	"net/http"
	"strconv"
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
	// Load environment variables
	_ = godotenv.Load()

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MySQL database
	mysqlDB, err := database.NewMySQLDB(cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer mysqlDB.Close()

	log.Println("✓ MySQL database connected successfully")

	// Initialize dependencies - USE MYSQL INSTEAD OF MEMORY!
	embeddingsCache := database.NewMySQLEmbeddingsCache(mysqlDB)
	aiClient := ai.NewClient(cfg.ApiKey, cfg.ApiUrl, cfg.DefaultModel)
	
	// CRITICAL FIX: Pass mysqlDB instead of creating a new MemoryDB
	noteService := service.NewNoteService(mysqlDB)
	aiService := service.NewAIService(aiClient, embeddingsCache, noteService)
	searchService := service.NewSearchService(noteService, aiService)
	resolver := resolvers.NewResolver(noteService, searchService)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("🚀 GraphQL server running on http://localhost:%d", cfg.Port)
	log.Printf("🎮 Playground at http://localhost:%d", cfg.Port)

	if aiClient.IsEnabled() {
		log.Printf("🤖 AI features: ENABLED (using %s)", cfg.DefaultModel)
	} else {
		log.Println("⚠️  AI features: DISABLED (no API key)")
	}
	
	log.Println("💾 Database: MySQL (persistent storage)")
	log.Println("=" + strings.Repeat("=", 50))

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Port), nil))
}