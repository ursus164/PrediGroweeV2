package main

import (
	"admin/clients"
	"admin/internal/api"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"log"
	"os"
)

func main() {
	// Initialize logger
	var err error
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			log.Fatalf("Failed to sync logger: %v", err)
		}
	}(logger)

	authClient := clients.NewRestAuthClient("http://auth:8080/auth", os.Getenv("INTERNAL_API_KEY"), logger)
	statsClient := clients.NewStatsRestClient("http://stats:8080/stats", os.Getenv("INTERNAL_API_KEY"), logger)
	quizClient := clients.NewQuizRestClient("http://quiz:8080/quiz", os.Getenv("INTERNAL_API_KEY"), logger)
	apiServer := api.NewApiServer(":8080", logger, authClient, statsClient, quizClient)
	apiServer.Run()
}
