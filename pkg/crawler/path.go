package crawler

import (
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/ueokande/hatenactl/pkg/hatena/blog"
)

func LandingPath() string {
	return "index.html"
}

func EntryPath(entry blog.Entry) string {
	return filepath.Join(entry.Path(), "index.html")
}

func ImagePath(entry blog.Entry, image string) string {
	return filepath.Join(entry.Path(), image)
}

func CategoryPath(category string) string {
	return filepath.Join("category", url.PathEscape(category), "index.html")
}

func ArchivePath(year int) string {
	return filepath.Join("archive", strconv.FormatInt(int64(year), 10), "index.html")
}
