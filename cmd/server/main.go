package main

import (
	"ccany/internal/webfs"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ccany/ent/user"
	"ccany/internal/app"
	"ccany/internal/client"
	"ccany/internal/config"
	"ccany/internal/database"
	"ccany/internal/handlers"
	"ccany/internal/i18n"
	"ccany/internal/logging"
	"ccany/internal/middleware"
	"ccany/internal/monitoring"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Version information (set at build time)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Get master key
	masterKey := database.GetMasterKeyFromEnv()

	// Initialize database
	log.Println("Initializing database...")
	dbConfig := database.DefaultConfig()
	db, err := database.InitializeDatabase(ctx, dbConfig, masterKey)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Create configuration manager
	log.Println("Initializing configuration manager...")
	configManager := app.NewConfigManager(db, ctx)
	if err := configManager.Initialize(); err != nil {
		log.Fatalf("Failed to initialize config manager: %v", err)
	}

	// Get configuration
	cfg, err := configManager.GetConfig()
	if err != nil {
		log.Fatalf("Failed to get configuration: %v", err)
	}

	// Validate configuration
	if err := configManager.ValidateConfig(); err != nil {
		log.Printf("Configuration validation warning: %v", err)
		// Don't exit on validation failure, allow users to configure via web interface
	}

	// Setup logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Print startup info
	printStartupInfo(cfg, logger)

	// Create i18n service
	logger.Println("Initializing i18n service...")
	i18nService := i18n.NewI18nService(logger)

	// Create OpenAI client
	var openaiClient *client.OpenAIClient
	if cfg.OpenAIAPIKey != "" {
		openaiClient = client.NewOpenAIClient(
			cfg.OpenAIAPIKey,
			cfg.OpenAIBaseURL,
			cfg.RequestTimeout,
			logger,
		)
	} else {
		logger.Warn("OpenAI API key not configured")
	}

	// Setup Gin
	if level >= logrus.WarnLevel {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Create authentication middleware
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-this-in-production"
		logger.Warn("JWT_SECRET not set, using default secret (change this in production)")
	}
	authMiddleware := middleware.NewAuthMiddleware(db, jwtSecret, logger)

	// Create i18n middleware
	i18nMiddleware := middleware.NewI18nMiddleware(i18nService, logger)

	// Add i18n middleware to router
	router.Use(i18nMiddleware.Handler())

	// Create request logger
	requestLogger := logging.NewRequestLogger(db.Client, logger)

	// Create system monitor
	systemMonitor := monitoring.NewSystemMonitor(db.Client, requestLogger, logger, nil)
	if err := systemMonitor.Start(ctx); err != nil {
		log.Printf("Failed to start system monitor: %v", err)
	}

	// Create handlers
	messagesHandler := handlers.NewMessagesHandler(cfg, openaiClient, requestLogger, logger)
	healthHandler := handlers.NewHealthHandler(cfg, openaiClient, logger)
	configHandler := handlers.NewConfigHandler(configManager, logger)
	usersHandler := handlers.NewUsersHandler(db, authMiddleware, logger)
	setupHandler := handlers.NewSetupHandler(db, logger)
	requestLogsHandler := handlers.NewRequestLogsHandler(requestLogger, logger)
	monitoringHandler := handlers.NewMonitoringHandler(systemMonitor, logger)
	i18nHandler := handlers.NewI18nHandler(i18nService, logger)

	// Setup routes
	setupRoutes(router, messagesHandler, healthHandler, configHandler, usersHandler, authMiddleware, setupHandler, requestLogsHandler, monitoringHandler, i18nHandler)

	// Serve static files - use embed filesystem
	router.StaticFS("/static", http.FS(webfs.GetStaticFS()))
	router.GET("/favicon.ico", func(c *gin.Context) {
		favicon, err := webfs.GetFavicon()
		if err != nil {
			c.String(http.StatusNotFound, "Favicon not found")
			return
		}
		c.Data(http.StatusOK, "image/x-icon", favicon)
	})

	// Decide which page to serve based on admin user existence
	router.GET("/", func(c *gin.Context) {
		// Check if admin user exists
		adminExists, err := db.Client.User.Query().Where(user.Role("admin")).Exist(c.Request.Context())
		if err != nil {
			logger.WithError(err).Error("Failed to check admin user")
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}

		if adminExists {
			// Admin user exists, serve normal login page
			indexHTML, err := webfs.GetIndexHTML()
			if err != nil {
				logger.WithError(err).Error("Failed to read index.html")
				c.String(http.StatusInternalServerError, "Internal server error")
				return
			}
			c.Data(http.StatusOK, "text/html", indexHTML)
		} else {
			// No admin user, serve setup wizard page
			setupHTML, err := webfs.GetSetupHTML()
			if err != nil {
				logger.WithError(err).Error("Failed to read setup.html")
				c.String(http.StatusInternalServerError, "Internal server error")
				return
			}
			c.Data(http.StatusOK, "text/html", setupHTML)
		}
	})

	// Provide direct access to setup page (check if admin exists)
	router.GET("/setup", func(c *gin.Context) {
		// Check if admin user exists
		adminExists, err := db.Client.User.Query().Where(user.Role("admin")).Exist(c.Request.Context())
		if err != nil {
			logger.WithError(err).Error("Failed to check admin user")
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}

		if adminExists {
			// Admin user exists, redirect to homepage
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}

		// No admin user, show setup wizard page
		setupHTML, err := webfs.GetSetupHTML()
		if err != nil {
			logger.WithError(err).Error("Failed to read setup.html")
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}
		c.Data(http.StatusOK, "text/html", setupHTML)
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	logger.WithField("address", addr).Info("Starting server")

	// Start server in goroutine
	go func() {
		if err := router.Run(addr); err != nil {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for signal
	<-sigChan
	logger.Info("Received shutdown signal, shutting down gracefully...")

	// Setup timeout context for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Perform cleanup
	logger.Info("Stopping system monitor...")
	systemMonitor.Stop()

	select {
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout exceeded")
	default:
		logger.Info("Server shutdown complete")
	}
}

func printStartupInfo(cfg *config.Config, logger *logrus.Logger) {
	fmt.Printf("🚀 Claude-to-OpenAI API Proxy %s\n", Version)
	fmt.Printf("🏗️  Built at: %s\n", BuildTime)
	fmt.Println("✅ Configuration loaded from database")

	// API configuration status
	if cfg.OpenAIAPIKey != "" {
		fmt.Printf("   OpenAI API: ✅ Configured\n")
		fmt.Printf("   OpenAI Base URL: %s\n", cfg.OpenAIBaseURL)
	} else {
		fmt.Printf("   OpenAI API: ❌ Not configured\n")
	}

	if cfg.ClaudeAPIKey != "" {
		fmt.Printf("   Claude API: ✅ Configured\n")
		fmt.Printf("   Claude Base URL: %s\n", cfg.ClaudeBaseURL)
	} else {
		fmt.Printf("   Claude API: ❌ Not configured\n")
	}

	// Model configuration
	fmt.Printf("   Big Model: %s\n", cfg.BigModel)
	fmt.Printf("   Small Model: %s\n", cfg.SmallModel)

	// Performance configuration
	fmt.Printf("   Max Tokens Limit: %d\n", cfg.MaxTokensLimit)
	fmt.Printf("   Request Timeout: %ds\n", cfg.RequestTimeout)
	fmt.Printf("   Temperature: %.2f\n", cfg.Temperature)
	fmt.Printf("   Stream Enabled: %t\n", cfg.StreamEnabled)

	// Server configuration
	fmt.Printf("   Server: %s:%d\n", cfg.Host, cfg.Port)
	fmt.Printf("   Log Level: %s\n", cfg.LogLevel)

	// Configuration hints
	if cfg.OpenAIAPIKey == "" && cfg.ClaudeAPIKey == "" {
		fmt.Println("⚠️  Warning: No API keys configured. Please configure via web interface.")
		fmt.Printf("   Web Interface: http://%s:%d/\n", cfg.Host, cfg.Port)
	}

	fmt.Println()
}

func setupRoutes(router *gin.Engine, messagesHandler *handlers.MessagesHandler, healthHandler *handlers.HealthHandler, configHandler *handlers.ConfigHandler, usersHandler *handlers.UsersHandler, authMiddleware *middleware.AuthMiddleware, setupHandler *handlers.SetupHandler, requestLogsHandler *handlers.RequestLogsHandler, monitoringHandler *handlers.MonitoringHandler, i18nHandler *handlers.I18nHandler) {
	// API routes
	v1 := router.Group("/v1")
	{
		// Claude API routes
		v1.POST("/messages", messagesHandler.CreateMessage)
		v1.POST("/messages/count_tokens", messagesHandler.CountTokens)

		// OpenAI API compatible routes
		v1.POST("/chat/completions", messagesHandler.CreateChatCompletion)
	}

	// Health and utility routes
	router.GET("/health", healthHandler.Health)
	router.GET("/test-connection", healthHandler.TestConnection)
	router.GET("/api", healthHandler.Root) // API info endpoint
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":    Version,
			"build_time": BuildTime,
		})
	})

	// System monitoring routes (public access for health checks)
	router.GET("/health/live", monitoringHandler.GetLivenessProbe)
	router.GET("/health/ready", monitoringHandler.GetReadinessProbe)
	router.GET("/health/startup", monitoringHandler.GetStartupProbe)
	router.GET("/health/status", monitoringHandler.GetHealthStatus)
	router.GET("/health/detailed", monitoringHandler.GetDetailedHealth)

	// Setup routes (only available when no admin exists)
	api := router.Group("/api")
	{
		api.GET("/setup/check", setupHandler.CheckSetupRequired)
		api.POST("/setup/admin", setupHandler.SetupAdmin)
	}

	// Authentication routes
	auth := router.Group("/auth")
	{
		auth.POST("/login", usersHandler.Login)
		auth.POST("/logout", usersHandler.Logout)
		auth.GET("/me", authMiddleware.AuthRequired(), usersHandler.GetCurrentUser)
		auth.PUT("/password", authMiddleware.AuthRequired(), usersHandler.ChangePassword)
	}

	// Admin routes (require authentication and admin role)
	admin := router.Group("/admin")
	admin.Use(authMiddleware.AuthRequired())
	admin.Use(authMiddleware.AdminRequired())
	{
		// User management
		admin.GET("/users", usersHandler.GetAllUsers)
		admin.POST("/users", usersHandler.CreateUser)
		admin.GET("/users/:id", usersHandler.GetUser)
		admin.PUT("/users/:id", usersHandler.UpdateUser)
		admin.DELETE("/users/:id", usersHandler.DeleteUser)

		// Configuration management
		admin.GET("/config", configHandler.GetAllConfigs)
		admin.GET("/config/:key", configHandler.GetConfig)
		admin.PUT("/config/:key", configHandler.UpdateConfig)
		admin.POST("/config/test", configHandler.TestConfig)
		admin.POST("/config/test-api-key", configHandler.TestAPIKey)

		// Request logs management
		admin.GET("/logs", requestLogsHandler.GetRequestLogs)
		admin.GET("/logs/stats", requestLogsHandler.GetRequestLogStats)
		admin.GET("/logs/dashboard", requestLogsHandler.GetDashboardData)
		admin.GET("/logs/:id", requestLogsHandler.GetRequestLogDetails)
		admin.DELETE("/logs/cleanup", requestLogsHandler.DeleteOldLogs)

		// System monitoring (admin only)
		admin.GET("/monitoring/metrics", monitoringHandler.GetSystemMetrics)
		admin.GET("/monitoring/info", monitoringHandler.GetSystemInfo)
	}

	// I18n routes (public access)
	i18n := router.Group("/i18n")
	{
		i18n.GET("/languages", i18nHandler.GetLanguages)
		i18n.GET("/messages/:lang", i18nHandler.GetMessages)
		i18n.GET("/current", i18nHandler.GetCurrentLanguage)
		i18n.POST("/language", i18nHandler.SetLanguage)
	}
}
