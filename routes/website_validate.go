package routes

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
	"strings"
	"time"
)

func WebsiteValidateHandler(c *fiber.Ctx) error {
	queryURL := c.Query("url")
	if queryURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing url query param"})
	}

	client := http.Client{Timeout: 3 * time.Second}

	// First try HEAD request
	resp, err := client.Head(queryURL)
	if err == nil && resp.StatusCode == 200 {
		return c.JSON(fiber.Map{"valid": true})
	}

	// Fallback to GET if HEAD failed or blocked
	resp, err = client.Get(queryURL)
	if err != nil {
		return c.JSON(fiber.Map{"valid": false, "reason": err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.JSON(fiber.Map{"valid": false, "reason": "non-200 status"})
	}

	// Quick check content-length or minimal body read
	if resp.ContentLength > 0 {
		return c.JSON(fiber.Map{"valid": true})
	}

	// Read tiny bit of body just to confirm non-empty
	buf := make([]byte, 512)
	n, _ := resp.Body.Read(buf)
	if n > 0 && len(strings.TrimSpace(string(buf[:n]))) > 0 {
		return c.JSON(fiber.Map{"valid": true})
	}

	return c.JSON(fiber.Map{"valid": false, "reason": "empty body"})
}
