// Package markup renders GitHub-flavored markdown to HTML with the same
// heading anchors and alert callouts GitHub produces.
package markup

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	reSpaces   = regexp.MustCompile(`\s+`)
	reNonWord  = regexp.MustCompile(`[^\p{L}\p{N}-]+`)
	reDashes   = regexp.MustCompile(`-{2,}`)
	reAlertTag = regexp.MustCompile(`^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]$`)
)

// Slug converts a heading into a GitHub-style anchor: lowercased Unicode
// letters, digits and hyphens only, spaces replaced with hyphens.
func Slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = reSpaces.ReplaceAllString(s, "-")
	s = reNonWord.ReplaceAllString(s, "")
	s = reDashes.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// Heading is one rendered heading collected from a document.
type Heading struct {
	Level int
	ID    string
	Text  string
}

// Transformer injects stable ids into headings, rewrites GitHub alert
// callouts into .markdown-alert blocks and collects headings for the TOC.
// It is shared across all documents of one build so that heading ids unique
// across the whole book.
type Transformer struct {
	used     map[string]bool
	Headings []Heading
}

// NewTransformer returns a Transformer pre-seeded with reserved ids.
func NewTransformer(reserved []string) *Transformer {
	t := &Transformer{used: map[string]bool{}}
	for _, id := range reserved {
		t.used[id] = true
	}
	return t
}

// Transform implements parser.ASTTransformer.
func (t *Transformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	src := reader.Source()
	t.Headings = nil
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Heading:
			text := nodeText(v, src)
			id := Slug(text)
			if t.used[id] {
				for i := 1; ; i++ {
					cand := id + "-" + itoa(i)
					if !t.used[cand] {
						id = cand
						break
					}
				}
			}
			t.used[id] = true
			v.SetAttributeString("id", []byte(id))
			t.Headings = append(t.Headings, Heading{Level: v.Level, ID: id, Text: text})
		case *ast.Blockquote:
			markAlert(v, src)
		}
		return ast.WalkContinue, nil
	})
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [12]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(b[pos:])
}

// markAlert converts `> [!TYPE]` callouts into GitHub-style alert blocks.
func markAlert(bq *ast.Blockquote, src []byte) {
	marker := bq.FirstChild()
	if marker == nil || marker.Kind() != ast.KindParagraph {
		return
	}
	text := strings.TrimSpace(nodeText(marker, src))
	m := reAlertTag.FindStringSubmatch(text)
	if m == nil {
		return
	}
	kind := strings.ToLower(m[1])
	bq.SetAttributeString("class", []byte("markdown-alert markdown-alert-"+kind))
	bq.RemoveChild(bq, marker)
}

func nodeText(n ast.Node, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := c.(type) {
		case *ast.Text:
			b.Write(v.Segment.Value(src))
		case *ast.CodeSpan:
			b.WriteString("code")
		case *ast.String:
			b.Write(v.Value)
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

// Renderer converts a single markdown body and returns its HTML plus the
// heading structure of that document.
type Renderer struct {
	md goldmark.Markdown
	tr *Transformer
}

// NewRenderer creates a GFM renderer with footnote support, GitHub-style
// alert callouts and global heading anchors. Reserved ids will never be
// reassigned to headings.
func NewRenderer(reserved []string) *Renderer {
	tr := NewTransformer(reserved)
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Footnote),
		goldmark.WithParserOptions(parser.WithASTTransformers(util.Prioritized(tr, 100))),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	return &Renderer{md: md, tr: tr}
}

// Render converts body markdown to HTML.
func (r *Renderer) Render(body string) (string, []Heading, error) {
	var b strings.Builder
	if err := r.md.Convert([]byte(body), &b); err != nil {
		return "", nil, err
	}
	hs := make([]Heading, len(r.tr.Headings))
	copy(hs, r.tr.Headings)
	return b.String(), hs, nil
}

// ExtractLinks returns the raw destinations of all links and images in body,
// as written in the markdown source. Reference definitions are resolved by
// goldmark, so `[text][ref]` yields the same destination as a direct link.
// Links written as raw HTML (<a href="...">) are not reported.
func ExtractLinks(body string) ([]string, error) {
	src := []byte(body)
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	doc := md.Parser().Parse(text.NewReader(src))
	var links []string
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Link:
			if len(v.Destination) > 0 {
				links = append(links, string(v.Destination))
			}
		case *ast.Image:
			if len(v.Destination) > 0 {
				links = append(links, string(v.Destination))
			}
		case *ast.AutoLink:
			if u := v.URL(src); len(u) > 0 {
				links = append(links, string(u))
			}
		}
		return ast.WalkContinue, nil
	})
	return links, nil
}
