package crawler

import (
	"bytes"
	"strings"
	"testing"
	"time"

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

func TestImagePathFilter(t *testing.T) {
	src := `<html><head></head><body>` +
		`<img src="https://my-cdn.example.com/2020/03/01/foobar.png"/>` +
		`<x-img src="https://my-cdn.example.com/2020/03/01/foobar.png"></x-img>` +
		`</body></html>`
	result := `<html><head></head><body>` +
		`<img src="foobar.png"/>` +
		`<x-img src="https://my-cdn.example.com/2020/03/01/foobar.png"></x-img>` +
		`</body></html>`

	f := ImagePathFilter{}
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

func TestCodeFilter(t *testing.T) {
	src := `<html><head></head><body>` +
		`<pre class="code lang-sh" data-lang="sh" data-unlink=""><span class="synStatement">echo</span><span class="synConstant"> </span><span class="synStatement">&#39;</span><span class="synConstant">Hello World, Goodbye</span><span class="synStatement">&#39;</span>` +
		`</pre>` +
		`</body></html>`
	result := `<html><head></head><body>` +
		`<pre data-lang="sh">echo &#39;Hello World, Goodbye&#39;` +
		`</pre>` +
		`</body></html>`

	f := CodeFilter{}
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

func TestDraftFilter(t *testing.T) {
	src := `<html><head></head><body></body></html>`
	result := `<html><head>` +
		`<meta property="hatena:draft" content="yes"/>` +
		`</head><body></body></html>`

	f := DraftFilter{}
	root, err := html.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}

	err = f.Process(blog.Entry{
		Control: blog.Control{Draft: "yes"},
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

func TestDateTimeFilter(t *testing.T) {
	src := `<html><head></head><body></body></html>`
	result := `<html><head>` +
		`<meta property="hatena:edited" content="2020-02-14T11:22:33Z"/>` +
		`<meta property="hatena:updated" content="2020-03-14T11:22:33Z"/>` +
		`<meta property="hatena:published" content="2020-04-14T11:22:33Z"/>` +
		`</head><body></body></html>`

	f := DateTimeFilter{}
	root, err := html.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}

	err = f.Process(blog.Entry{
		Edited:    time.Date(2020, 02, 14, 11, 22, 33, 0, time.UTC),
		Updated:   time.Date(2020, 03, 14, 11, 22, 33, 0, time.UTC),
		Published: time.Date(2020, 04, 14, 11, 22, 33, 0, time.UTC),
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

func TestLinkFilter(t *testing.T) {
	src := `<html><head></head><body></body></html>`
	result := `<html><head>` +
		`<meta property="hatena:alternate" content="https://example.com/"/>` +
		`</head><body></body></html>`

	f := LinkFilter{}
	root, err := html.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}

	err = f.Process(blog.Entry{
		Links: []blog.Link{
			{Rel: "alternate", Href: "https://example.com/"},
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

func TestEncodingFilter(t *testing.T) {
	src := `<html><head></head><body></body></html>`
	result := `<html><head>` +
		`<meta charset="UTF-8"/>` +
		`</head><body></body></html>`

	f := EncodingFilter{}
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

func TestAssetFilter(t *testing.T) {
	src := `<html><head></head><body></body></html>`
	result := `<html><head>` +
		`<link rel="stylesheet" type="text/css" href="theme1.css"/>` +
		`<link rel="stylesheet" type="text/css" href="theme2.css"/>` +
		`<script src="script1.js"></script>` +
		`<script src="script2.js"></script>` +
		`</head><body></body></html>`

	f := AssetFilter{
		CSSPaths:        []string{"theme1.css", "theme2.css"},
		JavaScriptPaths: []string{"script1.js", "script2.js"},
	}
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
