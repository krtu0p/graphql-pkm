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
	_ = godotenv.Load()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	mysqlDB, err := database.NewMySQLDB(cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer mysqlDB.Close()

	log.Println("✓ MySQL database connected successfully")

	embeddingsCache := database.NewMySQLEmbeddingsCache(mysqlDB)
	aiClient := ai.NewClient(cfg.ApiKey, cfg.ApiUrl, cfg.DefaultModel, cfg.JinaToken)

	
	noteService := service.NewNoteService(mysqlDB, aiClient)
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