package crawler

import (
	"context"
	"io"
	"net/http"
)

type Downloader struct {
	HTTPClient *http.Client
}

func (c *Downloader) Download(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
