package crawler

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func testDepthFirstTransformerRemoveNodes(t *testing.T) {
	prefix := `<html><head></head><body>`
	suffix := `</body></html>`

	cases := []struct {
		src    string
		result string
		tr     Transformer
	}{
		{
			src:    `<p>hello, <a href="#">golang</a> world</p>`,
			result: `<p>hello, <a href="#">golang</a> world</p>`,
			tr:     Transformer{},
		},
		{
			src:    `<p>hello, <a href="#">golang</a> world</p>`,
			result: `<p>hello,  world</p>`,
			tr: Transformer{
				Func: func(node *html.Node) (*html.Node, error) {
					if node.Type == html.ElementNode && node.Data == "a" {
						return nil, nil
					}
					return node, nil
				},
			},
		},
		{
			src:    `<p>hello, <a href="#">golang</a> world</p>`,
			result: ``,
			tr: Transformer{
				Func: func(node *html.Node) (*html.Node, error) {
					if node.Type == html.ElementNode && node.Data == "p" {
						return nil, nil
					}
					return node, nil
				},
			},
		},
		{
			src:    `<p>hello, <a href="#">golang</a> world</p>`,
			result: `<p><a href="#"></a></p>`,
			tr: Transformer{
				Func: func(node *html.Node) (*html.Node, error) {
					if node.Type == html.TextNode {
						return nil, nil
					}
					return node, nil
				},
			},
		},
	}

	for _, c := range cases {
		root, err := html.Parse(strings.NewReader(c.src))
		if err != nil {
			t.Error(err)
			continue
		}

		err = c.tr.WalkTransform(root)
		if err != nil {
			t.Error(err)
			continue
		}

		var buf bytes.Buffer
		err = html.Render(&buf, root)
		if err != nil {
			t.Error(err)
			continue
		}
		actual := buf.String()
		actual = strings.TrimPrefix(actual, prefix)
		actual = strings.TrimSuffix(actual, suffix)

		if actual != c.result {
			t.Errorf("%q != %q", actual, c.result)
		}
	}
}

func testDepthFirstTransformerTransformNodes(t *testing.T) {
	prefix := `<html><head></head><body>`
	suffix := `</body></html>`

	cases := []struct {
		src    string
		result string
		tr     Transformer
	}{
		{
			src:    `<p>hello, <a href="#">golang</a> world</p>`,
			result: `<p>hello, <a href="#" class="link">golang</a> world</p>`,
			tr: Transformer{
				Func: func(node *html.Node) (*html.Node, error) {
					if node.Type == html.ElementNode && node.Data == "a" {
						node.Attr = append(node.Attr, html.Attribute{
							Key: "class", Val: "link",
						})
					}
					return node, nil
				},
			},
		},
		{
			src:    `<p>hello, <a href="#">golang</a> world</p>`,
			result: `<p>hello, --suppressed-- world</p>`,
			tr: Transformer{
				Func: func(node *html.Node) (*html.Node, error) {
					if node.Type == html.ElementNode && node.Data == "a" {
						return &html.Node{
							Type: html.TextNode,
							Data: "--suppressed--",
						}, nil
					}
					return node, nil
				},
			},
		},
		{
			src:    `<p>hello, <a href="#">golang</a> world</p>`,
			result: `<p>hello, --suppressed-- world</p>`,
			tr: Transformer{
				Func: func(node *html.Node) (*html.Node, error) {
					if node.Type == html.ElementNode && node.Data == "a" {
						return &html.Node{
							Type: html.TextNode,
							Data: "--suppressed--",
						}, nil
					}
					return node, nil
				},
			},
		},
	}

	for _, c := range cases {
		root, err := html.Parse(strings.NewReader(c.src))
		if err != nil {
			t.Error(err)
			continue
		}

		err = c.tr.WalkTransform(root)
		if err != nil {
			t.Error(err)
			continue
		}

		var buf bytes.Buffer
		err = html.Render(&buf, root)
		if err != nil {
			t.Error(err)
			continue
		}
		actual := buf.String()
		actual = strings.TrimPrefix(actual, prefix)
		actual = strings.TrimSuffix(actual, suffix)

		if actual != c.result {
			t.Errorf("%q != %q", actual, c.result)
		}
	}
}

func TestDepthFirstTransformer(t *testing.T) {
	t.Run("RemoveNodes", testDepthFirstTransformerRemoveNodes)
	t.Run("TransformNodes", testDepthFirstTransformerTransformNodes)
}
