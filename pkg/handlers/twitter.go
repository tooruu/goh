package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tooruu/goh/internal/htmlutil"

	"golang.org/x/net/html"
)

func snowflakeTime(s int64) time.Time {
	const twitterEpoch int64 = 1288834974657
	millis := (s >> 22) + twitterEpoch
	return time.UnixMilli(millis).UTC()
}

func TwitterHandler(resp *http.Response) (time.Time, error) {
	mediaLink, err := htmlutil.GetElement(resp.Body, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "a" {
			link, _ := htmlutil.GetAttrValue(n, "href")
			return strings.Contains(link, "/status/")
		}
		return false
	})
	defer resp.Body.Close()
	if err != nil {
		return time.Time{}, err
	}
	path, _ := htmlutil.GetAttrValue(mediaLink, "href")
	postIdStr := strings.Split(path, "/")[5]
	postId, _ := strconv.ParseInt(postIdStr, 10, 64)
	return snowflakeTime(postId), nil
}
