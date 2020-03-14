package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/ueokande/hatenactl/pkg/crawler"
	"github.com/ueokande/hatenactl/pkg/hatena/blog"
	"github.com/ueokande/hatenactl/pkg/hatena/oauth1"
)

var (
	OAuthConsumerKey    = os.Getenv("OAUTH_CONSUMER_KEY")
	OAuthConsumerSecret = os.Getenv("OAUTH_CONSUMER_SECRET")
	OAuthToken          = os.Getenv("OAUTH_TOKEN")
	OAuthTokenSecret    = os.Getenv("OAUTH_TOKEN_SECRET")

	flgHatenaID = flag.String("hatena-id", "", "hatena account id")
	flgBlogID   = flag.String("blog-id", "", "hatena blog id")
)

func validate() error {
	if len(OAuthConsumerKey) == 0 {
		return errors.New("OAUTH_CONSUMER_KEY not set")
	}
	if len(OAuthConsumerSecret) == 0 {
		return errors.New("OAUTH_CONSUMER_SECRET not set")
	}
	if len(OAuthToken) == 0 {
		return errors.New("OAUTH_TOKEN not set")
	}
	if len(OAuthTokenSecret) == 0 {
		return errors.New("OAUTH_TOKEN_SECRET not set")
	}
	if len(*flgHatenaID) == 0 {
		return errors.New("--hatena-id not set")
	}
	if len(*flgBlogID) == 0 {
		return errors.New("--blog-id not set")
	}
	return nil
}

func run(ctx context.Context) error {
	err := validate()
	if err != nil {
		return err
	}

	c := &crawler.Crawler{
		HatenaID: *flgHatenaID,
		BlogID:   *flgBlogID,
		BlogClient: &blog.Client{
			HTTPClient: oauth1.NewHTTPClient(
				OAuthConsumerKey, OAuthConsumerSecret,
				OAuthToken, OAuthTokenSecret,
			),
		},
		DataStore: &crawler.DataStore{
			Directory: "/tmp/",
		},
		Filters: []crawler.Filter{
			&crawler.TitleFilter{},
			&crawler.HatenaKeywordFilter{},
		},
	}
	return c.Start(ctx)
}

func main() {
	flag.Parse()

	err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
