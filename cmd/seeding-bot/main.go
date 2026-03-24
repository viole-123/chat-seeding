package main

import (
	"fmt"
	"log"

	"uniscore-seeding-bot/internal/app"
)

func main() {
	if err := app.Bootstrap(); err != nil {
		log.Fatalf("app bootstrap failed: %v", err)
	}

	fmt.Println("ok")
}
