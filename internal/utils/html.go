package utils

import (
	"io"
	"golang.org/x/net/html"
	"github.com/PuerkitoBio/goquery"
)

func RenderHTML(n *goquery.Selection, w io.Writer) error {
	for _, node := range n.Nodes {
		err := html.Render(w, node)
		if err != nil {
			return err
		}
	}
	return nil
}
