package main

import (
	"colly-scrapper-op/routes"
	"github.com/gofiber/fiber/v2"
	"log"
	"time"
)

func main() {
	app := fiber.New(fiber.Config{
		Prefork:      true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
	})

	app.Get("/website-validate", routes.WebsiteValidateHandler)
	app.Post("/quick-scrape", routes.QuickScrapeHandler)
	app.Post("/product-scrape", routes.ProductScrapeHandler)

	log.Println("🚀 Scraper service running on :8000")
	log.Fatal(app.Listen(":8000"))
}
