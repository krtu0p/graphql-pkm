package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

// Shutdown budget: HTTP drain, then AI enrichment flush, before the deferred
// database close. If this server is ever containerized, "docker stop"
// defaults to ~10s before SIGKILL — the flush could be killed mid-drain
// (note text is already saved; only metadata generation would be lost).
// Either raise stop_grace_period above the sum of these constants or
// shorten them.
const (
	shutdownHTTPGraceTimeout  = 15 * time.Second
	aiEnrichmentFlushTimeout = 15 * time.Second
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

	settingsStore, ok := db.(database.AppSettingsStore)
	if !ok {
		log.Fatalf("Database backend does not support app settings")
	}
	settingsService := service.NewAISettingsService(settingsStore, aiClient, service.AISettingsDefaults{
		Enabled:      cfg.AIEnabled,
		Provider:     cfg.AIProvider,
		BaseURL:      cfg.AIBaseURL,
		APIKey:       cfg.AIAPIKey,
		Model:        cfg.AIModel,
		EmbedBaseURL: cfg.EmbedBaseURL,
		EmbedAPIKey:  cfg.EmbedAPIKey,
		EmbedModel:   cfg.EmbedModel,
	})
	if err := settingsService.Reload(); err != nil {
		log.Printf("⚠️  Could not load saved AI settings: %v (using environment configuration)", err)
	}

	resolver := resolvers.NewResolver(noteService, searchService, settingsService)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	http.Handle("/query", srv)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("🚀 GraphQL server running on http://%s", addr)
	log.Printf("🎮 Playground at http://%s", addr)

	if aiClient.IsEnabled() {
		if aiSettings, err := settingsService.GetAISettings(); err == nil {
			log.Printf("🤖 AI features: ENABLED (chat: %s @ %s, embed: %s)", aiSettings.Model, aiSettings.BaseURL, aiSettings.EmbedModel)
		} else {
			log.Println("🤖 AI features: ENABLED")
		}
	} else {
		log.Println("⚠️  AI features: DISABLED")
	}

	log.Printf("💾 Database: %s", dbDesc)
	log.Println("=" + strings.Repeat("=", 50))

	// Graceful shutdown: on SIGINT/SIGTERM stop accepting connections, then
	// drain pending AI enrichment before the deferred database close runs.
	httpServer := &http.Server{Addr: addr}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("Shutting down server (waiting for in-flight requests)...")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownHTTPGraceTimeout)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("Warning: HTTP server shutdown: %v", err)
		}
	}()

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server error: %v", err)
	}

	log.Println("Server stopped; draining pending AI enrichment")
	noteService.FlushPendingEnrichment(aiEnrichmentFlushTimeout)
}
