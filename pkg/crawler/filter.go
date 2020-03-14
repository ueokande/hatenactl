package crawler

import (
	"net/url"
	"path"

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
					meta := &html.Node{
						Type: html.ElementNode,
						Data: "meta",
						Attr: []html.Attribute{
							{Key: "property", Val: "hatena:category"},
							{Key: "content", Val: c.Term},
						},
					}
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
				for i, attr := range node.Attr {
					if attr.Key == "src" {
						u, err := url.Parse(attr.Val)
						if err != nil {
							return nil, err
						}
						node.Attr[i].Val = path.Base(u.Path)
					}
				}
			}
			return node, nil
		},
	}
	return tr.WalkTransform(root)
}
