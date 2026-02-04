package main

// cmd entrypoint for web-ui

import (
	"flag"
	"log"
)

func main() {
	configPath := flag.String("config", "./config.yaml", "Path to config file")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	server, err := NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
