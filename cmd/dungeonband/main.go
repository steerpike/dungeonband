// Package main is the entry point for DungeonBand.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/samdwyer/dungeonband/internal/game"
	"github.com/samdwyer/dungeonband/internal/telemetry"
)

func main() {
	// Load .env file for local development
	// This makes HONEYCOMB_DUNGEONBAND_API_KEY available
	if err := godotenv.Load(); err != nil {
		// Not fatal - env vars might be set directly
		log.Printf("Note: .env file not loaded: %v", err)
	}

	// Set up OTEL environment variables from our .env variables
	setupOTelEnv()

	ctx := context.Background()

	// Initialize telemetry
	shutdown, err := telemetry.Setup(ctx)
	if err != nil {
		log.Printf("Warning: telemetry setup failed: %v", err)
		log.Printf("Game will run without observability")
		// Continue without telemetry - game still works
	} else {
		defer func() {
			if err := shutdown(ctx); err != nil {
				log.Printf("Error shutting down telemetry: %v", err)
			}
		}()
	}

	// Create and run game
	g, err := game.New()
	if err != nil {
		log.Fatalf("Failed to initialize game: %v", err)
	}

	if err := g.Run(ctx); err != nil {
		log.Fatalf("Game error: %v", err)
	}
}

// setupOTelEnv configures OTEL environment variables from our custom env vars.
func setupOTelEnv() {
	// Always set endpoint to Honeycomb
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://api.honeycomb.io")

	// Always set headers from our API key - the .env file may have an unexpanded
	// variable reference that doesn't work, so we construct it properly here
	apiKey := os.Getenv("HONEYCOMB_DUNGEONBAND_API_KEY")
	dataset := os.Getenv("HONEYCOMB_DUNGEONBAND_DATASET")
	if dataset == "" {
		dataset = "dungeonband" // default dataset name
	}
	if apiKey != "" {
		os.Setenv("OTEL_EXPORTER_OTLP_HEADERS",
			fmt.Sprintf("x-honeycomb-team=%s,x-honeycomb-dataset=%s", apiKey, dataset))
	}
}
