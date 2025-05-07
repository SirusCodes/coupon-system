package config

import (
	"os"
)

type Config struct {
	DatabasePath    string
	ServerPort      string
	CacheSize       int
	CacheTTLMinutes int
}

func LoadConfig() (*Config, error) {
	// Example: Load from environment variables
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./coupons.db" // Default path
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080" // Default port
	}

	cacheSize := 100 // Default cache size
	cacheTTL := 10   // Default cache TTL in minutes

	return &Config{
		DatabasePath:    dbPath,
		ServerPort:      serverPort,
		CacheSize:       cacheSize,
		CacheTTLMinutes: cacheTTL,
	}, nil
}
