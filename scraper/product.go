package scraper

import (
	"log"
	"sync"
	"time"

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

	c.SetRequestTimeout(3 * time.Second)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 20,
		RandomDelay: 100 * time.Millisecond,
	})

	var mu sync.Mutex
	var pages []PageData
	var visited sync.Map // 🟢 concurrency-safe map

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/114.0.0.0 Safari/537.36")
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")
		r.Headers.Set("Referer", startURL)
		log.Printf("[Colly] Visiting: %s", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("[Colly] Response status: %d | Length: %d bytes", r.StatusCode, len(r.Body))
		log.Println("Sample content:\n", string(r.Body[:min(500, len(r.Body))]))
	})

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
			_ = c.Visit(link)
		}
	})
	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting: %s", r.URL.String())
	})

	_ = c.Visit(startURL)
	c.Wait()

	return ScrapeResult{Pages: pages}
}
