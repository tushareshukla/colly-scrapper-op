package scraper

import (
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

type ScrapedPage struct {
	URL      string `json:"url"`
	Text     string `json:"text"`
	External bool   `json:"external"`
}

var eventKeywords = []string{
	"event", "events", "webinar", "conference", "summit", "expo", "forum", "seminar",
}

var platformDomainPatterns = []string{
	`hopin\.com`, `zoom\.us`, `webex\.com`, `airmeet\.com`, `vfairs\.com`, `eventbrite\.com`,
	`cvent\.com`, `bizzabo\.com`, `on24\.com`, `remo\.co`, `whova\.com`, `brella\.io`,
	`runtheworld\.today`, `splashthat\.com`, `accelevents\.com`, `bigmarker\.com`, `6connex\.com`,
	`gotowebinar\.com`, `gotomeeting\.com`, `slido\.com`, `inevent\.com`, `pheedloop\.com`,
	`swapcard\.com`, `eventzilla\.net`, `eventscase\.com`, `hubilo\.com`, `convene\.com`,
	`attendify\.com`, `socio\.events`, `eventcadence\.com`, `heysummit\.com`, `meetyoo\.com`,
	`gathertown\.com`, `shindig\.com`, `hexafair\.com`, `veertly\.com`, `eventsair\.com`,
	`sched\.com`, `glisser\.com`, `meetingplay\.com`, `vconferenceonline\.com`, `expopass\.com`,
	`bevy\.com`, `hubspot\.com`, `demio\.com`, `conferize\.com`, `tampevents\.com`, `tame\.events`,
	`spotme\.com`, `evvnt\.com`, `tickettailor\.com`, `ticketspice\.com`, `brighttalk\.com`,
}

func EventCrawler(startURL string) []ScrapedPage {
	startDomain := getDomain(startURL)
	c := colly.NewCollector(colly.Async(true))
	c.SetRequestTimeout(2 * 1e9)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 50, RandomDelay: 50 * 1e6})

	var mu sync.Mutex
	var result []ScrapedPage
	var visited sync.Map
	externalLinks := make(chan string, 100)

	matchEventKeyword := regexp.MustCompile(strings.Join(eventKeywords, `|`))
	platformRegexes := make([]*regexp.Regexp, 0, len(platformDomainPatterns))
	for _, p := range platformDomainPatterns {
		platformRegexes = append(platformRegexes, regexp.MustCompile(p))
	}

	c.OnHTML("body", func(e *colly.HTMLElement) {
		text := cleanAndTrim(stripHTML(e.DOM.Text()), 50, 10000)
		mu.Lock()
		result = append(result, ScrapedPage{
			URL:      e.Request.URL.String(),
			Text:     text,
			External: false,
		})
		mu.Unlock()

		e.DOM.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			if href == "" {
				return
			}
			host := getDomain(href)
			for _, re := range platformRegexes {
				if re.MatchString(host) {
					if _, loaded := visited.LoadOrStore(href, true); !loaded {
						externalLinks <- href
					}
				}
			}
		})
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" {
			return
		}
		if _, loaded := visited.LoadOrStore(link, true); loaded {
			return
		}
		if getDomain(link) == startDomain && matchEventKeyword.MatchString(link) {
			c.Visit(link)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting: %s", r.URL.String())
	})

	go func() {
		c.Visit(startURL)
		c.Wait()
		close(externalLinks)
	}()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ec := colly.NewCollector()
			ec.SetRequestTimeout(2 * 1e9)
			ec.OnHTML("body", func(e *colly.HTMLElement) {
				text := cleanAndTrim(stripHTML(e.DOM.Text()), 50, 10000)
				mu.Lock()
				result = append(result, ScrapedPage{
					URL:      e.Request.URL.String(),
					Text:     text,
					External: true,
				})
				mu.Unlock()
			})
			for link := range externalLinks {
				ec.Visit(link)
			}
		}()
	}

	wg.Wait()
	return result
}
