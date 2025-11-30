// Command images is the entrypoint for the Images microservice.
package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"images/internal/api"
	"images/internal/clients"
	"log"
	"os"
	"time"
)

const PingDbAttempts = 3

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

	// Initialize database
	db, err := connectToPostgres()
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatal("Failed to close database connection: %v", zap.Error(err))
		}
	}(db)

	// Set up database connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify database connection
	for i := 1; i <= PingDbAttempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = db.PingContext(ctx)
		cancel()
		if err == nil {
			break
		}
		logger.Error(fmt.Sprintf("Failed to ping the database (attempt: %d/%d)", i, PingDbAttempts), zap.Error(err))
		time.Sleep(2 * time.Second)
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err = db.PingContext(ctx); err != nil {
			logger.Fatal("Failed to ping database, exiting", zap.Error(err))
		}
	}

	authClient := clients.NewAuthClient("http://auth:8080/auth", logger)
	server := api.NewServer(":8080", logger, authClient, db)
	server.Run()

}
func connectToPostgres() (*sql.DB, error) {
	//env := os.Getenv("ENV")
	//sslMode := "require"
	//if env == "local" {
	//	sslMode = "disable"
	//}
	sslMode := "disable"
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbName, sslMode)
	return sql.Open("postgres", connString)
}
