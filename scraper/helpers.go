package scraper

import (
	"regexp"
	"strings"
)

type ScrapeResult struct {
	Pages []PageData `json:"pages"`
}

type PageData struct {
	URL  string `json:"url"`
	Text string `json:"text"`
}

func containsAny(link string, keywords []string) bool {
	link = strings.ToLower(link)
	for _, k := range keywords {
		if strings.Contains(link, k) {
			return true
		}
	}
	return false
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

func getDomain(link string) string {
	re := regexp.MustCompile(`https?://([^/]+)/?`)
	m := re.FindStringSubmatch(link)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}
