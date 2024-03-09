package htmlutil

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type predicate = func(*html.Node) bool

func getElement(n *html.Node, p predicate) (*html.Node, error) {
	if p(n) {
		return n, nil
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		elem, err := getElement(c, p)
		if err == nil {
			return elem, nil
		}
	}
	return nil, errors.New("no matching node found")
}

func GetElement(blob io.Reader, p predicate) (*html.Node, error) {
	doc, err := html.Parse(blob)
	if err != nil {
		return nil, err
	}
	elem, err := getElement(doc, p)
	if err != nil {
		return nil, err
	}
	return elem, nil
}

func GetAttrValue(n *html.Node, attrName string) (string, error) {
	for _, attr := range n.Attr {
		if strings.EqualFold(attr.Key, attrName) {
			return attr.Val, nil
		}
	}
	return "", fmt.Errorf("attribute not found: %s", attrName)
}
