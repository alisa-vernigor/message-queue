package grabber

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var excludedPatterns = []string{
	"File:",
	"Talk:",
	"Category:",
	"Special:",
	"Wikipedia:",
	"Help:",
	"Portal:",
}

func Grab(webPage string) ([]string, error) {
	parsedURL, err := url.Parse(webPage)
	if err != nil {
		return nil, err
	}
	prefix := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Hostname())

	resp, err := http.Get(webPage)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	f := func(i int, s *goquery.Selection) bool {
		link, _ := s.Attr("href")

		if !strings.HasPrefix(link, "/wiki/") {
			return false
		}

		link = strings.TrimPrefix(link, "/wiki/")

		for _, pattern := range excludedPatterns {
			if strings.HasPrefix(link, pattern) {
				return false
			}
		}

		return true
	}

	var res []string
	doc.Find("body a").FilterFunction(f).Each(func(_ int, tag *goquery.Selection) {
		link, exists := tag.Attr("href")
		if exists {
			link := fmt.Sprintf("%s%s", prefix, link)
			res = append(res, link)
		}
	})

	return res, nil
}
