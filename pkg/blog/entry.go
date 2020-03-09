package blog

import (
	"encoding/xml"
	"strings"
	"time"
)

// A Feed represents a feed with entries from hatena
type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Xmlns   string   `xml:"xmlns,attr"`
	App     string   `xml:"app,attr"`

	ID       string `xml:"id"`
	Title    string `xml:"title"`
	Subtitle string `xml:"subtitle"`
	Links    []Link `xml:"link"`

	Author Author `xml:"author"`

	Entries []Entry `xml:"entry"`
}

// NextPage returns a next page id of the entry list.  It returns empty string
// if the next link is not presented.
func (f *Feed) NextPage() string {
	var first Link
	for _, l := range f.Links {
		if l.Rel == "first" {
			first = l
		}
	}
	if len(first.Href) == 0 {
		return ""
	}

	for _, l := range f.Links {
		if l.Rel == "next" {
			return strings.TrimPrefix(l.Href, first.Href+"?page=")
		}
	}
	return ""
}

// A Entry represents an entry of the blog.
type Entry struct {
	ID     string `xml:"id"`
	Author Author `xml:"author"`
	Title  string `xml:"title"`
	Links  []Link `xml:"link"`

	Updated   time.Time `xml:"updated"`
	Published time.Time `xml:"published"`
	Edited    time.Time `xml:"edited"`

	Summary          Content `xml:"summary"`
	Content          Content `xml:"content"`
	FormattedContent Content `xml:"formatted-content"`

	Control Control `xml:"control"`

	Categories []Category `xml:"category"`
}

// A Link represents a link
type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr"`
}

// An Author represents an author of the entry
type Author struct {
	Name string `xml:"name"`
}

// A Content represents a content of the entry.
type Content struct {
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// A Category represents a category of the entry
type Category struct {
	Term string `xml:"term,attr"`
}

// A Control represents a meta information of the entry (such as the entry is
// draft).
type Control struct {
	Draft string `xml:"draft"`
}
