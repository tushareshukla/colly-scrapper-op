package scraper

import (
	"log"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

var productKeywords = []string{
	"product", "products", "service", "services",
	"solution", "solutions", "offering", "offerings",
	"platform", "feature", "features",
}

func ProductScrape(startURL string) ScrapeResult {
	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowedDomains(getDomain(startURL)),
	)

	c.SetRequestTimeout(3 * 1e9) // 3s
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 20,
		RandomDelay: 100 * 1e6, // 100ms
	})

	var mu sync.Mutex
	var pages []PageData
	visited := make(map[string]bool)

	c.OnHTML("body", func(e *colly.HTMLElement) {
		cleanText := cleanAndTrim(stripHTML(e.DOM.Text()), 400, 1000)
		mu.Lock()
		pages = append(pages, PageData{
			URL:  e.Request.URL.String(),
			Text: cleanText,
		})
		mu.Unlock()
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" || visited[link] {
			return
		}
		if containsAnyKeyword(link, productKeywords) {
			visited[link] = true
			c.Visit(link)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting: %s", r.URL.String())
	})

	c.Visit(startURL)
	c.Wait()

	return ScrapeResult{Pages: pages}
}

// Helper to check if link contains any keyword
func containsAnyKeyword(link string, keywords []string) bool {
	link = strings.ToLower(link)
	for _, k := range keywords {
		if strings.Contains(link, k) {
			return true
		}
	}
	return false
}
