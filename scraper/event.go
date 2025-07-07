package scraper

import (
	"log"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

var eventKeywords = []string{
	"event", "webinar", "conference", "summit", "expo", "forum", "seminar",
}

type EventPage struct {
	URL      string   `json:"url"`
	Title    string   `json:"title"`
	Date     string   `json:"date"`
	Speakers []string `json:"speakers"`
	Host     string   `json:"host"`
}

type EventResult struct {
	Pages []EventPage `json:"pages"`
}

func EventScrape(startURL string) EventResult {
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
	var pages []EventPage
	visited := make(map[string]bool)

	c.OnHTML("body", func(e *colly.HTMLElement) {
		title := e.DOM.Find("h1").First().Text()
		if title == "" {
			title = e.DOM.Find("title").Text()
		}

		date := ""
		e.DOM.Find("time").EachWithBreak(func(_ int, s *goquery.Selection) bool {
			if t, exists := s.Attr("datetime"); exists {
				date = t
				return false
			}
			date = s.Text()
			return false
		})
		if date == "" {
			e.DOM.Find(`meta[property="article:published_time"]`).EachWithBreak(func(_ int, s *goquery.Selection) bool {
				if t, exists := s.Attr("content"); exists {
					date = t
					return false
				}
				return true
			})
		}

		var speakers []string
		e.DOM.Find(".speaker, .speakers, .presenter, .author").Each(func(_ int, s *goquery.Selection) {
			txt := strings.TrimSpace(s.Text())
			if txt != "" {
				speakers = append(speakers, txt)
			}
		})

		host := getDomain(startURL)

		mu.Lock()
		pages = append(pages, EventPage{
			URL:      e.Request.URL.String(),
			Title:    cleanAndTrim(title, 10, 200),
			Date:     date,
			Speakers: speakers,
			Host:     host,
		})
		mu.Unlock()
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" || visited[link] {
			return
		}
		if containsAny(link, eventKeywords) {
			visited[link] = true
			c.Visit(link)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting: %s", r.URL.String())
	})

	c.Visit(startURL)
	c.Wait()

	return EventResult{Pages: pages}
}
