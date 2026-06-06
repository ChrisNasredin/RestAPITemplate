package main

import (
	"HiveAPI/internal/config"
	"log"
)

func main() {
	cfg := config.MustLoad()
	log.Printf("config: %+v", cfg)
}
