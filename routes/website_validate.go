package routes

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Create a shared HTTP client with:
// - Timeout
// - TLS config that disables HTTP/2 (forces HTTP/1.1)
// - Custom User-Agent
var transport = &http.Transport{
	TLSClientConfig: &tls.Config{
		NextProtos: []string{"http/1.1"}, // 🔒 disable HTTP/2
	},
	DialContext: (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 5 * time.Second,
	}).DialContext,
	MaxIdleConns:          10,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   5 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

var client = &http.Client{
	Timeout:   5 * time.Second,
	Transport: transport,
}

func makeRequest(method string, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Set User-Agent like a real browser
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")

	return client.Do(req)
}

func WebsiteValidateHandler(c *fiber.Ctx) error {
	queryURL := c.Query("url")
	if queryURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing ?url= query parameter"})
	}

	// Step 1: Try HEAD request
	resp, err := makeRequest("HEAD", queryURL)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK && resp.ContentLength > 0 {
			return c.JSON(fiber.Map{"valid": true, "method": "HEAD"})
		}
	}

	// Step 2: Fallback to GET
	resp, err = makeRequest("GET", queryURL)
	if err != nil {
		return c.JSON(fiber.Map{"valid": false, "method": "GET", "reason": err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.JSON(fiber.Map{"valid": false, "method": "GET", "reason": "Non-200 status code"})
	}

	// Step 3: Read a small chunk of body content
	buf := make([]byte, 512)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		return c.JSON(fiber.Map{"valid": false, "method": "GET", "reason": "Body read error"})
	}

	content := strings.TrimSpace(string(buf[:n]))
	if len(content) > 0 {
		return c.JSON(fiber.Map{"valid": true, "method": "GET"})
	}

	return c.JSON(fiber.Map{"valid": false, "method": "GET", "reason": "Empty body"})
}
