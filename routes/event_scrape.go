package routes

import (
	"colly-scrapper-op/scraper"
	"github.com/gofiber/fiber/v2"
)

func EventScrapeHandler(c *fiber.Ctx) error {
	var req ScrapeRequest
	if err := c.BodyParser(&req); err != nil || req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Provide valid JSON body with 'url'"})
	}
	result := scraper.EventScrape(req.URL)
	return c.JSON(result)
}
