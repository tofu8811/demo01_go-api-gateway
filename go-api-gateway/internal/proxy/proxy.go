package proxy

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Public(targetURL string, method string) fiber.Handler {
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

func WithGateway(targetURL string, method string, gatewaySecret string) fiber.Handler {
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