package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
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

	// Create Redpanda consumer
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.KafkaBrokers},
		Topic:    cfg.KafkaTopic,
		GroupID:  "cot-publisher",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	// Create appropriate sender based on TAK_MODE
	var sender cot.Sender
	takMode := os.Getenv("TAK_MODE")

	switch takMode {
	case "tcp":
		serverIP := os.Getenv("TAK_SERVER_IP")
		if serverIP == "" {
			log.Fatal("TAK_SERVER_IP not set in .env")
		}

		portStr := os.Getenv("TAK_SERVER_PORT")
		if portStr == "" {
			portStr = "8088"
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Fatal("Invalid TAK_SERVER_PORT:", err)
		}

		sender, err = cot.NewTCPSender(serverIP, port)
		if err != nil {
			log.Fatal("TCP sender failed:", err)
		}
		log.Printf("✅ Connected to TAK Server: %s:%d (TCP)", serverIP, port)

	case "direct":
		targetIP := os.Getenv("TAK_TARGET_IP")
		if targetIP == "" {
			log.Fatal("TAK_TARGET_IP not set in .env")
		}

		portStr := os.Getenv("TAK_TARGET_PORT")
		if portStr == "" {
			portStr = "6969"
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Fatal("Invalid TAK_TARGET_PORT:", err)
		}

		sender, err = cot.NewDirectSender(targetIP, port)
		if err != nil {
			log.Fatal("Direct sender failed:", err)
		}
		log.Printf("✅ Sending to direct IP: %s:%d (UDP)", targetIP, port)

	case "multicast":
		sender, err = cot.NewMulticastSender()
		if err != nil {
			log.Fatal("Multicast sender failed:", err)
		}
		log.Println("✅ Sending to multicast: 239.2.3.1:6969 (UDP)")

	default:
		log.Fatal("TAK_MODE must be 'tcp', 'direct', or 'multicast' (check .env file)")
	}
	defer sender.Close()

	log.Println("📡 Listening for detections on Redpanda...")

	// Process messages
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("🛑 Shutting down...")
		cancel()
	}()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("Fetch error: %v", err)
			continue
		}

		var detection models.IncomingPacket
		if err := json.Unmarshal(msg.Value, &detection); err != nil {
			log.Printf("Parse error: %v", err)
			reader.CommitMessages(ctx, msg)
			continue
		}

		cotXML, err := cot.ConvertToCoT(detection)
		if err != nil {
			log.Printf("CoT conversion error: %v", err)
			reader.CommitMessages(ctx, msg)
			continue
		}

		if err := sender.Send(cotXML); err != nil {
			log.Printf("❌ TAK send error: %v", err)
		} else {
			log.Printf("✅ Sent to TAK: UAS=%s (Mode: %s)", detection.UASID, takMode)
		}

		reader.CommitMessages(ctx, msg)
	}

	log.Println("✅ CoT Publisher stopped")
}
