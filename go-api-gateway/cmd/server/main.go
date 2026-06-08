package main

import (
	"log"

	"go-api-gateway/internal/config"
	"go-api-gateway/internal/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.Load()

	app := fiber.New()

	routes.Register(app, cfg)

	log.Fatal(app.Listen(":" + cfg.AppPort))
}