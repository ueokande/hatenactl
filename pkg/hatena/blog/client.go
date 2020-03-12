package blog

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

// A Client is a client for HatenaBlog using Atom API.
//
// http://developer.hatena.ne.jp/ja/documents/blog/apis/atom
type Client struct {
	HTTPClient *http.Client
}

// ListEntriesInput represents an input parameter of the Client.ListEntries
type ListEntriesInput struct {
	// HatenaID (blog owner)
	HatenaID string
	// BlobID
	BlogID string

	// Page for a pagenation
	Page string
}

func (c *Client) ListEntries(ctx context.Context, input ListEntriesInput) (*Feed, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   "blog.hatena.ne.jp",
		Path:   path.Join(input.HatenaID, input.BlogID, "atom", "entry"),
	}
	if len(input.Page) > 0 {
		q := u.Query()
		q.Add("page", input.Page)
		u.RawQuery = q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("server returns %d (%s): %s", resp.StatusCode, resp.Status, respBody)
	}

	var feed Feed
	err = xml.NewDecoder(resp.Body).Decode(&feed)
	if err != nil {
		return nil, err
	}
	return &feed, nil
}
