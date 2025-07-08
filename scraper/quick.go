package scraper

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

var quickKeywords = []string{"about", "about-us", "info", "contact"}

func QuickScrape(startURL string) ScrapeResult {
	var mu sync.Mutex
	var pages []PageData
	visited := make(map[string]bool)

	domain := getDomain(startURL)
	log.Printf("[Colly] Allowed domain: %s", domain)

	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowedDomains(domain),
	)

	c.SetRequestTimeout(10 * time.Second)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 10,
		RandomDelay: 100 * time.Millisecond,
	})

	c.OnRequest(func(r *colly.Request) {
		log.Printf("➡️ Visiting: %s", r.URL.String())
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/114.0.0.0 Safari/537.36")
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")
		r.Headers.Set("Referer", startURL)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("❌ Request to %s failed with status %d: %v", r.Request.URL.String(), r.StatusCode, err)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("✅ Response from %s | Status: %d | Size: %d bytes", r.Request.URL.String(), r.StatusCode, len(r.Body))
		if len(r.Body) < 100 {
			log.Printf("⚠️ Response body too small from %s", r.Request.URL.String())
		}
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.DOM.Find("script, style, .gpu-banner, .hero-animation").Remove()
		raw := strings.TrimSpace(e.DOM.Text())
		text := cleanAndTrim(raw, 400, 5000)

		log.Printf("📄 Scraped from %s | Text length: %d", e.Request.URL.String(), len(text))

		if text != "" {
			mu.Lock()
			pages = append(pages, PageData{
				URL:  e.Request.URL.String(),
				Text: text,
			})
			mu.Unlock()
		} else {
			log.Printf("⚠️ No valid text found on %s", e.Request.URL.String())
		}
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" || visited[link] {
			return
		}
		if containsAny(link, quickKeywords) {
			log.Printf("🔗 Following link: %s", link)
			visited[link] = true
			_ = c.Visit(link)
		}
	})

	_ = c.Visit(startURL)
	c.Wait()

	if len(pages) == 0 {
		log.Printf("❌ No content found after visiting %d link(s)", len(visited)+1)
	}

	return ScrapeResult{Pages: pages}
}
