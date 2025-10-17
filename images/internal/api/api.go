package api

import (
	"context"
	"database/sql"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"images/internal/clients"
	"images/internal/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ApiServer struct {
	addr       string
	logger     *zap.Logger
	authClient *clients.AuthClient
	db         *sql.DB
}

func NewApiServer(addr string, logger *zap.Logger, authClient *clients.AuthClient, db *sql.DB) *ApiServer {
	return &ApiServer{
		addr:       addr,
		logger:     logger,
		authClient: authClient,
		db:         db,
	}
}
func (a *ApiServer) Run() {
	mux := http.NewServeMux()
	a.registerRoutes(mux)
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "https://predigrowee.agh.edu.pl",
			"https://www.predigrowee.agh.edu.pl"}, // Allow requests from this origin
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},  // Add the methods you need
		AllowedHeaders:   []string{"Authorization", "Content-Type"}, // Add the headers you need
	})
	srv := &http.Server{
		Addr:         a.addr,
		Handler:      corsMiddleware.Handler(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	// Start server
	go func() {
		a.logger.Info("Starting server on " + a.addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	a.logger.Info("Server exiting")

}
func (a *ApiServer) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /images/questions/{questionId}/image/{id}", middleware.VerifyToken(NewQuestionImagesHandler(a.logger, a.db).Handle, a.authClient))

	paramImagesHandler := NewParamImagesHandler(a.logger, a.db)
	mux.HandleFunc("GET /images/params/{id}", paramImagesHandler.GetImage)
	mux.HandleFunc("POST /images/params/{id}", middleware.VerifyToken(paramImagesHandler.PostImage, a.authClient))
}
