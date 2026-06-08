package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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
	gatewaySecret := os.Getenv("GATEWAY_SECRET")

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
	app.Get("/api/products", protected(authURL, proxyWithGateway(productURL+"/products", http.MethodGet, gatewaySecret)))
	app.Get("/api/product/:id", protected(authURL, func(c *fiber.Ctx) error {
		return proxyWithGateway(productURL+"/product/"+c.Params("id"), http.MethodGet, gatewaySecret)(c)
	}))
	app.Post("/api/product/create", protected(authURL, proxyWithGateway(productURL+"/product/create", http.MethodPost, gatewaySecret)))
	app.Put("/api/product/update/:id", protected(authURL, func(c *fiber.Ctx) error {
		return proxyWithGateway(productURL+"/product/update/"+c.Params("id"), http.MethodPut, gatewaySecret)(c)
	}))
	app.Delete("/api/product/delete/:id", protected(authURL, func(c *fiber.Ctx) error {
		return proxyWithGateway(productURL+"/product/delete/"+c.Params("id"), http.MethodDelete, gatewaySecret)(c)
	}))

	// Order routes
	app.Get("/api/orders", protected(authURL, proxyWithGateway(orderURL+"/orders", http.MethodGet, gatewaySecret)))
	app.Get("/api/order/:id", protected(authURL, func(c *fiber.Ctx) error {
		return proxyWithGateway(orderURL+"/order/"+c.Params("id"), http.MethodGet, gatewaySecret)(c)
	}))
	app.Post("/api/order/create", protected(authURL, proxyWithGateway(orderURL+"/order/create", http.MethodPost, gatewaySecret)))
	app.Put("/api/order/update/:id", protected(authURL, func(c *fiber.Ctx) error {
		return proxyWithGateway(orderURL+"/order/update/"+c.Params("id"), http.MethodPut, gatewaySecret)(c)
	}))
	app.Delete("/api/order/delete/:id", protected(authURL, func(c *fiber.Ctx) error {
		return proxyWithGateway(orderURL+"/order/delete/"+c.Params("id"), http.MethodDelete, gatewaySecret)(c)
	}))

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

func protected(authURL string, next fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"status":  "error",
				"message": "Missing Authorization header",
			})
		}

		req, err := http.NewRequest(http.MethodGet, authURL+"/user/profile", nil)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to create auth request",
			})
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", authHeader)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{
				"status":  "error",
				"message": "Auth service unavailable",
			})
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			c.Set("Content-Type", "application/json")
			return c.Status(resp.StatusCode).Send(respBody)
		}

		var authResponse map[string]interface{}
		if err := json.Unmarshal(respBody, &authResponse); err == nil {
			c.Locals("auth_user", authResponse["data"])
		}

		return next(c)
	}
}

func proxyWithGateway(targetURL string, method string, gatewaySecret string) fiber.Handler {
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
		req.Header.Set("X-Gateway-Secret", gatewaySecret)

		if auth := c.Get("Authorization"); auth != "" {
			req.Header.Set("Authorization", auth)
		}

		if user, ok := c.Locals("auth_user").(map[string]interface{}); ok {
			if id, exists := user["id"]; exists {
				req.Header.Set("X-User-ID", toString(id))
			}
			if email, exists := user["email"]; exists {
				req.Header.Set("X-User-Email", toString(email))
			}
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

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.Itoa(int(v))
	default:
		return ""
	}
}
