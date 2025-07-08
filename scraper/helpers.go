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

func cleanAndTrim(raw string, minLen, maxLen int) string {
	// Remove JS and CSS
	reScript := regexp.MustCompile(`(?s)<script.*?>.*?</script>`)
	reStyle := regexp.MustCompile(`(?s)<style.*?>.*?</style>`)
	reJSLike := regexp.MustCompile(`(?i)(var|let|function)\s+\w+\s*=?\s*[\{\(]`)
	reCSSLike := regexp.MustCompile(`(?i)\.\w+\s*\{`)

	cleaned := reScript.ReplaceAllString(raw, "")
	cleaned = reStyle.ReplaceAllString(cleaned, "")
	cleaned = reJSLike.ReplaceAllString(cleaned, "")
	cleaned = reCSSLike.ReplaceAllString(cleaned, "")

	// Normalize white spaces
	cleaned = strings.TrimSpace(cleaned)
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")

	// Bound length
	if len(cleaned) < minLen {
		return ""
	}
	if len(cleaned) > maxLen {
		cleaned = cleaned[:maxLen]
	}
	return cleaned
}

func getDomain(link string) string {
	re := regexp.MustCompile(`https?://([^/]+)/?`)
	m := re.FindStringSubmatch(link)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}
