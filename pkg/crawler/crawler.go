package crawler

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/ueokande/hatenactl/pkg/hatena/blog"
	"golang.org/x/net/html"
)

var downloader Downloader

type Crawler struct {
	BlogClient *blog.Client
	DataStore  *DataStore
	Path       *Path
	CSSPath    string

	HatenaID string
	BlogID   string

	Filters []Filter
}

func (c Crawler) Start(ctx context.Context) error {
	byCategory := make(map[string][]blog.Entry)
	byYear := make(map[int][]blog.Entry)
	urlext := ImageURLExtractor{}

	// 1. Download entries and images contained in the entry
	err := c.listAllEntries(ctx, func(ctx context.Context, entry blog.Entry) error {
		if entry.FormattedContent.Type != "text/html" {
			return errors.New("unknown content type: " + entry.FormattedContent.Type)
		}
		for _, cat := range entry.Categories {
			byCategory[cat.Term] = append(byCategory[cat.Term], entry)
		}
		byYear[entry.Published.Year()] = append(byYear[entry.Published.Year()], entry)

		p := c.Path.EntryFilePath(entry)
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
	if err != nil {
		return err
	}

	// 2. Generate index page by a category
	for cat, entries := range byCategory {
		err := func(category string, entries []blog.Entry) error {
			p := c.Path.CategoryFilePath(category)
			f, err := c.DataStore.Writer(p)
			if err != nil {
				return err
			}
			defer f.Close()

			err = c.RenderCategoryIndex(f, category, entries)
			if err != nil {
				return err
			}
			fmt.Println("saved", p)

			return nil
		}(cat, entries)
		if err != nil {
			return fmt.Errorf("unable to create a category index '%s': %w", cat, err)
		}
	}

	var years []int
	for year := range byYear {
		years = append(years, year)
	}
	sort.Ints(years)

	// 3. Generate archive page grouped by a year
	for _, year := range years {
		err := func(year int, entries []blog.Entry) error {
			p := c.Path.ArchiveFilePath(year)
			f, err := c.DataStore.Writer(p)
			if err != nil {
				return err
			}
			defer f.Close()

			err = c.RenderArchiveIndex(f, year, entries)
			if err != nil {
				return err
			}
			fmt.Println("saved", p)

			return nil
		}(year, byYear[year])
		if err != nil {
			return fmt.Errorf("unable to create an archive index '%d-01-01': %w", year, err)
		}
	}

	// 4. Generate landing page
	err = func() error {
		p := c.Path.LandingFilePath()
		f, err := c.DataStore.Writer(p)
		if err != nil {
			return err
		}
		defer f.Close()

		var categories []string
		for category := range byCategory {
			categories = append(categories, category)
		}
		sort.Strings(categories)
		err = c.RenderLanding(f, c.BlogID, categories, years)
		if err != nil {
			return err
		}
		fmt.Println("saved", p)
		return nil
	}()
	return err
}

func (c Crawler) downloadImages(ctx context.Context, entry blog.Entry, urls []string) error {
	download := func(ctx context.Context, src string) error {
		url, err := url.Parse(src)
		if err != nil {
			return err
		}
		basename := path.Base(url.Path)

		// convert long file name
		if len(basename) > 127 {
			basename = fmt.Sprintf("%X", sha1.Sum([]byte(basename)))
		}

		resp, err := downloader.Download(ctx, src)
		if err != nil {
			return err
		}
		defer resp.Close()

		p := c.Path.ImageFilePath(entry, basename)
		f, err := c.DataStore.Writer(p)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, resp)
		if err != nil {
			return err
		}
		fmt.Println("saved", p)
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
