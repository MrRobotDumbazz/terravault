package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/terravault/oracle/internal/api"
	"github.com/terravault/oracle/internal/storage"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Fprintln(os.Stderr, "No .env file found, using system environment")
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	defer logger.Sync()

	// Initialize database
	dbURL := mustEnv("DATABASE_URL")
	db, err := storage.NewPostgres(dbURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations("internal/storage/migrations"); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	// Initialize IPFS client
	ipfsURL := getEnvOrDefault("IPFS_API_URL", "/ip4/127.0.0.1/tcp/5001")
	ipfsClient := storage.NewIPFSClient(ipfsURL)

	// Build router
	jwtSecret := mustEnv("JWT_SECRET")
	apiKeyInternal := mustEnv("INTERNAL_API_KEY")
	solanaRPCURL := getEnvOrDefault("SOLANA_RPC_URL", "https://api.devnet.solana.com")

	// ADMIN_WALLETS is a comma-separated list of Solana wallet addresses granted admin role.
	adminWallets := make(map[string]struct{})
	if raw := os.Getenv("ADMIN_WALLETS"); raw != "" {
		for _, w := range strings.Split(raw, ",") {
			if w = strings.TrimSpace(w); w != "" {
				adminWallets[w] = struct{}{}
			}
		}
	}

	router := api.NewRouter(api.Config{
		DB:             db,
		IPFS:           ipfsClient,
		Logger:         logger,
		JWTSecret:      []byte(jwtSecret),
		InternalAPIKey: apiKeyInternal,
		SolanaRPCURL:   solanaRPCURL,
		AdminWallets:   adminWallets,
	})

	port := getEnvOrDefault("API_PORT", "8080")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("TerraVault API server starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-stop
	logger.Info("shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
	}
	logger.Info("server stopped")
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return v
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
