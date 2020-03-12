package crawler

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ueokande/hatenactl/pkg/hatena/blog"
)

type Crawler struct {
	BlogClient *blog.Client
	DataStore  *DataStore

	HatenaID string
	BlogID   string
}

func (c Crawler) Start(ctx context.Context) error {
	return c.listAllEntries(ctx, func(ctx context.Context, e blog.Entry) error {
		if e.FormattedContent.Type != "text/html" {
			return errors.New("unknown content type: " + e.FormattedContent.Type)
		}

		p := filepath.Join(c.BlogID, e.Path(), "index.md")
		w, err := c.DataStore.Writer(p)
		if err != nil {
			return err
		}
		defer w.Close()

		fmt.Fprintf(w, "<h1>%s</h1>\n\n", e.Title)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(w, e.FormattedContent.Content)
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
