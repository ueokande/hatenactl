package crawler

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/ueokande/hatenactl/pkg/hatena/blog"
	"golang.org/x/net/html"
)

type Filter interface {
	Process(entry blog.Entry, root *html.Node) error
}

// TitleFilter presents a filter to add <title> into <head> and <h1> tag to the
// body from the entry..
type TitleFilter struct{}

func (f TitleFilter) Process(entry blog.Entry, root *html.Node) error {
	tr := &Transformer{
		Func: func(node *html.Node) (*html.Node, error) {
			if node.Type != html.ElementNode {
				return node, nil
			}
			if node.Data == "head" {
				title := &html.Node{Type: html.ElementNode, Data: "title", FirstChild: &html.Node{
					Type: html.TextNode, Data: entry.Title,
				}}
				node.InsertBefore(title, node.FirstChild)
			}
			if node.Data == "body" {
				h1 := &html.Node{Type: html.ElementNode, Data: "h1", FirstChild: &html.Node{
					Type: html.TextNode, Data: entry.Title,
				}}
				node.InsertBefore(h1, node.FirstChild)
			}
			return node, nil
		},
	}
	return tr.WalkTransform(root)
}

// HatenaKeywordFilter presents a filter to remove links of hatena keyword from
// HTML from the entry.
type HatenaKeywordFilter struct{}

func (f HatenaKeywordFilter) Process(entry blog.Entry, root *html.Node) error {
	tr := &Transformer{
		Func: func(node *html.Node) (*html.Node, error) {
			if node.Type != html.ElementNode {
				return node, nil
			}
			if node.Data == "a" {
				for _, attr := range node.Attr {
					if attr.Key == "class" && attr.Val == "keyword" {
						if node.FirstChild == nil {
							return nil, nil
						}
						return &html.Node{
							Type: html.TextNode,
							Data: node.FirstChild.Data,
						}, nil
					}
				}
			}
			return node, nil
		},
	}
	return tr.WalkTransform(root)
}

// CategoryFilter presents a filter to add categories into the <head> tag.
// The categories are provided by <meta> tags as the following:
//
//    <meta property="hatena:category" content="Games" />
//    <meta property="hatena:category" content="Hobby" />
type CategoryFilter struct{}

func (f CategoryFilter) Process(entry blog.Entry, root *html.Node) error {
	tr := &Transformer{
		Func: func(node *html.Node) (*html.Node, error) {
			if node.Type != html.ElementNode {
				return node, nil
			}
			if node.Data == "head" {
				for _, c := range entry.Categories {
					meta := makeMetaTag("category", c.Term)
					node.AppendChild(meta)
				}
			}
			return node, nil
		},
	}
	return tr.WalkTransform(root)
}

// ImagePathFilter presents a filter to fix image's url as a relative path as a
// base name.
//
// It converts a src attribute in the <img> tag:
//
//    <img src="https://my-cdn.example.com/2020/03/01/foobar.png" />
//
// to:
//
//    <img src="foobar.png" />
type ImagePathFilter struct{}

func (f ImagePathFilter) Process(entry blog.Entry, root *html.Node) error {
	tr := &Transformer{
		Func: func(node *html.Node) (*html.Node, error) {
			if node.Type != html.ElementNode {
				return node, nil
			}
			if node.Data == "img" {
				var src string
				for i, attr := range node.Attr {
					if attr.Key == "src" {
						src = attr.Val
						u, err := url.Parse(attr.Val)
						if err != nil {
							return nil, err
						}
						basename := path.Base(u.Path)
						if len(basename) > 127 {
							basename = fmt.Sprintf("%X", sha1.Sum([]byte(basename)))
						}
						node.Attr[i].Val = basename
					}
				}
				if len(src) > 0 {
					node.Attr = append(node.Attr, html.Attribute{
						Key: "data-original-url",
						Val: src,
					})
				}
			}
			return node, nil
		},
	}
	return tr.WalkTransform(root)
}

// CodeFilter presents a filter to make styled codes to plain text in the <pre>
// tags.
//
type CodeFilter struct{}

