package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ArminDashti/radar-agent/internal/config"
	"github.com/ArminDashti/radar-agent/internal/hub"
	agentloop "github.com/ArminDashti/radar-agent/internal/loop"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := hub.NewClient(cfg.HubURL, cfg.AgentToken, cfg.HTTPTimeout+cfg.ICMPTimeout)
	log.Printf("radar-agent connected to %s", cfg.HubURL)
	if err := agentloop.New(client, cfg.HTTPTimeout, cfg.ICMPTimeout).Run(ctx); err != nil {
		log.Fatal(err)
	}
	log.Print("radar-agent stopped")
}
