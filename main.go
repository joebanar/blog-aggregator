package main

import (
	"fmt"
	"log"
	"os"

	"gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("read config: %v", err)
	}

	user := os.Getenv("USER")
	if user == "" {
		user = "joe"
	}

	if err := cfg.SetUser(user); err != nil {
		log.Fatalf("set user: %v", err)
	}

	cfg2, err := config.Read()
	if err != nil {
		log.Fatalf("read config second time: %v", err)
	}

	fmt.Printf("%+v\n", cfg2)
}
