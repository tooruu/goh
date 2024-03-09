package handlers

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/tooruu/goh/internal/htmlutil"

	"golang.org/x/net/html"
)

var dateRegex = regexp.MustCompile(`[0-9]{4}-[0-9]{2}-[0-9]{2}`)

func F95zoneHandler(resp *http.Response) (time.Time, error) {
	postTitle, err := htmlutil.GetElement(resp.Body, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "title"
	})
	defer resp.Body.Close()
	if err != nil {
		return time.Time{}, err
	}
	lastUpdate := dateRegex.FindString(postTitle.FirstChild.Data)
	if lastUpdate == "" {
		return time.Time{}, errors.New("date not found in title")
	}
	return time.Parse(time.DateOnly, lastUpdate)
}
