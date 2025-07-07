package scraper

import (
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

type ScrapeResult struct {
	Pages []PageData `json:"pages"`
}

type PageData struct {
	URL  string `json:"url"`
	Text string `json:"text"`
}

var matchKeywords = []string{
	"about", "about-us", "info", "contact",
}

func QuickScrape(startURL string) ScrapeResult {
	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowedDomains(getDomain(startURL)),
	)

	c.SetRequestTimeout(10 * 1000000000) // 10 sec
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 5,
	})

	var mu sync.Mutex
	var pages []PageData
	visited := make(map[string]bool)

	c.OnHTML("body", func(e *colly.HTMLElement) {
		cleanText := cleanAndTrim(stripHTML(e.DOM.Text()), 400, 2000)
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
		if containsKeyword(link) {
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

func containsKeyword(link string) bool {
	l := strings.ToLower(link)
	for _, k := range matchKeywords {
		if strings.Contains(l, k) {
			return true
		}
	}
	return false
}

func getDomain(link string) string {
	re := regexp.MustCompile(`https?://([^/]+)/?`)
	m := re.FindStringSubmatch(link)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

func stripHTML(input string) string {
	reTags := regexp.MustCompile(`(?is)<[^>]+>`)
	return strings.Join(strings.Fields(reTags.ReplaceAllString(input, "")), " ")
}

func cleanAndTrim(text string, min, max int) string {
	text = strings.TrimSpace(text)
	if len(text) > max {
		return text[:max]
	}
	return text
}
