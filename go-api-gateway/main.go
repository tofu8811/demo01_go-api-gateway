package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	app := fiber.New()

	authURL := os.Getenv("AUTH_SERVICE_URL")
	productURL := os.Getenv("PRODUCT_SERVICE_URL")
	orderURL := os.Getenv("ORDER_SERVICE_URL")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Go Fiber API Gateway is running",
		})
	})

	// Auth routes
	app.Post("/api/user/register", proxy(authURL+"/user/register", http.MethodPost))
	app.Post("/api/auth/login", proxy(authURL+"/auth/login", http.MethodPost))

	app.Get("/api/user/profile", proxy(authURL+"/user/profile", http.MethodGet))
	app.Put("/api/user/update", proxy(authURL+"/user/update", http.MethodPut))
	app.Put("/api/auth/token-refresh", proxy(authURL+"/auth/token-refresh", http.MethodPut))
	app.Delete("/api/auth/token-revoke", proxy(authURL+"/auth/token-revoke", http.MethodDelete))

	// Product routes
	app.Get("/api/products", proxy(productURL+"/products", http.MethodGet))
	app.Get("/api/product/:id", func(c *fiber.Ctx) error {
		return proxy(productURL+"/product/"+c.Params("id"), http.MethodGet)(c)
	})
	app.Post("/api/product/create", proxy(productURL+"/product/create", http.MethodPost))
	app.Put("/api/product/update/:id", func(c *fiber.Ctx) error {
		return proxy(productURL+"/product/update/"+c.Params("id"), http.MethodPut)(c)
	})
	app.Delete("/api/product/delete/:id", func(c *fiber.Ctx) error {
		return proxy(productURL+"/product/delete/"+c.Params("id"), http.MethodDelete)(c)
	})

	// Order routes
	app.Get("/api/orders", proxy(orderURL+"/orders", http.MethodGet))
	app.Get("/api/order/:id", func(c *fiber.Ctx) error {
		return proxy(orderURL+"/order/"+c.Params("id"), http.MethodGet)(c)
	})
	app.Post("/api/order/create", proxy(orderURL+"/order/create", http.MethodPost))
	app.Put("/api/order/update/:id", func(c *fiber.Ctx) error {
		return proxy(orderURL+"/order/update/"+c.Params("id"), http.MethodPut)(c)
	})
	app.Delete("/api/order/delete/:id", func(c *fiber.Ctx) error {
		return proxy(orderURL+"/order/delete/"+c.Params("id"), http.MethodDelete)(c)
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "4000"
	}

	log.Fatal(app.Listen(":" + port))
}

func proxy(targetURL string, method string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		var body io.Reader
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
			body = bytes.NewReader(c.Body())
		}

		req, err := http.NewRequest(method, targetURL, body)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to create request",
				"error":   err.Error(),
			})
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Forward Authorization header nếu Postman có gửi token
		if auth := c.Get("Authorization"); auth != "" {
			req.Header.Set("Authorization", auth)
		}

		resp, err := client.Do(req)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{
				"status":  "error",
				"message": "Service unavailable",
				"target":  targetURL,
				"error":   err.Error(),
			})
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to read response",
				"error":   err.Error(),
			})
		}

		c.Set("Content-Type", "application/json")
		return c.Status(resp.StatusCode).Send(respBody)
	}
}