func (f CodeFilter) Process(entry blog.Entry, root *html.Node) error {
	tr := &Transformer{
		Func: func(node *html.Node) (*html.Node, error) {
			if node.Type != html.ElementNode {
				return node, nil
			}
			if node.Data == "span" && node.Parent.Data == "pre" {
				return &html.Node{
					Type: html.TextNode,
					Data: node.FirstChild.Data,
				}, nil
			} else if node.Data == "pre" {
				var attrs []html.Attribute
				for _, attr := range node.Attr {
					if attr.Key == "data-lang" {
						attrs = append(attrs, attr)
					}
				}
				node.Attr = attrs
			}
			return node, nil
		},
	}
	return tr.WalkTransform(root)
}

// DraftFilter presents a filter to add draft information as a meta tag.
type DraftFilter struct{}

func (f DraftFilter) Process(entry blog.Entry, root *html.Node) error {
	tr := &Transformer{
		Func: func(node *html.Node) (*html.Node, error) {
			if node.Type == html.ElementNode && node.Data == "head" {
				meta := makeMetaTag("draft", entry.Control.Draft)
				node.AppendChild(meta)
			}
			return node, nil
		},
	}
	return tr.WalkTransform(root)
}

// DateTimeFilter presents a filter to add timestamp on editted, published, and
// updated the entry as a meta tag.
type DateTimeFilter struct{}

func (f DateTimeFilter) Process(entry blog.Entry, root *html.Node) error {
	tr := &Transformer{
		Func: func(node *html.Node) (*html.Node, error) {
			if node.Type == html.ElementNode && node.Data == "head" {
				node.AppendChild(makeMetaTag("edited", entry.Edited.Format(time.RFC3339)))
				node.AppendChild(makeMetaTag("updated", entry.Updated.Format(time.RFC3339)))
				node.AppendChild(makeMetaTag("published", entry.Published.Format(time.RFC3339)))
			}
			return node, nil
		},
	}
	return tr.WalkTransform(root)
}

type LinkFilter struct{}

func (f LinkFilter) Process(entry blog.Entry, root *html.Node) error {
	tr := &Transformer{
		Func: func(node *html.Node) (*html.Node, error) {
			if node.Type == html.ElementNode && node.Data == "head" {
				node.AppendChild(makeMetaTag("alternate", entry.OriginalLink().Href))
			}
			return node, nil
		},
	}
	return tr.WalkTransform(root)
}

func makeMetaTag(property, content string) *html.Node {
	return &html.Node{
		Type: html.ElementNode,
		Data: "meta",
		Attr: []html.Attribute{
			{Key: "property", Val: "hatena:" + property},
			{Key: "content", Val: content},
		},
	}
}

// EncodingFilter presents a filter to provide a charset attribute (UTF-8) by
// the <meta> tag.
type EncodingFilter struct{}

func (f EncodingFilter) Process(entry blog.Entry, root *html.Node) error {
	tr := &Transformer{
		Func: func(node *html.Node) (*html.Node, error) {
			if node.Type == html.ElementNode && node.Data == "head" {
				node.AppendChild(&html.Node{
					Type: html.ElementNode,
					Data: "meta",
					Attr: []html.Attribute{
						{Key: "charset", Val: "UTF-8"},
					},
				})
			}
			return node, nil
		},
	}
	return tr.WalkTransform(root)
}

// AssetFilter presents as filter to add assets (css and javascript links) in
// the header.

type AssetFilter struct {
	CSSPaths        []string
	JavaScriptPaths []string
}

func (f AssetFilter) Process(entry blog.Entry, root *html.Node) error {
	tr := &Transformer{
		Func: func(node *html.Node) (*html.Node, error) {
			if node.Type == html.ElementNode && node.Data == "head" {
				// <link rel="stylesheet" type="text/css" href="theme.css">
				for _, p := range f.CSSPaths {
					node.AppendChild(&html.Node{
						Type: html.ElementNode,
						Data: "link",
						Attr: []html.Attribute{
							{Key: "rel", Val: "stylesheet"},
							{Key: "type", Val: "text/css"},
							{Key: "href", Val: p},
						},
					})
				}
				for _, p := range f.JavaScriptPaths {
					// <script src="myscripts.js"></script>
					node.AppendChild(&html.Node{
						Type: html.ElementNode,
						Data: "script",
						Attr: []html.Attribute{
							{Key: "src", Val: p},
						},
					})
				}
			}
			return node, nil
		},
	}
	return tr.WalkTransform(root)
}
