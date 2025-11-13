package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	// Kafka/Redpanda
	KafkaBrokers string
	KafkaTopic   string

	// API
	APIPort   string
	APISecret string

	// Security
	CertPath       string
	CACertFile     string
	ServerCertFile string
	ServerKeyFile  string

	// Logging
	LogLevel string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file (optional in production)
	_ = godotenv.Load()

	config := &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "silentraven"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// Kafka/Redpanda
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "drone-detections"),

		// API
		APIPort:   getEnv("API_PORT", "8080"),
		APISecret: getEnv("API_SECRET", ""),

		// Security
		CertPath:       getEnv("CERT_PATH", "./certs"),
		CACertFile:     getEnv("CA_CERT_FILE", "ca.crt"),
		ServerCertFile: getEnv("SERVER_CERT_FILE", "server.crt"),
		ServerKeyFile:  getEnv("SERVER_KEY_FILE", "server.key"),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	// Validate required fields
	if config.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}
	if config.APISecret == "" {
		return nil, fmt.Errorf("API_SECRET is required")
	}

	return config, nil
}

// GetDBConnectionString returns PostgreSQL connection string
func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// GetAPIAddress returns the full API listen address
func (c *Config) GetAPIAddress() string {
	return ":" + c.APIPort
}

// getEnv reads environment variable with fallback default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
