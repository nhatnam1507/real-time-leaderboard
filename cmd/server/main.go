// Package main provides the entry point for the real-time leaderboard server.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"real-time-leaderboard/api"
	"real-time-leaderboard/internal/config"
	v1Auth "real-time-leaderboard/internal/module/auth/adapters/rest/v1"
	authApp "real-time-leaderboard/internal/module/auth/application"
	authJWT "real-time-leaderboard/internal/module/auth/infrastructure/jwt"
	authInfra "real-time-leaderboard/internal/module/auth/infrastructure/repository"
	v1Leaderboard "real-time-leaderboard/internal/module/leaderboard/adapters/rest/v1"
	"real-time-leaderboard/internal/module/leaderboard/adapters/websocket"
	leaderboardApp "real-time-leaderboard/internal/module/leaderboard/application"
	leaderboardInfra "real-time-leaderboard/internal/module/leaderboard/infrastructure/repository"
	v1Report "real-time-leaderboard/internal/module/report/adapters/rest/v1"
	reportApp "real-time-leaderboard/internal/module/report/application"
	reportInfra "real-time-leaderboard/internal/module/report/infrastructure/repository"
	v1Score "real-time-leaderboard/internal/module/score/adapters/rest/v1"
	scoreApp "real-time-leaderboard/internal/module/score/application"
	scoreInfra "real-time-leaderboard/internal/module/score/infrastructure/repository"
	"real-time-leaderboard/internal/shared/database"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/middleware"
	redisInfra "real-time-leaderboard/internal/shared/redis"
	"real-time-leaderboard/internal/shared/response"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	l := logger.New(cfg.Logger.Level, cfg.Logger.Pretty)

	// Initialize database
	db, err := database.NewPostgres(cfg.Database, l)
	if err != nil {
		l.Errorf(context.TODO(), "Failed to connect to database: %v", err)
		return
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := redisInfra.NewClient(cfg.Redis, l)
	if err != nil {
		l.Errorf(context.TODO(), "Failed to connect to Redis: %v", err)
		return
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			l.Errorf(context.TODO(), "Failed to close Redis connection: %v", err)
		}
	}()

	// Initialize repositories
	userRepo := authInfra.NewPostgresUserRepository(db.Pool)
	jwtMgr := authJWT.NewManager(cfg.JWT.SecretKey, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)

	scoreRepo := scoreInfra.NewPostgresScoreRepository(db.Pool)
	scoreLeaderboardRepo := scoreInfra.NewRedisLeaderboardRepository(redisClient.GetClient())

	leaderboardRepo := leaderboardInfra.NewRedisLeaderboardRepository(redisClient.GetClient())

	reportRedisRepo := reportInfra.NewRedisReportRepository(redisClient.GetClient())
	reportPostgresRepo := reportInfra.NewPostgresReportRepository(db.Pool)
	reportRepo := reportInfra.NewCompositeReportRepository(reportRedisRepo, reportPostgresRepo)

	// Initialize use cases
	authUseCase := authApp.NewAuthUseCase(userRepo, jwtMgr, l)
	scoreUseCase := scoreApp.NewScoreUseCase(scoreRepo, scoreLeaderboardRepo, l)
	leaderboardUseCase := leaderboardApp.NewLeaderboardUseCase(leaderboardRepo, l)
	reportUseCase := reportApp.NewReportUseCase(reportRepo, l)

	// Initialize handlers
	authHandler := v1Auth.NewHandler(authUseCase)
	scoreHandler := v1Score.NewHandler(scoreUseCase)
	leaderboardHandler := v1Leaderboard.NewHandler(leaderboardUseCase)
	reportHandler := v1Report.NewHandler(reportUseCase)

	// Initialize WebSocket hub
	leaderboardHub := websocket.NewHub(leaderboardUseCase)
	go leaderboardHub.Run()

	// Setup router
	router := setupRouter(cfg, l, authUseCase, authHandler, scoreHandler, leaderboardHandler, reportHandler, leaderboardHub)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.GetAddr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		l.Infof(context.TODO(), "Server starting on %s", cfg.Server.GetAddr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Errorf(context.TODO(), "Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info(context.TODO(), "Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		l.Errorf(context.TODO(), "Server forced to shutdown: %v", err)
	}

	l.Info(context.TODO(), "Server exited")
}

