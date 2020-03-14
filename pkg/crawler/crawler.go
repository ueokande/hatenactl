package crawler

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ueokande/hatenactl/pkg/hatena/blog"
	"golang.org/x/net/html"
)

type Crawler struct {
	BlogClient *blog.Client
	DataStore  *DataStore

	HatenaID string
	BlogID   string
}

func (c Crawler) Start(ctx context.Context) error {
	var title string
	tr := &Transformer{
		Func: func(node *html.Node) (*html.Node, error) {
			if node.Type != html.ElementNode {
				return node, nil
			}
			if node.Data == "body" {
				h1 := &html.Node{Type: html.ElementNode, Data: "h1", FirstChild: &html.Node{
					Type: html.TextNode, Data: title,
				}}
				node.InsertBefore(h1, node.FirstChild)
			}
			if node.Data == "a" {
				for _, attr := range node.Attr {
					if attr.Key == "class" && attr.Val == "keyword" {
						if node.FirstChild == nil {
							return nil, nil
						}
						return &html.Node{
							Type: html.TextNode,
							Data: node.FirstChild.Data,
						}, nil
					}
				}
			}
			return node, nil
		},
	}
	return c.listAllEntries(ctx, func(ctx context.Context, e blog.Entry) error {
		title = e.Title
		if e.FormattedContent.Type != "text/html" {
			return errors.New("unknown content type: " + e.FormattedContent.Type)
		}

		p := filepath.Join(c.BlogID, e.Path(), "index.html")
		w, err := c.DataStore.Writer(p)
		if err != nil {
			return err
		}
		defer w.Close()

		root, err := html.Parse(strings.NewReader(e.FormattedContent.Content))
		if err != nil {
			return fmt.Errorf("unable to parse as html: %w", err)
		}

		err = tr.WalkTransform(root)
		if err != nil {
			return err
		}

		err = html.Render(w, root)
		if err != nil {
			return err
		}
		fmt.Println("saved", p)
		return nil
	})
}

func (c Crawler) listAllEntries(ctx context.Context, fn func(ctx context.Context, entry blog.Entry) error) error {
	input := blog.ListEntriesInput{
		HatenaID: c.HatenaID,
		BlogID:   c.BlogID,
	}
	for {
		feed, err := c.BlogClient.ListEntries(ctx, input)
		if err != nil {
			return fmt.Errorf("unable to list blog entries: %w", err)
		}

		for _, entry := range feed.Entries {
			err := fn(ctx, entry)
			if err != nil {
				return fmt.Errorf("unable process %s (%s): %w", entry.Path(), entry.ID, err)
			}
		}

		next := feed.NextPage()
		if len(next) == 0 {
			break
		}
		input.Page = next

		time.Sleep(1 * time.Second)
	}
	return nil
}
