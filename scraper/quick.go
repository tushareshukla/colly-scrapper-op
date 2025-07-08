package scraper

import (
	"log"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

var quickKeywords = []string{"about", "about-us", "info", "contact"}

func QuickScrape(startURL string) ScrapeResult {
	var mu sync.Mutex
	var pages []PageData
	visited := make(map[string]bool)

	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowedDomains(getDomain(startURL)),
	)

	c.SetRequestTimeout(3 * time.Second)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 10,
		RandomDelay: 100 * time.Millisecond,
	})

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
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.DOM.Find("script, style, .gpu-banner, .hero-animation").Remove()
		text := cleanAndTrim(e.DOM.Text(), 400, 1000)
		if text != "" {
			mu.Lock()
			pages = append(pages, PageData{
				URL:  e.Request.URL.String(),
				Text: text,
			})
			mu.Unlock()
		}
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" || visited[link] {
			return
		}
		if containsAny(link, quickKeywords) {
			visited[link] = true
			_ = c.Visit(link)
		}
	})

	_ = c.Visit(startURL)
	c.Wait()

	if len(pages) > 0 {
		log.Println("✅ Colly succeeded")
		return ScrapeResult{Pages: pages}
	}

	log.Println("❌ Colly returned no content.")
	return ScrapeResult{Pages: nil}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
