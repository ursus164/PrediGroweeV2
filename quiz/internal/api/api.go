package api

import (
	"context"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"quiz/internal/clients"
	"quiz/internal/handlers"
	"quiz/internal/middleware"
	"quiz/internal/storage"
	"syscall"
	"time"
)

type ApiServer struct {
	addr        string
	storage     storage.Store
	logger      *zap.Logger
	authClient  *clients.AuthClient
	statsClient *clients.StatsClient
}

func NewApiServer(addr string, store storage.Store, logger *zap.Logger, authClient *clients.AuthClient, statsClient *clients.StatsClient) *ApiServer {
	return &ApiServer{
		addr:        addr,
		storage:     store,
		logger:      logger,
		authClient:  authClient,
		statsClient: statsClient,
	}
}

func (a *ApiServer) Run() {
	a.logger.Info("run server")
	mux := http.NewServeMux()
	a.registerRoutes(mux)
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "https://predigrowee.agh.edu.pl",
			"https://www.predigrowee.agh.edu.pl"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposedHeaders:   []string{"X-Quiz-Is-Last"},
	})
	srv := &http.Server{
		Addr:         a.addr,
		Handler:      corsMiddleware.Handler(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	a.logger.Info("about to start the server")
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
	a.logger.Info("registering routes")

	// user actions, external api
	mux.HandleFunc("GET /quiz/sessions", middleware.VerifyToken(handlers.NewGetUserActiveSessionsHandler(a.storage, a.logger).Handle, a.authClient))
	mux.HandleFunc("POST /quiz/sessions/new", middleware.VerifyToken(handlers.NewStartQuizHandler(a.storage, a.logger, a.statsClient).Handle, a.authClient))
	mux.HandleFunc("GET /quiz/sessions/{quizSessionId}/nextQuestion", middleware.VerifyToken(handlers.NewGetNextQuestionHandler(a.storage, a.logger).Handle, a.authClient))
	mux.HandleFunc("POST /quiz/sessions/{quizSessionId}/answer", middleware.VerifyToken(handlers.NewSubmitAnswerHandler(a.storage, a.logger, a.statsClient).Handle, a.authClient))
	mux.HandleFunc("POST /quiz/sessions/{quizSessionId}/finish", middleware.VerifyToken(handlers.NewFinishQuizHandler(a.storage, a.logger, a.statsClient).Handle, a.authClient))

	// internal api
	apiKey := os.Getenv("INTERNAL_API_KEY")
	mux.HandleFunc("GET /quiz/summary", middleware.InternalAuth(handlers.NewSummaryHandler(a.storage, a.logger).Handle, a.logger, apiKey))
	mux.HandleFunc("POST /quiz/approve", middleware.InternalAuth(handlers.NewApprovalHandler(a.storage, a.logger).Approve, a.logger, apiKey))
	mux.HandleFunc("GET /quiz/approved", middleware.InternalAuth(handlers.NewApprovalHandler(a.storage, a.logger).GetApproved, a.logger, apiKey))
	mux.HandleFunc("POST /quiz/unapprove", middleware.InternalAuth(handlers.NewApprovalHandler(a.storage, a.logger).Unapprove, a.logger, apiKey))

	// questions
	questionHandler := handlers.NewQuestionHandler(a.storage, a.logger)
	mux.HandleFunc("GET /quiz/q/{id}", middleware.VerifyToken(questionHandler.GetQuestion, a.authClient))
	mux.HandleFunc("GET /quiz/questions/{id}", middleware.InternalAuth(questionHandler.GetQuestion, a.logger, apiKey))
	mux.HandleFunc("POST /quiz/questions", middleware.InternalAuth(questionHandler.CreateQuestion, a.logger, apiKey))
	mux.HandleFunc("PATCH /quiz/questions/{id}", middleware.InternalAuth(questionHandler.UpdateQuestion, a.logger, apiKey))
	mux.HandleFunc("DELETE /quiz/questions/{id}", middleware.InternalAuth(questionHandler.DeleteQuestion, a.logger, apiKey))
	mux.HandleFunc("GET /quiz/questions", middleware.InternalAuth(questionHandler.GetAllQuestions, a.logger, apiKey))

	// options
	optionsHandler := handlers.NewOptionsHandler(a.storage, a.logger)
	mux.HandleFunc("GET /quiz/options", middleware.InternalAuth(optionsHandler.GetAllOptions, a.logger, apiKey))
	mux.HandleFunc("POST /quiz/options", middleware.InternalAuth(optionsHandler.CreateOption, a.logger, apiKey))
	mux.HandleFunc("PATCH /quiz/options/{id}", middleware.InternalAuth(optionsHandler.UpdateOption, a.logger, apiKey))
	mux.HandleFunc("DELETE /quiz/options/{id}", middleware.InternalAuth(optionsHandler.DeleteOption, a.logger, apiKey))

	// groups
	groupHandler := handlers.NewGroupHandler(a.storage, a.logger)
	mux.HandleFunc("POST /quiz/groups", middleware.InternalAuth(groupHandler.CreateGroup, a.logger, apiKey))
	mux.HandleFunc("PUT /quiz/groups/{id}", middleware.InternalAuth(groupHandler.UpdateGroup, a.logger, apiKey))
	mux.HandleFunc("DELETE /quiz/groups/{id}", middleware.InternalAuth(groupHandler.DeleteGroup, a.logger, apiKey))
	mux.HandleFunc("GET /quiz/groups", middleware.InternalAuth(groupHandler.GetAllGroups, a.logger, apiKey))

	// parameters & settings
	parameterHandler := handlers.NewParameterHandler(a.storage, a.logger)
	mux.HandleFunc("GET /quiz/parameters", parameterHandler.GetAllParameters)
	mux.HandleFunc("POST /quiz/parameters", middleware.InternalAuth(parameterHandler.CreateParameter, a.logger, apiKey))
	mux.HandleFunc("PATCH /quiz/parameters/{id}", middleware.InternalAuth(parameterHandler.UpdateParameter, a.logger, apiKey))
	mux.HandleFunc("DELETE /quiz/parameters/{id}", middleware.InternalAuth(parameterHandler.DeleteParameter, a.logger, apiKey))
	mux.HandleFunc("PUT /quiz/parameters/order", middleware.InternalAuth(parameterHandler.UpdateOrder, a.logger, apiKey))
	mux.HandleFunc("POST /quiz/settings", middleware.InternalAuth(handlers.NewSettingsHandler(a.storage, a.logger).UpdateSettings, a.logger, apiKey))
	mux.HandleFunc("GET /quiz/settings", middleware.InternalAuth(handlers.NewSettingsHandler(a.storage, a.logger).GetSettings, a.logger, apiKey))

	// BUG REPORTS
	reportHandler := handlers.NewCaseReportHandler(a.storage, a.logger, a.authClient)
	mux.HandleFunc("POST /quiz/cases/{caseId}/report", middleware.VerifyToken(reportHandler.Report, a.authClient))          // user
	mux.HandleFunc("GET /quiz/reports", middleware.InternalAuth(reportHandler.List, a.logger, apiKey))                     // admin
	mux.HandleFunc("DELETE /quiz/reports/{id}", middleware.InternalAuth(reportHandler.Delete, a.logger, apiKey))           // admin
	mux.HandleFunc("PUT /quiz/reports/{id}/note",
    middleware.InternalAuth(reportHandler.SetNote, a.logger, apiKey))
    mux.HandleFunc("GET /quiz/reports/pendingCount",
    middleware.InternalAuth(reportHandler.PendingCount, a.logger, apiKey))
	caseHandler := handlers.NewCaseHandler(a.storage, a.logger)
	mux.HandleFunc("GET /quiz/cases/{id}/parameters/v3", middleware.VerifyToken(caseHandler.GetCaseParametersV3, a.authClient))

	// teacher endpoints (tests)
	testsHandler := handlers.NewTestsHandler(a.storage, a.logger)
	mux.HandleFunc("POST /quiz/tests", middleware.VerifyToken(testsHandler.Create, a.authClient))
	mux.HandleFunc("GET /quiz/tests", middleware.VerifyToken(testsHandler.ListMine, a.authClient))
	mux.HandleFunc("GET /quiz/tests/{code}/progress", middleware.VerifyToken(testsHandler.ProgressByCode, a.authClient))
	mux.HandleFunc("GET /quiz/tests/{id}", middleware.VerifyToken(
	handlers.NewTeacherGetTestHandler(a.storage, a.logger).Handle, a.authClient))
	mux.HandleFunc("DELETE /quiz/tests/{id}", middleware.VerifyToken(
	handlers.NewTeacherDeleteTestHandler(a.storage, a.logger).Handle, a.authClient))

	// difficulty voting
    diffMark := handlers.NewMarkDifficultyHandler(a.storage, a.logger)
    mux.HandleFunc("POST /quiz/questions/{id}/difficulty",
    middleware.VerifyToken(diffMark.Handle, a.authClient))

    diffMe := handlers.NewGetMyDifficultyHandler(a.storage, a.logger)
    mux.HandleFunc("GET /quiz/questions/{id}/difficulty/me",
    middleware.VerifyToken(diffMe.Handle, a.authClient))

    diffSum := handlers.NewGetDifficultySummaryHandler(a.storage, a.logger)
    mux.HandleFunc("GET /quiz/questions/{id}/difficulty/summary",
    middleware.VerifyToken(diffSum.Handle, a.authClient))
    batch := handlers.NewGetDifficultySummaryBatchHandler(a.storage, a.logger)
    mux.HandleFunc("GET /quiz/questions/difficulty/summary",
    middleware.VerifyToken(batch.Handle, a.authClient))

    las := handlers.NewListActiveSessionsHandler(a.storage, a.logger)
    mux.HandleFunc("GET /quiz/sessions/active", middleware.InternalAuth(las.Handle, a.logger, apiKey))

	mux.HandleFunc("GET /quiz/tests/{code}/sessions",
	middleware.InternalAuth(handlers.NewTestSessionsHandler(a.storage, a.logger).ListByCode, a.logger, apiKey))

	// favorites
	favHandler := handlers.NewFavoriteHandler(a.storage, a.logger, a.authClient)
	mux.HandleFunc("POST /quiz/cases/{caseId}/favorite", middleware.VerifyToken(favHandler.Add, a.authClient))
	mux.HandleFunc("DELETE /quiz/cases/{caseId}/favorite", middleware.VerifyToken(favHandler.Remove, a.authClient))
	mux.HandleFunc("GET /quiz/favorites", middleware.VerifyToken(favHandler.List, a.authClient))
	mux.HandleFunc("PUT /quiz/cases/{caseId}/favorite/note", middleware.VerifyToken(favHandler.SetNote, a.authClient))
}
