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

	Filters []Filter
}

func (c Crawler) Start(ctx context.Context) error {
	return c.listAllEntries(ctx, func(ctx context.Context, e blog.Entry) error {
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

		for _, f := range c.Filters {
			err = f.Process(e, root)
			if err != nil {
				return fmt.Errorf("unable process a document: %w", err)
			}
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
