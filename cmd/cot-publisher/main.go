package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"

	"silentraven/internal/cot"
	"silentraven/internal/models"
	"silentraven/pkg/config"
)

func main() {
	log.Println("🚀 Starting CoT Publisher Service...")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config load failed:", err)
	}

	// Create Redpanda consumer (same topic as Ingestion)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.KafkaBrokers},
		Topic:    cfg.KafkaTopic, // "drone-detections"
		GroupID:  "cot-publisher",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	// Create CoT sender (UDP multicast for now)
	sender, err := cot.NewMulticastSender()
	if err != nil {
		log.Fatal("CoT sender failed:", err)
	}
	defer sender.Close()

	log.Println("✅ CoT Publisher started")
	log.Println("📡 Listening for detections on Redpanda...")

	// Process messages
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("🛑 Shutting down...")
		cancel()
	}()

	// Main loop
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break // Context cancelled
			}
			log.Printf("Fetch error: %v", err)
			continue
		}

		// Parse detection
		var detection models.DroneDetection
		if err := json.Unmarshal(msg.Value, &detection); err != nil {
			log.Printf("Parse error: %v", err)
			reader.CommitMessages(ctx, msg)
			continue
		}

		// Convert to CoT
		cotXML, err := cot.ConvertToCoT(detection)
		if err != nil {
			log.Printf("CoT conversion error: %v", err)
			reader.CommitMessages(ctx, msg)
			continue
		}

		// Send to TAK
		if err := sender.Send(cotXML); err != nil {
			log.Printf("TAK send error: %v", err)
		} else {
			log.Printf("✅ Sent to TAK: UAS=%s", detection.UASID)
		}

		reader.CommitMessages(ctx, msg)
	}

	log.Println("✅ CoT Publisher stopped")
}
