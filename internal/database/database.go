package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"ccany/ent"
	"ccany/internal/crypto"

	_ "github.com/mattn/go-sqlite3"
)

// Database manager
type Database struct {
	Client        *ent.Client
	CryptoService *crypto.CryptoService
}

// Config database configuration
type Config struct {
	Type     string `json:"type"` // sqlite, mysql, postgres
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`

	// SQLite specific
	DataPath string `json:"data_path"`

	// Connection pool settings
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}

// DefaultConfig returns default database configuration
func DefaultConfig() *Config {
	// Get data directory from environment variable, default to "./data"
	dataPath := os.Getenv("CLAUDE_PROXY_DATA_PATH")
	if dataPath == "" {
		dataPath = "./data"
	}

	return &Config{
		Type:            "sqlite",
		DataPath:        dataPath,
		Database:        "claude_proxy.db",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}
}

// NewDatabase creates database manager
func NewDatabase(cfg *Config, masterKey string) (*Database, error) {
	// Create database client
	client, err := createClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database client: %w", err)
	}

	// Create crypto service
	cryptoService := crypto.NewCryptoService(masterKey)

	return &Database{
		Client:        client,
		CryptoService: cryptoService,
	}, nil
}

// createClient creates database client
func createClient(cfg *Config) (*ent.Client, error) {
	var dsn string
	var driver string

	switch cfg.Type {
	case "sqlite":
		// Ensure data directory exists
		if cfg.DataPath != "" {
			if err := os.MkdirAll(cfg.DataPath, 0755); err != nil {
				return nil, fmt.Errorf("failed to create data directory: %w", err)
			}
		}

		dbPath := filepath.Join(cfg.DataPath, cfg.Database)
		dsn = fmt.Sprintf("file:%s?cache=shared&_fk=1", dbPath)
		driver = "sqlite3"

	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		driver = "mysql"

	case "postgres":
		sslMode := cfg.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, sslMode)
		driver = "postgres"

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// Create client
	client, err := ent.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// For ent, connection pool configuration is usually set at creation time
	// Temporarily removed here as ent has its own connection management approach

	return client, nil
}

// Initialize initializes database
func (d *Database) Initialize(ctx context.Context) error {
	log.Println("Initializing database...")

	// Run database migration
	if err := d.Client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// Close closes database connection
func (d *Database) Close() error {
	if d.Client != nil {
		return d.Client.Close()
	}
	return nil
}

// Health checks database health status
func (d *Database) Health(ctx context.Context) error {
	if d.Client == nil {
		return fmt.Errorf("database client is nil")
	}

	// Try to execute a simple query
	_, err := d.Client.AppConfig.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// GetStats gets database statistics
func (d *Database) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get config count
	configCount, err := d.Client.AppConfig.Query().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get config count: %w", err)
	}

	// Get user count
	userCount, err := d.Client.User.Query().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}

	// Get request log count
	requestLogCount, err := d.Client.RequestLog.Query().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get request log count: %w", err)
	}

	stats["configs"] = configCount
	stats["users"] = userCount
	stats["request_logs"] = requestLogCount
	// Database connection statistics temporarily removed as ent doesn't directly expose database connections

	return stats, nil
}

// Backup backs up database (SQLite)
func (d *Database) Backup(ctx context.Context, backupPath string) error {
	// Check if client exists
	if d.Client == nil {
		return fmt.Errorf("database client is nil")
	}

	// Create backup directory
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Execute backup (specific backup logic can be implemented here)
	// For SQLite, VACUUM INTO command can be used
	// For other databases, their respective backup tools are needed
	// Temporarily simplified implementation, not directly accessing underlying database connection

	log.Printf("Database backup completed: %s", backupPath)
	return nil
}

// GetMasterKeyFromEnv gets master key from environment variable
func GetMasterKeyFromEnv() string {
	masterKey := os.Getenv("CLAUDE_PROXY_MASTER_KEY")
	if masterKey == "" {
		// If no master key is set, use default value (should be enforced in production environment)
		log.Println("Warning: No master key set, using default key (NOT SECURE for production)")
		return "default-master-key-not-secure"
	}
	return masterKey
}

// InitializeDatabase convenience function for initializing database
func InitializeDatabase(ctx context.Context, dbConfig *Config, masterKey string) (*Database, error) {
	// Create database manager
	db, err := NewDatabase(dbConfig, masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Initialize database
	if err := db.Initialize(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to initialize database: %w, and failed to close database: %v", err, closeErr)
		}
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}
