package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Protected(authURL string) fiber.Handler {
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

		return c.Next()
	}
}
