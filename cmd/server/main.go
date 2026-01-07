// Package main provides the entry point for the real-time leaderboard server.
//
//	@title			Real-Time Leaderboard API
//	@version		1.0
//	@description	Real-time leaderboard system with WebSocket support for live updates
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//	@contact.url	http://www.example.com/support
//	@contact.email	support@example.com
//
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@securityDefinitions.bearer	Bearer
//	@securityDefinitions.bearer	type	apiKey
//	@securityDefinitions.bearer	in	header
//	@securityDefinitions.bearer	name	Authorization
//	@securityDefinitions.bearer	bearerFormat	JWT
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
	authREST "real-time-leaderboard/internal/module/auth/adapters/rest"
	authApp "real-time-leaderboard/internal/module/auth/application"
	authJWT "real-time-leaderboard/internal/module/auth/infrastructure/jwt"
	authInfra "real-time-leaderboard/internal/module/auth/infrastructure/repository"
	leaderboardREST "real-time-leaderboard/internal/module/leaderboard/adapters/rest"
	"real-time-leaderboard/internal/module/leaderboard/adapters/websocket"
	leaderboardApp "real-time-leaderboard/internal/module/leaderboard/application"
	leaderboardInfra "real-time-leaderboard/internal/module/leaderboard/infrastructure/repository"
	reportREST "real-time-leaderboard/internal/module/report/adapters/rest"
	reportApp "real-time-leaderboard/internal/module/report/application"
	reportInfra "real-time-leaderboard/internal/module/report/infrastructure/repository"
	scoreREST "real-time-leaderboard/internal/module/score/adapters/rest"
	scoreApp "real-time-leaderboard/internal/module/score/application"
	scoreInfra "real-time-leaderboard/internal/module/score/infrastructure/repository"
	"real-time-leaderboard/internal/shared/database"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/middleware"
	redisInfra "real-time-leaderboard/internal/shared/redis"
	"real-time-leaderboard/internal/shared/response"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "real-time-leaderboard/docs" // Swagger docs
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
	authHandler := authREST.NewHandler(authUseCase)
	scoreHandler := scoreREST.NewHandler(scoreUseCase)
	leaderboardHandler := leaderboardREST.NewHandler(leaderboardUseCase)
	reportHandler := reportREST.NewHandler(reportUseCase)

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
	authHandler *authREST.Handler,
	scoreHandler *scoreREST.Handler,
	leaderboardHandler *leaderboardREST.Handler,
	reportHandler *reportREST.Handler,
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
		response.ErrorWithStatus(c, http.StatusNotFound, response.CodeNotFound, "Route not found")
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
