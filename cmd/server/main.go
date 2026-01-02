package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"real-time-leaderboard/internal/config"
	authHTTP "real-time-leaderboard/internal/module/auth/adapters/http"
	authApp "real-time-leaderboard/internal/module/auth/application"
	authJWT "real-time-leaderboard/internal/module/auth/infrastructure/jwt"
	authInfra "real-time-leaderboard/internal/module/auth/infrastructure/repository"
	leaderboardHTTP "real-time-leaderboard/internal/module/leaderboard/adapters/http"
	"real-time-leaderboard/internal/module/leaderboard/adapters/websocket"
	leaderboardApp "real-time-leaderboard/internal/module/leaderboard/application"
	leaderboardInfra "real-time-leaderboard/internal/module/leaderboard/infrastructure/repository"
	reportHTTP "real-time-leaderboard/internal/module/report/adapters/http"
	reportApp "real-time-leaderboard/internal/module/report/application"
	reportInfra "real-time-leaderboard/internal/module/report/infrastructure/repository"
	scoreHTTP "real-time-leaderboard/internal/module/score/adapters/http"
	scoreApp "real-time-leaderboard/internal/module/score/application"
	scoreInfra "real-time-leaderboard/internal/module/score/infrastructure/repository"
	"real-time-leaderboard/internal/shared/database"
	"real-time-leaderboard/internal/shared/errors"
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
		l.Errorf("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := redisInfra.NewClient(cfg.Redis, l)
	if err != nil {
		l.Errorf("Failed to connect to Redis: %v", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// Initialize repositories
	userRepo := authInfra.NewPostgresUserRepository(db.Pool)
	jwtMgr := authJWT.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)

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
	authHandler := authHTTP.NewHandler(authUseCase)
	scoreHandler := scoreHTTP.NewHandler(scoreUseCase)
	leaderboardHandler := leaderboardHTTP.NewHandler(leaderboardUseCase)
	reportHandler := reportHTTP.NewHandler(reportUseCase)

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
		l.Infof("Server starting on %s", cfg.Server.GetAddr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Errorf("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		l.Errorf("Server forced to shutdown: %v", err)
	}

	l.Info("Server exited")
}

func setupRouter(
	cfg *config.Config,
	l *logger.Logger,
	authUseCase *authApp.AuthUseCase,
	authHandler *authHTTP.Handler,
	scoreHandler *scoreHTTP.Handler,
	leaderboardHandler *leaderboardHTTP.Handler,
	reportHandler *reportHTTP.Handler,
	leaderboardHub *websocket.Hub,
) *gin.Engine {
	if cfg.Logger.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware (order matters!)
	// 1. Recovery - First to catch panics from all other middleware
	// 2. RequestID - Early to generate ID for all subsequent middleware and logs
	// 3. CORS - After RequestID so responses include request ID, but early for OPTIONS handling
	// 4. RequestLogger - Last to log after request processing completes
	router.Use(middleware.Recovery(l))
	router.Use(middleware.RequestID())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestLogger(l))

	// 404 Not Found handler
	router.NoRoute(func(c *gin.Context) {
		response.ErrorWithStatus(c, http.StatusNotFound, errors.CodeNotFound, "Route not found")
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes (no auth required)
		authHandler.RegisterRoutes(api)

		// Protected routes
		authMiddleware := middleware.NewAuthMiddleware(authUseCase.ValidateToken)
		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			scoreHandler.RegisterRoutes(protected)
		}

		// Public routes
		leaderboardHandler.RegisterRoutes(api)
		reportHandler.RegisterRoutes(api)
	}

	// WebSocket route
	router.GET("/ws/leaderboard", websocket.HandleWebSocket(leaderboardHub))

	return router
}