func setupRouter(
	cfg *config.Config,
	l *logger.Logger,
	authUseCase *authApp.AuthUseCase,
	authHandler *v1Auth.Handler,
	scoreHandler *v1Score.Handler,
	leaderboardHandler *v1Leaderboard.Handler,
	reportHandler *v1Report.Handler,
	leaderboardHub *websocket.Hub,
) *gin.Engine {
	// Set gin mode based on config
	if cfg.Logger.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router with correct settings
	router := gin.New()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// WebSocket route
	router.GET("/ws/leaderboard", websocket.HandleWebSocket(leaderboardHub))

	// Setup API router (with middleware, grouped by /api)
	setupAPIRouter(router, l, authUseCase, authHandler, scoreHandler, leaderboardHandler, reportHandler)

	// Setup docs router (without middleware, prefixed by /docs)
	setupDocsRouter(router)

	// 404 Not Found handler for routes outside /api and /docs
	router.NoRoute(func(c *gin.Context) {
		response.ErrorWithStatus(c, http.StatusNotFound, response.CodeNotFound, "Route not found")
	})

	return router
}

func setupAPIRouter(
	router *gin.Engine,
	l *logger.Logger,
	authUseCase *authApp.AuthUseCase,
	authHandler *v1Auth.Handler,
	scoreHandler *v1Score.Handler,
	leaderboardHandler *v1Leaderboard.Handler,
	reportHandler *v1Report.Handler,
) {
	// Group API routes by /api prefix
	apiGroup := router.Group("/api")

	// Middleware (order matters!)
	// 1. Recovery - First to catch panics from all other middleware
	// 2. RequestID - Early to generate ID for all subsequent middleware and logs
	// 3. CORS - After RequestID so responses include request ID, but early for OPTIONS handling
	// 4. RequestLogger - Last to log after request processing completes
	apiGroup.Use(middleware.Recovery(l))
	apiGroup.Use(middleware.RequestID())
	apiGroup.Use(middleware.CORS())
	apiGroup.Use(middleware.RequestLogger(l))

	// API v1 routes
	v1Group := apiGroup.Group("/v1")

	// Public routes group (no auth required)
	v1PublicGroup := v1Group.Group("")
	{
		// OpenAPI 3.0 spec endpoints (versioned) - using embedded files
		v1PublicGroup.GET("/openapi.yaml", func(c *gin.Context) {
			c.Data(http.StatusOK, "application/x-yaml", api.OpenAPIV1YAML)
		})

		v1PublicGroup.GET("/openapi.json", func(c *gin.Context) {
			c.Data(http.StatusOK, "application/json", api.OpenAPIV1JSON)
		})

		// Auth routes (no auth required)
		authHandler.RegisterRoutes(v1PublicGroup)

		// Public routes
		leaderboardHandler.RegisterRoutes(v1PublicGroup)
		reportHandler.RegisterRoutes(v1PublicGroup)
	}

	// Protected routes group (auth required)
	authMiddleware := middleware.NewAuthMiddleware(authUseCase.ValidateToken)
	v1ProtectedGroup := v1Group.Group("")
	v1ProtectedGroup.Use(authMiddleware.RequireAuth())
	{
		scoreHandler.RegisterRoutes(v1ProtectedGroup)
	}
}

func setupDocsRouter(router *gin.Engine) {
	// Swagger UI for OpenAPI 3.0 (with version selection) - using embedded file
	// Prefixed by /docs, no middleware applied
	docsGroup := router.Group("/docs")
	{
		docsGroup.GET("", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
		})

		docsGroup.GET("/index.html", func(c *gin.Context) {
			c.Data(http.StatusOK, "text/html", api.SwaggerUIHTML)
		})
	}
}
