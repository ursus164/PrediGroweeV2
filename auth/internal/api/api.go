package api

import (
	"auth/internal/handlers"
	"auth/internal/middleware"
	"auth/internal/storage"
	"context"
	"encoding/json"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ApiServer struct {
	addr    string
	storage storage.Store
	logger  *zap.Logger
}

func NewApiServer(addr string, store storage.Store, logger *zap.Logger) *ApiServer {
	return &ApiServer{
		addr:    addr,
		storage: store,
		logger:  logger,
	}
}

func (a *ApiServer) Run() {
	mux := http.NewServeMux()
	a.registerRoutes(mux)
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:80", "https://predigrowee.agh.edu.pl",
			"https://www.predigrowee.agh.edu.pl"}, // Allow requests from this origin
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},  // Add the methods you need
		AllowedHeaders:   []string{"Authorization", "Content-Type"}, // Add the headers you need
	})
	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      corsMiddleware.Handler(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	// Start server
	go func() {
		a.logger.Info("Starting server on :8080")
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

func (a *ApiServer) registerRoutes(router *http.ServeMux) {
	// external
	router.HandleFunc("GET /auth/health", a.HealthCheckHandler)
	router.HandleFunc("POST /auth/register", handlers.NewRegisterHandler(a.storage, a.logger).Register)
	router.HandleFunc("POST /auth/login", handlers.NewLoginHandler(a.storage, a.logger).Handle)
	router.HandleFunc("POST /auth/login/google", handlers.NewOauthLoginHandler(a.storage, a.logger).HandleGoogle)
	router.HandleFunc("GET /auth/user", middleware.ValidateAccessToken(handlers.NewGetUserHandler(a.storage, a.logger).Handle, a.storage))
	router.HandleFunc("PUT /auth/users/{id}", middleware.ValidateAccessToken(handlers.NewUpdateUserHandler(a.storage, a.logger).Handle, a.storage))
	router.HandleFunc("POST /auth/verify", middleware.ValidateAccessToken(handlers.NewVerifyTokenHandler().Handle, a.storage))
	router.HandleFunc("GET /auth/verifySession", middleware.ValidateSession(handlers.NewVerifySessionHandler(a.logger).Handle, a.storage))
	router.HandleFunc("POST /auth/refresh", middleware.ValidateSession(handlers.NewRefreshTokenHandler(a.storage, a.logger).Handle, a.storage))
	router.HandleFunc("POST /auth/logout", middleware.ValidateSession(handlers.NewLogOutHandler(a.storage, a.logger).Handle, a.storage))
	resetHandler := handlers.NewResetPasswordHandler(a.storage, a.logger)
	router.HandleFunc("POST /auth/reset-password", resetHandler.RequestReset)
	router.HandleFunc("POST /auth/reset-password/confirm", resetHandler.Reset)
	router.HandleFunc("GET /auth/reset-password/verify", resetHandler.VerifyToken)

	router.HandleFunc("GET /auth/verify-email", handlers.NewRegisterHandler(a.storage, a.logger).Verify)

	// internal
	internalApiKey := os.Getenv("INTERNAL_API_KEY")
	router.HandleFunc("GET /auth/users", middleware.InternalAuth(handlers.NewGetAllUsersHandler(a.storage, a.logger).Handle, a.logger, internalApiKey))
	router.HandleFunc("GET /auth/users/{id}", middleware.InternalAuth(handlers.NewAdminGetUserHandler(a.storage, a.logger).Handle, a.logger, internalApiKey))
	router.HandleFunc("PATCH /auth/users/{id}", middleware.InternalAuth(handlers.NewAdminUpdateUserHandler(a.storage, a.logger).Handle, a.logger, internalApiKey))
	router.HandleFunc("DELETE /auth/users/{id}", middleware.InternalAuth(handlers.NewDeleteUserHandler(a.storage, a.logger).Handle, a.logger, internalApiKey))
	router.HandleFunc("GET /auth/roles", middleware.InternalAuth(handlers.NewGetAllRolesHandler(a.storage, a.logger).Handle, a.logger, internalApiKey))
	router.HandleFunc("POST /auth/roles", middleware.InternalAuth(handlers.NewCreateRoleHandler(a.storage, a.logger).Handle, a.logger, internalApiKey))
	router.HandleFunc("PUT /auth/roles/{id}", middleware.InternalAuth(handlers.NewUpdateRoleHandler(a.storage, a.logger).Handle, a.logger, internalApiKey))
	router.HandleFunc("DELETE /auth/roles/{id}", middleware.InternalAuth(handlers.NewDeleteRoleHandler(a.storage, a.logger).Handle, a.logger, internalApiKey))

	router.HandleFunc("GET /auth/summary", middleware.InternalAuth(handlers.NewSummaryHandler(a.storage, a.logger).Handle, a.logger, internalApiKey))

	router.HandleFunc("POST /auth/notify-approved",
	middleware.InternalAuth(handlers.NewNotifyApprovedHandler(a.storage, a.logger).Handle, a.logger, internalApiKey))
}

func (a *ApiServer) HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	if err := a.storage.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy"})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
