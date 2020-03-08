package hatena

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

// A Client is a client to get entries from Hatena blob via AtomPub
//
// http://developer.hatena.ne.jp/ja/documents/blog/apis/atom
type Client struct {
	http *http.Client
}

// NewClient creates a new client with the user and the API token for
// the authentication.
func NewClient(user, token string, client *http.Client) *Client {
	if client == nil {
		return &Client{
			http: &http.Client{
				Transport: WSSETransport{
					Username: user,
					Password: token,
				},
			},
		}
	}
	return &Client{
		http: http.DefaultClient,
	}
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

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server returns %d (%s)", resp.StatusCode, resp.Status)
	}

	var feed Feed
	err = xml.NewDecoder(resp.Body).Decode(&feed)
	if err != nil {
		return nil, err
	}
	return &feed, nil
}
