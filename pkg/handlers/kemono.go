package handlers

import (
	"net/http"
	"time"

	"github.com/tooruu/goh/internal/htmlutil"

	"golang.org/x/net/html"
)

func KemonoHandler(resp *http.Response) (time.Time, error) {
	articleTime, err := htmlutil.GetElement(resp.Body, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "time"
	})
	defer resp.Body.Close()
	if err != nil {
		return time.Time{}, err
	}
	dt, _ := htmlutil.GetAttrValue(articleTime, "datetime")
	return time.Parse(time.DateTime, dt)
}
