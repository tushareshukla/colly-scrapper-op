package scraper

import (
	"log"
	"sync"

	"github.com/gocolly/colly/v2"
)

var quickKeywords = []string{"about", "about-us", "info", "company", "contact"}

func QuickScrape(startURL string) ScrapeResult {
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
	visited := make(map[string]bool)

	c.OnHTML("body", func(e *colly.HTMLElement) {
		text := cleanAndTrim(stripHTML(e.DOM.Text()), 400, 1000)
		mu.Lock()
		pages = append(pages, PageData{URL: e.Request.URL.String(), Text: text})
		mu.Unlock()
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" || visited[link] {
			return
		}
		if containsAny(link, quickKeywords) {
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
