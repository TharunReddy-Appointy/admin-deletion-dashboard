package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"go.appointy.com/admin-deletion-dashboard/internal/auth"
	"go.appointy.com/admin-deletion-dashboard/internal/handler"
	"go.appointy.com/admin-deletion-dashboard/internal/service"
)

//go:embed web/*
var webFiles embed.FS

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	log.Println("Loading configuration...")
	// Get configuration from environment
	config := loadConfig()
	log.Printf("Config loaded - Port: %s, Environment: %s", config.Port, config.Environment)

	log.Println("Connecting to database...")
	// Initialize database
	db, err := initDatabase(config.DatabaseURL)
	if err != nil {
		log.Printf("⚠️ WARNING: Failed to connect to database: %v", err)
		log.Println("⚠️ Starting server WITHOUT database connection")
		db = nil // Continue without database
	} else {
		log.Println("✅ Database connected successfully")
		defer db.Close()
	}

	// Initialize services
	authConfig := auth.NewAuthConfig(
		config.GoogleClientID,
		config.GoogleClientSecret,
		config.GoogleRedirectURL,
		config.JWTSecret,
	)

	accountService := service.NewAccountService(db)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authConfig)
	accountHandler := handler.NewAccountHandler(accountService)

	// Setup router
	router := setupRouter(authConfig, authHandler, accountHandler)

	// Start server
	addr := fmt.Sprintf(":%s", config.Port)
	log.Printf("🚀 Server starting on %s", addr)
	log.Printf("📱 Dashboard: http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// Config holds application configuration
type Config struct {
	Port               string
	DatabaseURL        string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	JWTSecret          string
	Environment        string
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/callback"),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Environment:        getEnv("ENVIRONMENT", "development"),
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// initDatabase initializes database connection
func initDatabase(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	log.Printf("Opening database connection...")
	log.Printf("Database host: %s", maskPassword(databaseURL))

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Printf("❌ Failed to open database: %v", err)
		return nil, err
	}

	// Set connection timeouts and limits
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(5)

	log.Printf("Pinging database to test connection (with 15s timeout)...")
	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Printf("❌ Failed to ping database: %v", err)
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	log.Println("✅ Database connected successfully")
	return db, nil
}

// maskPassword masks the password in the connection string for logging
func maskPassword(connStr string) string {
	// Simple masking: postgres://user:password@host:port/db -> postgres://user:***@host:port/db
	if idx := strings.Index(connStr, "://"); idx != -1 {
		afterProto := connStr[idx+3:]
		if idx2 := strings.Index(afterProto, ":"); idx2 != -1 {
			if idx3 := strings.Index(afterProto[idx2+1:], "@"); idx3 != -1 {
				return connStr[:idx+3+idx2+1] + "***" + afterProto[idx2+1+idx3:]
			}
		}
	}
	return connStr
}

// setupRouter sets up the Gin router with all routes
func setupRouter(authConfig *auth.Config, authHandler *handler.AuthHandler, accountHandler *handler.AccountHandler) *gin.Engine {
	// Set Gin mode based on environment
	if getEnv("ENVIRONMENT", "development") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS middleware for development
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (public)
		authRoutes := api.Group("/auth")
		{
			authRoutes.GET("/login", authHandler.HandleLogin)
			authRoutes.GET("/callback", authHandler.HandleCallback)
			authRoutes.POST("/logout", authHandler.HandleLogout)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(authConfig.AuthMiddleware())
		{
			protected.GET("/auth/me", authHandler.HandleMe)
			protected.POST("/account/lookup", accountHandler.HandleLookup)
			protected.POST("/account/delete", accountHandler.HandleDelete)
			protected.GET("/account/audit-logs", accountHandler.HandleGetAuditLogs)
		}
	}

	// Serve embedded web files
	webFS, err := fs.Sub(webFiles, "web")
	if err != nil {
		log.Fatal("Failed to load web files:", err)
	}

	// Serve static files
	router.GET("/", func(c *gin.Context) {
		c.FileFromFS("/", http.FS(webFS))
	})

	router.NoRoute(func(c *gin.Context) {
		c.FileFromFS("/", http.FS(webFS))
	})

	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
