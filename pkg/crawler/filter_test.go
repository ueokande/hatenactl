package crawler

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ueokande/hatenactl/pkg/hatena/blog"
	"golang.org/x/net/html"
)

func TestTitleFilter(t *testing.T) {
	src := `<html><head></head><body>Hello, world</body></html>
`
	result := `<html><head><title>Greeting</title></head><body><h1>Greeting</h1>Hello, world
</body></html>`

	f := TitleFilter{}
	root, err := html.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}

	err = f.Process(blog.Entry{Title: "Greeting"}, root)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = html.Render(&buf, root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := buf.String()

	if rendered != result {
		t.Errorf("%q != %q", rendered, result)
	}
}

func TestHatenaKeywordFilter(t *testing.T) {
	src := `<html><head></head><body>
<h1>Greeting in <a class="keyword">title</a></h1>
Hello, <a class="keyword">golang</a> world (<a href="#">see also</a>)</body></html>
`
	result := `<html><head></head><body>
<h1>Greeting in title</h1>
Hello, golang world (<a href="#">see also</a>)
</body></html>`

	f := HatenaKeywordFilter{}
	root, err := html.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}

	err = f.Process(blog.Entry{}, root)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = html.Render(&buf, root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := buf.String()

	if rendered != result {
		t.Errorf("%q != %q", rendered, result)
	}
}

func TestCategoryFilter(t *testing.T) {
	src := `<html><head></head><body></body></html>`
	result := `<html><head>` +
		`<meta property="hatena:category" content="Games"/>` +
		`<meta property="hatena:category" content="Hobby"/>` +
		`</head><body></body></html>`

	f := CategoryFilter{}
	root, err := html.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}

	err = f.Process(blog.Entry{
		Categories: []blog.Category{
			{Term: "Games"}, {Term: "Hobby"},
		},
	}, root)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = html.Render(&buf, root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := buf.String()

	if rendered != result {
		t.Errorf("%q != %q", rendered, result)
	}
}
