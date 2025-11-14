package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/segmentio/kafka-go"

	"silentraven/internal/models"
	"silentraven/pkg/config"
)

type Gateway struct {
	config *config.Config
	writer *kafka.Writer
	router *mux.Router
}

func main() {
	log.Println("🚀 Starting SilentRaven Gateway Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Create gateway instance
	gateway := NewGateway(cfg)
	defer gateway.Close()

	// Setup routes
	gateway.setupRoutes()

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetAPIAddress(),
		Handler:      corsHandler.Handler(gateway.router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("✅ Gateway listening on %s", cfg.GetAPIAddress())
		log.Println("📡 Ready to receive drone detections")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down gateway...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Gateway stopped gracefully")
}

// NewGateway creates a new gateway instance
func NewGateway(cfg *config.Config) *Gateway {
	// Create Kafka writer for Redpanda
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.KafkaBrokers),
		Topic:        cfg.KafkaTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    10,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	log.Printf("✅ Connected to Redpanda at %s", cfg.KafkaBrokers)

	return &Gateway{
		config: cfg,
		writer: writer,
		router: mux.NewRouter(),
	}
}

// Close closes all connections
func (g *Gateway) Close() {
	if g.writer != nil {
		g.writer.Close()
	}
}

// setupRoutes configures HTTP routes
func (g *Gateway) setupRoutes() {
	// Health check
	g.router.HandleFunc("/health", g.handleHealth).Methods("GET")

	// Receive drone detection
	g.router.HandleFunc("/api/v1/detection", g.handleDetection).Methods("POST")

	// Test endpoint
	g.router.HandleFunc("/api/v1/test", g.handleTest).Methods("GET")
}

// handleHealth returns service health status
func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := models.APIResponse{
		Success: true,
		Message: "Gateway service is healthy",
		Data: map[string]string{
			"service": "gateway",
			"status":  "running",
			"version": "1.0.0",
		},
	}
	sendJSON(w, http.StatusOK, response)
}

// handleTest returns test response
func (g *Gateway) handleTest(w http.ResponseWriter, r *http.Request) {
	response := models.APIResponse{
		Success: true,
		Message: "Test endpoint working",
		Data: map[string]string{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
	sendJSON(w, http.StatusOK, response)
}

// handleDetection processes incoming drone detection
func (g *Gateway) handleDetection(w http.ResponseWriter, r *http.Request) {
	// Parse incoming packet
	var packet models.IncomingPacket
	if err := json.NewDecoder(r.Body).Decode(&packet); err != nil {
		log.Printf("❌ Invalid JSON: %v", err)
		response := models.APIResponse{
			Success: false,
			Error:   "Invalid JSON format",
		}
		sendJSON(w, http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if packet.SN == "" || packet.UASID == "" {
		log.Println("❌ Missing required fields")
		response := models.APIResponse{
			Success: false,
			Error:   "Missing required fields: SN or UASID",
		}
		sendJSON(w, http.StatusBadRequest, response)
		return
	}

	// Add timestamp if not present
	if packet.Timestamp == "" {
		packet.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Log received detection
	log.Printf("📡 Received detection: UASID=%s, SN=%s, Type=%s",
		packet.UASID, packet.SN, packet.DroneType)

	// Convert to JSON for Kafka
	packetJSON, err := json.Marshal(packet)
	if err != nil {
		log.Printf("❌ Failed to marshal packet: %v", err)
		response := models.APIResponse{
			Success: false,
			Error:   "Internal processing error",
		}
		sendJSON(w, http.StatusInternalServerError, response)
		return
	}

	// Publish to Redpanda
	err = g.writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(packet.UASID),
			Value: packetJSON,
			Time:  time.Now(),
		},
	)

	if err != nil {
		log.Printf("❌ Failed to publish to Redpanda: %v", err)
		response := models.APIResponse{
			Success: false,
			Error:   "Failed to queue message",
		}
		sendJSON(w, http.StatusInternalServerError, response)
		return
	}

	log.Printf("✅ Published to Redpanda: %s", packet.UASID)

	// Send success response
	response := models.APIResponse{
		Success: true,
		Message: "Detection received and queued",
		Data: map[string]string{
			"uas_id":    packet.UASID,
			"sn":        packet.SN,
			"timestamp": packet.Timestamp,
		},
	}
	sendJSON(w, http.StatusOK, response)
}

// sendJSON sends JSON response
func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
