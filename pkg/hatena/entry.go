package hatena

import (
	"encoding/xml"
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
