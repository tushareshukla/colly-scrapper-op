package main

import (
	"colly-scrapper-op/routes"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New(fiber.Config{
		Prefork:      false,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
	})

	app.Get("/website-validate", routes.WebsiteValidateHandler)
	app.Post("/quick-scrape", routes.QuickScrapeHandler)
	app.Post("/product-scrape", routes.ProductScrapeHandler)
	app.Post("/event-scrape", routes.EventScrapeHandler)
	log.Println("🚀 Scraper service running on :3000")
	log.Fatal(app.Listen(":3000"))
}
