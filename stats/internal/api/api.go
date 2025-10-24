// Package api provides the HTTP server implementation for the stats service.
package api

import (
	"context"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"stats/internal/clients"
	"stats/internal/handlers"
	"stats/internal/middleware"
	"stats/internal/storage"
	"syscall"
	"time"
)

// Server is the HTTP server for the stats service.
type Server struct {
	addr       string
	authClient *clients.AuthClient
	storage    storage.Storage
	logger     *zap.Logger
}

// NewServer creates a new Server instance.
func NewServer(addr string, storage storage.Storage, logger *zap.Logger, authClient *clients.AuthClient) *Server {
	return &Server{
		addr:       addr,
		authClient: authClient,
		storage:    storage,
		logger:     logger,
	}
}

// Run starts the HTTP server and blocks until it's shut down.
func (a *Server) Run() {
	mux := http.NewServeMux()
	a.registerRoutes(mux)
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:3001", "https://predigrowee.agh.edu.pl",
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

func (a *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})
	internalAPIKey := os.Getenv("INTERNAL_API_KEY")
	//internal
	mux.HandleFunc("POST /stats/sessions/save", middleware.InternalAuth(handlers.NewQuizStatsHandler(a.storage, a.logger).SaveSession, a.logger, internalAPIKey))
	mux.HandleFunc("POST /stats/sessions/{quizSessionId}/respond", middleware.InternalAuth(handlers.NewQuizStatsHandler(a.storage, a.logger).SaveResponse, a.logger, internalAPIKey))
	mux.HandleFunc("POST /stats/sessions/{quizSessionId}/finish", middleware.InternalAuth(handlers.NewQuizStatsHandler(a.storage, a.logger).FinishSession, a.logger, internalAPIKey))
	// admin
	allStatsHandler := handlers.NewGetAllStatsHandler(a.storage, a.logger)
	userStatsHandler := handlers.NewUserStatsHandler(a.storage, a.logger)
	mux.HandleFunc("GET /stats/users/{id}", middleware.InternalAuth(userStatsHandler.Handle, a.logger, internalAPIKey))
	mux.HandleFunc("GET /stats/responses", middleware.InternalAuth(allStatsHandler.GetResponses, a.logger, internalAPIKey))
	mux.HandleFunc("GET /stats/questions/{id}/stats", middleware.InternalAuth(allStatsHandler.GetStatsForQuestion, a.logger, internalAPIKey))
	mux.HandleFunc("GET /stats/activity", middleware.InternalAuth(allStatsHandler.GetActivity, a.logger, internalAPIKey))
	mux.HandleFunc("GET /stats/summary", middleware.InternalAuth(allStatsHandler.GetSummary, a.logger, internalAPIKey))
	mux.HandleFunc("GET /stats/surveys/users/{id}", middleware.InternalAuth(handlers.NewSurveysHandler(a.storage, a.logger).GetSurvey, a.logger, internalAPIKey))
	mux.HandleFunc("GET /stats/grouped", middleware.InternalAuth(allStatsHandler.GetStatsGroupedBySurvey, a.logger, internalAPIKey))
		mux.HandleFunc("DELETE /stats/internal/users/{id}/responses", middleware.InternalAuth(userStatsHandler.DeleteUserResponses, a.logger, internalAPIKey))
	mux.HandleFunc("DELETE /stats/responses/{id}", middleware.InternalAuth(allStatsHandler.DeleteResponse, a.logger, internalAPIKey))
	mux.HandleFunc("POST /stats/internal/users/stats", middleware.InternalAuth(userStatsHandler.GetAllUsersStats, a.logger, internalAPIKey))

	//external
	mux.HandleFunc("GET /stats/userStats", middleware.VerifyToken(handlers.NewUserStatsHandler(a.storage, a.logger).Handle, a.authClient))
	mux.HandleFunc("GET /stats/quiz/{quizSessionId}", middleware.VerifyToken(handlers.NewQuizStatsHandler(a.storage, a.logger).GetStats, a.authClient))
	mux.HandleFunc("GET /stats/sessions", middleware.VerifyToken(handlers.NewUserStatsHandler(a.storage, a.logger).GetUserSessions, a.authClient))
	mux.HandleFunc("POST /stats/survey", middleware.VerifyToken(handlers.NewSurveysHandler(a.storage, a.logger).Save, a.authClient))
	mux.HandleFunc("GET /stats/survey", middleware.VerifyToken(handlers.NewSurveysHandler(a.storage, a.logger).GetSurvey, a.authClient))

	leaderboardHandler := handlers.NewLeaderboardHandler(a.storage, a.logger)
    mux.HandleFunc("GET /stats/leaderboard", leaderboardHandler.Get)

	mux.HandleFunc("GET /stats/sessions/accuracy", middleware.InternalAuth(handlers.NewSessionsAccuracyHandler(a.storage, a.logger).Handle, a.logger, internalAPIKey))
}
