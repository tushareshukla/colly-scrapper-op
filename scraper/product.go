package scraper

import (
	"log"
	"sync"

	"github.com/gocolly/colly/v2"
)

var productKeywords = []string{
	"product", "products", "service", "services",
	"solution", "solutions", "offering", "offerings",
	"platform", "feature", "features", "app", "apps",
}

func ProductScrape(startURL string) ScrapeResult {
	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowedDomains(getDomain(startURL)),
	)

	c.SetRequestTimeout(3 * 1e9)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 20,
		RandomDelay: 100 * 1e6,
	})

	var mu sync.Mutex
	var pages []PageData
	var visited sync.Map // 🟢 concurrency-safe map

	c.OnHTML("body", func(e *colly.HTMLElement) {
		text := cleanAndTrim(stripHTML(e.DOM.Text()), 400, 4000)
		mu.Lock()
		pages = append(pages, PageData{
			URL:  e.Request.URL.String(),
			Text: text,
		})
		mu.Unlock()
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" {
			return
		}

		// atomically check and store in sync.Map
		if _, loaded := visited.LoadOrStore(link, true); loaded {
			return
		}

		if containsAny(link, productKeywords) {
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
