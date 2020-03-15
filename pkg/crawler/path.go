package crawler

import (
	"net/url"
	"path"
	"path/filepath"
	"strconv"

	"github.com/ueokande/hatenactl/pkg/hatena/blog"
)

type Path struct {
	URLPrefix string
}

func (p Path) LandingURLPath() string {
	return path.Join("/", p.URLPrefix, "index.html")
}

func (p Path) LandingFilePath() string {
	return filepath.Join("index.html")
}

func (p Path) EntryURLPath(entry blog.Entry) string {
	return path.Join("/", p.URLPrefix, entry.Path(), "index.html")
}

func (p Path) EntryFilePath(entry blog.Entry) string {
	return filepath.Join(entry.Path(), "index.html")
}

func (p Path) ImageURLPath(entry blog.Entry, name string) string {
	return path.Join("/", p.URLPrefix, entry.Path(), name)
}

func (p Path) ImageFilePath(entry blog.Entry, name string) string {
	return filepath.Join(entry.Path(), name)
}

func (p Path) CategoryUrlPath(name string) string {
	// Escape twice to access escaped directory in the file system
	return path.Join("/", p.URLPrefix, "category", url.PathEscape(url.PathEscape(name)), "index.html")
}

func (p Path) CategoryFilePath(name string) string {
	return filepath.Join("category", url.PathEscape(name), "index.html")
}

func (p Path) ArchiveUrlPath(year int) string {
	return path.Join("/", p.URLPrefix, "archive", strconv.FormatInt(int64(year), 10), "index.html")
}

func (p Path) ArchiveFilePath(year int) string {
	return filepath.Join("archive", strconv.FormatInt(int64(year), 10), "index.html")
}
