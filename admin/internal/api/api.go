package api

import (
	"admin/clients"
	"admin/internal/handlers"
	"admin/internal/middleware"
	"context"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ApiServer struct {
	addr        string
	logger      *zap.Logger
	authClient  clients.AuthClient
	statsClient clients.StatsClient
	quizClient  clients.QuizClient
}

func NewApiServer(addr string, logger *zap.Logger, authClient clients.AuthClient, statsClient clients.StatsClient, quizClient clients.QuizClient) *ApiServer {
	return &ApiServer{
		addr:        addr,
		logger:      logger,
		authClient:  authClient,
		statsClient: statsClient,
		quizClient:  quizClient,
	}
}

func (a *ApiServer) Run() {
	mux := http.NewServeMux()
	a.registerRoutes(mux)

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://predigrowee.agh.edu.pl"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
	})
	srv := &http.Server{
		Addr:         a.addr,
		Handler:      corsMiddleware.Handler(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		a.logger.Info("Starting server on " + a.addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

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

	// users
	usersHandler := handlers.NewUsersHandler(a.logger, a.authClient, a.statsClient)
	mux.HandleFunc("GET /admin/users", middleware.VerifyAdmin(usersHandler.GetUsers, a.authClient))
	mux.HandleFunc("GET /admin/users/{id}", middleware.VerifyAdmin(usersHandler.GetUserDetails, a.authClient))
	mux.HandleFunc("PATCH /admin/users/{id}", middleware.VerifyAdmin(usersHandler.UpdateUser, a.authClient))
	mux.HandleFunc("DELETE /admin/users/{id}", middleware.VerifyAdmin(usersHandler.DeleteUser, a.authClient))
	mux.HandleFunc("GET /admin/users/-/surveys", middleware.VerifyAdmin(usersHandler.GetAllUsersSurveys, a.authClient))

	// quiz
	quizHandler := handlers.NewQuizHandler(a.logger, a.quizClient, a.statsClient)
	mux.HandleFunc("GET /admin/questions", middleware.VerifyAdmin(quizHandler.GetAllQuestions, a.authClient))
	mux.HandleFunc("GET /admin/questions/{id}", middleware.VerifyAdmin(quizHandler.GetQuestion, a.authClient))
	mux.HandleFunc("PATCH /admin/questions/{id}", middleware.VerifyAdmin(quizHandler.UpdateQuestion, a.authClient))

	// parameters
	mux.HandleFunc("GET /admin/parameters", middleware.VerifyAdmin(quizHandler.GetAllParameters, a.authClient))
	mux.HandleFunc("POST /admin/parameters", middleware.VerifyAdmin(quizHandler.CreateParameter, a.authClient))
	mux.HandleFunc("PATCH /admin/parameters/{id}", middleware.VerifyAdmin(quizHandler.UpdateParameter, a.authClient))
	mux.HandleFunc("DELETE /admin/parameters/{id}", middleware.VerifyAdmin(quizHandler.DeleteParameter, a.authClient))
	mux.HandleFunc("PUT /admin/parameters/order", middleware.VerifyAdmin(quizHandler.UpdateParametersOrder, a.authClient))

	// options
	mux.HandleFunc("GET /admin/options", middleware.VerifyAdmin(quizHandler.GetAllOptions, a.authClient))
	mux.HandleFunc("POST /admin/options", middleware.VerifyAdmin(quizHandler.CreateOption, a.authClient))
	mux.HandleFunc("PATCH /admin/options/{id}", middleware.VerifyAdmin(quizHandler.UpdateOption, a.authClient))
	mux.HandleFunc("DELETE /admin/options/{id}", middleware.VerifyAdmin(quizHandler.DeleteOption, a.authClient))

	// settings
	mux.HandleFunc("GET /admin/settings", middleware.VerifyAdmin(quizHandler.GetSettings, a.authClient))
	mux.HandleFunc("PATCH /admin/settings", middleware.VerifyAdmin(quizHandler.UpdateSettings, a.authClient))

	// stats
	statsHandler := handlers.NewAllStatsHandler(a.logger, a.statsClient)
	mux.HandleFunc("GET /admin/responses", middleware.VerifyAdmin(statsHandler.GetAllResponses, a.authClient))
	mux.HandleFunc("DELETE /admin/responses/{id}", middleware.VerifyAdmin(statsHandler.DeleteResponse, a.authClient))
	mux.HandleFunc("GET /admin/stats/questions/{questionId}", middleware.VerifyAdmin(statsHandler.GetStatsForQuestion, a.authClient))
	mux.HandleFunc("GET /admin/stats/questions", middleware.VerifyAdmin(statsHandler.GetStatsForAllQuestions, a.authClient))
	mux.HandleFunc("GET /admin/stats/activity", middleware.VerifyAdmin(statsHandler.GetActivityStats, a.authClient))
	mux.HandleFunc("GET /admin/stats/grouped", middleware.VerifyAdmin(statsHandler.GetStatsGroupedBySurvey, a.authClient))
	mux.HandleFunc("GET /admin/stats/users", middleware.VerifyAdmin(statsHandler.GetStatsForUsers, a.authClient))

	// dashboard
	mux.HandleFunc("GET /admin/dashboard", middleware.VerifyAdmin(
		handlers.NewSummaryHandler(a.logger, a.authClient, a.statsClient, a.quizClient).GetSummary, a.authClient))

	mux.HandleFunc("GET /admin/live/sessions/active", middleware.VerifyAdmin(quizHandler.ListActiveSessions, a.authClient))

	mux.HandleFunc("GET /admin/tests/{code}/progress", middleware.VerifyAdmin(handlers.NewTestsProgressHandler(a.quizClient, a.statsClient, a.logger).Get,a.authClient,))

}
