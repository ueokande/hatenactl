package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ueokande/hatenactl/pkg/hatena/blog"
	"golang.org/x/net/html"
)

var downloader Downloader

type Crawler struct {
	BlogClient *blog.Client
	DataStore  *DataStore

	HatenaID string
	BlogID   string

	Filters []Filter
}

func (c Crawler) Start(ctx context.Context) error {
	urlext := ImageURLExtractor{}

	return c.listAllEntries(ctx, func(ctx context.Context, entry blog.Entry) error {
		if entry.FormattedContent.Type != "text/html" {
			return errors.New("unknown content type: " + entry.FormattedContent.Type)
		}

		p := filepath.Join(c.BlogID, entry.Path(), "index.html")
		w, err := c.DataStore.Writer(p)
		if err != nil {
			return err
		}
		defer w.Close()

		root, err := html.Parse(strings.NewReader(entry.FormattedContent.Content))
		if err != nil {
			return fmt.Errorf("unable to parse as html: %w", err)
		}

		urls := urlext.ExtractImageURLs(root)
		err = c.downloadImages(ctx, entry, urls)
		if err != nil {
			return err
		}

		for _, f := range c.Filters {
			err = f.Process(entry, root)
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

func (c Crawler) downloadImages(ctx context.Context, entry blog.Entry, urls []string) error {
	download := func(ctx context.Context, src string) error {
		url, err := url.Parse(src)
		if err != nil {
			return err
		}
		basename := path.Base(url.Path)

		resp, err := downloader.Download(ctx, src)
		if err != nil {
			return err
		}
		defer resp.Close()

		f, err := c.DataStore.Writer(filepath.Join(c.BlogID, entry.Path(), basename))
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, resp)
		if err != nil {
			return err
		}
		return nil
	}
	for _, u := range urls {
		err := download(ctx, u)
		if err != nil {
			return fmt.Errorf("unable download %s: %w", u, err)
		}
	}
	return nil
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

type ImageURLExtractor struct{}

func (e *ImageURLExtractor) ExtractImageURLs(root *html.Node) []string {
	var urls []string
	w := Walker{
		Func: func(node *html.Node) error {
			if node.Type == html.ElementNode && node.Data == "img" {
				for _, attr := range node.Attr {
					if attr.Key == "src" {
						urls = append(urls, attr.Val)
					}
				}
			}
			return nil
		},
	}
	w.Walk(root)

	return urls
}
