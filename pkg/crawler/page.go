package crawler

import (
	"fmt"
	"io"
	"strconv"
	"text/template"

	"github.com/ueokande/hatenactl/pkg/hatena/blog"
)

type IndexPageValue struct {
	Title   string
	Entries []struct {
		Title string
		Link  string
	}
}

const IndexPageTemplate = `
<!DOCTYPE html>
<head>
  <meta charset="UTF-8"/>
  <title>{{ .Title }}</title>
</head>
<html>
<h1>{{ .Title }}</h1>
{{range .Entries}}
  <ul>
    <li><a href="{{ .Link }}">{{ .Title }}</a></li>
  </ul>
  <h1></h1>
{{end}}
</html>
`

type LandingValue struct {
	Title      string
	Categories []struct {
		Name string
		Path string
	}
	Archives []struct {
		Name string
		Path string
	}
}

const LandingPageTemplate = `
<!DOCTYPE html>
<head>
  <meta charset="UTF-8"/>
  <title>{{ .Title }}</title>
</head>
<html>
<h1>{{ .Title }}</h1>
<h2>Archives</h2>
  <ul>
{{range .Archives}}
    <li><a href="{{ .Path }}">{{ .Name }}</a></li>
{{end}}
  </ul>
<h2>By category</h2>
  <ul>
{{range .Categories}}
    <li><a href="{{ .Path }}">{{ .Name }}</a></li>
{{end}}
  </ul>
</html>
`

func (c Crawler) RenderCategoryIndex(w io.Writer, category string, entries []blog.Entry) error {
	var val IndexPageValue
	val.Title = "Category: " + category
	for _, e := range entries {
		val.Entries = append(val.Entries, struct {
			Title string
			Link  string
		}{
			Title: e.Title,
			Link:  c.Path.EntryURLPath(e),
		})
	}

	tmpl, err := template.New("category index").Parse(IndexPageTemplate)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, val)
}

func (c Crawler) RenderArchiveIndex(w io.Writer, year int, entries []blog.Entry) error {
	var val IndexPageValue
	val.Title = fmt.Sprintf("Entries from %d-01-01 to 1 year", year)
	for _, e := range entries {
		val.Entries = append(val.Entries, struct {
			Title string
			Link  string
		}{
			Title: e.Title,
			Link:  c.Path.EntryURLPath(e),
		})
	}

	tmpl, err := template.New("archive index").Parse(IndexPageTemplate)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, val)
}

func (c Crawler) RenderLanding(w io.Writer, title string, categories []string, years []int) error {
	var val LandingValue
	val.Title = title
	for _, year := range years {
		val.Archives = append(val.Archives, struct {
			Name string
			Path string
		}{
			Name: strconv.FormatInt(int64(year), 10),
			Path: c.Path.ArchiveUrlPath(year),
		})
	}
	for _, name := range categories {
		val.Categories = append(val.Categories, struct {
			Name string
			Path string
		}{
			Name: name,
			Path: c.Path.CategoryUrlPath(name),
		})
	}

	tmpl, err := template.New("landing page").Parse(LandingPageTemplate)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, val)
}
