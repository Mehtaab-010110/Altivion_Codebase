package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"silentraven/internal/database"
	"silentraven/internal/models"
	"silentraven/pkg/config"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

type IngestionService struct {
	config *config.Config
	db     *database.DB
	reader *kafka.Reader
}

func main() {
	log.Println("🚀 Starting SilentRaven Ingestion Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create ingestion service
	service := NewIngestionService(cfg, db)
	defer service.Close()

	log.Println("✅ Ingestion service started successfully")
	log.Println("📡 Listening for drone detections from Redpanda...")

	// Start consuming messages
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n🛑 Shutdown signal received...")
		cancel()
	}()

	// Start processing
	if err := service.ProcessMessages(ctx); err != nil {
		log.Fatal("Processing failed:", err)
	}

	log.Println("✅ Ingestion service stopped gracefully")
}

// NewIngestionService creates a new ingestion service
func NewIngestionService(cfg *config.Config, db *database.DB) *IngestionService {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cfg.KafkaBrokers},
		Topic:          cfg.KafkaTopic,
		GroupID:        "silentraven-ingestion",
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	log.Printf("✅ Connected to Redpanda topic: %s", cfg.KafkaTopic)

	return &IngestionService{
		config: cfg,
		db:     db,
		reader: reader,
	}
}

// Close closes all connections
func (s *IngestionService) Close() {
	if s.reader != nil {
		s.reader.Close()
	}
}

// ProcessMessages reads and processes messages from Redpanda
func (s *IngestionService) ProcessMessages(ctx context.Context) error {
	messageCount := 0

	for {
		select {
		case <-ctx.Done():
			log.Printf("📊 Processed %d messages total", messageCount)
			return nil
		default:
			// Read message with timeout
			m, err := s.reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				log.Printf("❌ Error fetching message: %v", err)
				time.Sleep(time.Second)
				continue
			}

			// Process the message
			if err := s.processMessage(m); err != nil {
				log.Printf("❌ Error processing message: %v", err)
			} else {
				messageCount++
			}

			// Commit the message
			if err := s.reader.CommitMessages(ctx, m); err != nil {
				log.Printf("⚠️  Failed to commit message: %v", err)
			}
		}
	}
}

// processMessage processes a single message
func (s *IngestionService) processMessage(m kafka.Message) error {
	// Parse incoming packet
	var packet models.IncomingPacket
	if err := json.Unmarshal(m.Value, &packet); err != nil {
		return err
	}

	log.Printf("📥 Processing: UASID=%s, SN=%s, Type=%s",
		packet.UASID, packet.SN, packet.DroneType)

	// Convert to database model
	detection := &models.DroneDetection{
		DetectionTime:     time.Now(),
		SN:                packet.SN,
		UASID:             packet.UASID,
		DroneType:         packet.DroneType,
		Latitude:          packet.Latitude,
		Longitude:         packet.Longitude,
		Height:            packet.Height,
		Direction:         packet.Direction,
		SpeedHorizontal:   packet.SpeedHorizontal,
		SpeedVertical:     packet.SpeedVertical,
		OperatorLatitude:  packet.OperatorLatitude,
		OperatorLongitude: packet.OperatorLongitude,
		NodeID:            packet.NodeID,
		Signature:         packet.Signature,
	}

	// Parse timestamp if provided
	if packet.Timestamp != "" {
		if ts, err := time.Parse(time.RFC3339, packet.Timestamp); err == nil {
			detection.DetectionTime = ts
		}
	}

	// Store raw JSON data
	rawJSON, _ := json.Marshal(packet)
	detection.RawData = string(rawJSON)

	// Insert into database
	if err := s.db.InsertDroneDetection(detection); err != nil {
		return err
	}

	log.Printf("✅ Stored detection ID=%d, UASID=%s", detection.ID, detection.UASID)
	return nil
}
