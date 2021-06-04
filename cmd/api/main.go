package main

import (
	"log"

	"ffxiv.anid.dev/internal/config"
)

func main() {
	r, err := InitializeServer(&config.DefaultConfig)
	if err != nil {
		log.Fatalf("failed to init server: %s", err)
	}

	r.Start()
}
