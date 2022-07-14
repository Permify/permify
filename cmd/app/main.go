package main

import (
	"log"

	"github.com/Permify/permify/internal/app"
	"github.com/Permify/permify/internal/config"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
