package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/terravault/oracle/internal/listener"
	"github.com/terravault/oracle/internal/oracle"
	"github.com/terravault/oracle/internal/storage"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Fprintln(os.Stderr, "No .env file found, using system environment")
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	defer logger.Sync()

	dbURL := mustEnv("DATABASE_URL")
	db, err := storage.NewPostgres(dbURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Load oracle signing keypair from file or env
	keypairPath := getEnvOrDefault("ORACLE_KEYPAIR_PATH", "oracle-keypair.json")
	signer, err := oracle.NewSignerFromFile(keypairPath)
	if err != nil {
		logger.Fatal("failed to load oracle keypair", zap.Error(err))
	}
	logger.Info("oracle pubkey", zap.String("pubkey", signer.PublicKey().String()))

	programID := mustEnv("TERRAVAULT_PROGRAM_ID")
	solanaWSURL := getEnvOrDefault("SOLANA_WS_URL", "wss://api.devnet.solana.com")
	solanaRPCURL := getEnvOrDefault("SOLANA_RPC_URL", "https://api.devnet.solana.com")

	milestoneHandler := oracle.NewMilestoneHandler(oracle.MilestoneConfig{
		DB:           db,
		Signer:       signer,
		ProgramID:    programID,
		SolanaRPCURL: solanaRPCURL,
		Logger:       logger,
	})

	ws := listener.NewWebSocketListener(listener.WSConfig{
		WSURL:     solanaWSURL,
		ProgramID: programID,
		DB:        db,
		Logger:    logger,
		Handlers: listener.Handlers{
			OnMilestoneSubmitted: milestoneHandler.HandleMilestoneSubmitted,
			OnFundraisingStarted: milestoneHandler.HandleFundraisingStarted,
			OnProjectActivated:   milestoneHandler.HandleProjectActivated,
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("starting Solana program log listener", zap.String("program", programID))
		if err := ws.Listen(ctx); err != nil {
			logger.Error("listener error", zap.Error(err))
			cancel()
		}
	}()

	<-stop
	logger.Info("shutting down listener gracefully...")
	cancel()
	logger.Info("listener stopped")
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
