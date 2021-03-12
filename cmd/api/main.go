package main

import (
	"ffxiv.anid.dev/internal/config"
)

func main() {
	r, _ := InitializeServer(&config.DefaultConfig)

	r.Start()
}
