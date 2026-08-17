// Package book assembles scanned markdown documents into a single HTML book.
package book

import (
	"bytes"
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	xhtml "golang.org/x/net/html"

	"github.com/gothchibjo/md-book-builder/internal/assets"
	"github.com/gothchibjo/md-book-builder/internal/config"
	"github.com/gothchibjo/md-book-builder/internal/markup"
	"github.com/gothchibjo/md-book-builder/internal/source"
)

// Stats reports book-level metrics used by the verify command.
type Stats struct {
	Documents     int
	FrontTables   int
	TOCEntries    int
	InternalLinks int
	Flattened     int
	BrokenAnchors int
}

// Book is a fully assembled HTML book.
type Book struct {
	Title    string
	Subtitle string
	HTML     string
	Stats    Stats
}

// Builder builds a Book from a Config.
type Builder struct {
	cfg    *config.Config
	kbRoot string
	docs   []source.Doc
	relIDs map[string]string // relative path -> section id
	ids    []string          // all document section ids
	stats  Stats
}

// New resolves and scans sources for cfg.
func New(cfg *config.Config) (*Builder, error) {
	kbRoot, err := filepath.Abs(cfg.KBRoot)
	if err != nil {
		return nil, err
	}
	docs, err := source.Scan(kbRoot, cfg.Include)
	if err != nil {
		return nil, err
	}
	docs, err = expandByLinks(kbRoot, docs, cfg.Include, cfg.Transitive())
	if err != nil {
		return nil, err
	}
	docs, err = source.Exclude(kbRoot, docs, cfg.Exclude)
	if err != nil {
		return nil, err
	}
	b := &Builder{
		cfg:    cfg,
		kbRoot: kbRoot,
		docs:   docs,
		relIDs: map[string]string{},
	}
	used := map[string]bool{}
	for i := range docs {
		d := &docs[i]
		id := sectionID(d)
		for used[id] {
			id += "-1"
		}
		used[id] = true
		b.ids = append(b.ids, id)
		b.relIDs[d.RelPath] = id
	}
	return b, nil
}

func sectionID(d *source.Doc) string {
	if d.Code != "" {
		return markup.Slug(d.Code)
	}
	base := strings.TrimSuffix(d.RelPath, filepath.Ext(d.RelPath))
	return markup.Slug(strings.ReplaceAll(base, "/", " "))
}

// expandByLinks appends documents that are referenced by internal links and
// resolve into one of the declared link-root directories (plain directory
// entries in include). The link-root directories themselves are never
// browsed. When transitive is true, documents pulled in by links are followed
// further until a fixpoint.
func expandByLinks(kbRoot string, docs []source.Doc, include []string, transitive bool) ([]source.Doc, error) {
	_, linkRoots := source.ClassifyPatterns(kbRoot, include)
	if len(linkRoots) == 0 {
		return docs, nil
	}
	set := map[string]bool{}
	for _, d := range docs {
		set[d.RelPath] = true
	}
	queue := append([]source.Doc(nil), docs...)
	pulled := map[string]source.Doc{}
	for len(queue) > 0 {
		d := queue[0]
		queue = queue[1:]
		links, err := markup.ExtractLinks(d.Body)
		if err != nil {
			return nil, fmt.Errorf("collect links from %s: %w", d.RelPath, err)
		}
		for _, href := range links {
			if isExternal(href) {
				continue
			}
			p, _ := splitFragment(href)
			if p == "" {
				continue
			}
			rel, ok := resolveRel(kbRoot, d.AbsPath, p)
			if !ok {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(rel), ".md") {
				continue
			}
			if set[rel] || !underAnyLinkRoot(rel, linkRoots) {
				continue
			}
			nd, err := source.Load(kbRoot, rel)
			if err != nil {
				return nil, fmt.Errorf("load %s: %w", rel, err)
			}
			set[rel] = true
			pulled[rel] = nd
			if transitive {
				queue = append(queue, nd)
			}
		}
	}
	if len(pulled) == 0 {
		return docs, nil
	}
	rels := make([]string, 0, len(pulled))
	for rel := range pulled {
		rels = append(rels, rel)
	}
	source.SortRels(rels)
	for _, rel := range rels {
		docs = append(docs, pulled[rel])
	}
	return docs, nil
}

// underAnyLinkRoot reports whether rel is inside one of the roots.
func underAnyLinkRoot(rel string, roots []string) bool {
	for _, r := range roots {
		if rel == r || strings.HasPrefix(rel, r+"/") {
			return true
		}
	}
	return false
}

// resolveRel resolves a relative reference ref (path, fragment or both)
// against the directory of dAbs and returns the cleaned relative path under
// kbRoot.
func resolveRel(kbRoot, dAbs, ref string) (string, bool) {
	target := filepath.Clean(filepath.Join(filepath.Dir(dAbs), filepath.FromSlash(ref)))
	rel, err := filepath.Rel(kbRoot, target)
	if err != nil {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

// Build renders every document, fixes links and assembles the final HTML.
func (b *Builder) Build() (*Book, error) {
	if len(b.docs) == 0 {
		return nil, fmt.Errorf("no documents matched the include patterns")
	}

	renderer := markup.NewRenderer(b.ids)
	tocLevels := map[int]bool{}
	for _, l := range b.cfg.TOCLevels {
		tocLevels[l] = true
	}

	var tocItems []string
	var body strings.Builder

	for _, d := range b.docs {
		id := b.relIDs[d.RelPath]
		htmlBody, headings, err := renderer.Render(d.Body)
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", d.RelPath, err)
		}
		fixed, err := b.fixLinks(htmlBody, d)
		if err != nil {
			return nil, fmt.Errorf("fix links %s: %w", d.RelPath, err)
		}

		body.WriteString("<section class=\"book-doc\" id=\"")
		body.WriteString(id)
		body.WriteString("\">")
		if len(d.Front) > 0 {
			b.stats.FrontTables++
			body.WriteString(frontmatterTable(d.Front))
		}
		body.WriteString(fixed)
		body.WriteString("</section>\n")

		label := docLabel(&d)
		tocItems = append(tocItems, b.tocItem(id, label, headings, tocLevels))
	}

	b.stats.Documents = len(b.docs)
	if b.cfg.ShowTOC() {
		b.stats.TOCEntries = len(b.docs)
	}

	full := `<main class="markdown-body">` + body.String() + `</main>`
	if b.cfg.ShowTOC() {
		full = `<nav class="book-toc"><h1>Содержание</h1><ul>` + strings.Join(tocItems, "\n") + `</ul></nav>` + full
	}
	if b.cfg.ShowCover() {
		subtitle := ""
		if b.cfg.Subtitle != "" {
			subtitle = "<p>" + html.EscapeString(b.cfg.Subtitle) + "</p>"
		}
		full = `<article class="book-cover"><h1>` + html.EscapeString(b.cfg.Title) + `</h1>` + subtitle + `</article>` + full
	}

	bookHTML := wrapHTML(full, b.cfg.Title)
	b.stats.BrokenAnchors = countBrokenAnchors(bookHTML)

	return &Book{
		Title:    b.cfg.Title,
		Subtitle: b.cfg.Subtitle,
		HTML:     bookHTML,
		Stats:    b.stats,
	}, nil
}

func docLabel(d *source.Doc) string {
	text := d.Title
	if text == "" {
		text = d.H1
	}
	if text == "" {
		text = d.RelPath
	}
	if d.Code != "" {
		return d.Code + " " + text
	}
	return text
}

func (b *Builder) tocItem(id, label string, headings []markup.Heading, levels map[int]bool) string {
	var bd strings.Builder
	bd.WriteString(`<li><a href="#` + id + `">` + html.EscapeString(label) + `</a>`)
	if levels[2] {
		var subs []string
		for _, h := range headings {
			if h.Level != 2 {
				continue
			}
			subs = append(subs, `<li><a href="#`+h.ID+`">`+html.EscapeString(h.Text)+`</a></li>`)
		}
		if len(subs) > 0 {
			bd.WriteString(`<ul>` + strings.Join(subs, "\n") + `</ul>`)
		}
	}
	bd.WriteString(`</li>`)
	return bd.String()
}

// frontmatterTable renders the frontmatter fields the way GitHub shows them.
func frontmatterTable(front []source.KV) string {
	if len(front) == 0 {
		return ""
	}
	var bd strings.Builder
	bd.WriteString(`<table class="book-frontmatter"><tbody>`)
	for _, kv := range front {
		bd.WriteString("<tr><td><code>" + html.EscapeString(kv.Key) + "</code></td><td>")
		bd.WriteString(html.EscapeString(kv.Value))
		bd.WriteString("</td></tr>")
	}
	bd.WriteString(`</tbody></table>`)
	return bd.String()
}

// fixLinks rewrites internal cross-references to in-book anchors and strips
// links that point outside the book, keeping their label text.
func (b *Builder) fixLinks(docHTML string, d source.Doc) (string, error) {
	root, err := xhtml.Parse(strings.NewReader(docHTML))
	if err != nil {
		return "", err
	}
	var nodes []*xhtml.Node
	for n := range allNodes(root) {
		if n.Data == "a" || n.Data == "img" {
			nodes = append(nodes, n)
		}
	}
	for _, n := range nodes {
		switch n.Data {
		case "a":
			href := attr(n, "href")
			if href == "" || isExternal(href) {
				continue
			}
			p, _ := splitFragment(href)
			if p == "" {
				continue
			}
			rel, ok := resolveRel(b.kbRoot, d.AbsPath, p)
			if ok {
				if id, ok := b.relIDs[rel]; ok {
					b.stats.InternalLinks++
					setAttr(n, "href", "#"+id)
					continue
				}
			}
			b.stats.Flattened++
			flatten(n)
		case "img":
			src := attr(n, "src")
			if src == "" || isExternal(src) {
				continue
			}
			target := filepath.Clean(filepath.Join(filepath.Dir(d.AbsPath), filepath.FromSlash(src)))
			setAttr(n, "src", "file://"+filepath.ToSlash(target))
		}
	}

	var buf bytes.Buffer
	if err := xhtml.Render(&buf, root); err != nil {
		return "", err
	}
	return innerBody(buf.String()), nil
}

// innerBody returns everything between <body> and </body>.
func innerBody(s string) string {
	start := strings.Index(s, "<body>")
	if start < 0 {
		return s
	}
	start += len("<body>")
	end := strings.LastIndex(s, "</body>")
	if end < 0 {
		return s[start:]
	}
	return s[start:end]
}

func splitFragment(href string) (path, frag string) {
	if i := strings.IndexByte(href, '#'); i >= 0 {
		return href[:i], href[i:]
	}
	return href, ""
}

func isExternal(href string) bool {
	lower := strings.ToLower(href)
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "mailto:") ||
		strings.HasPrefix(lower, "tel:") ||
		strings.HasPrefix(lower, "data:") ||
		strings.HasPrefix(lower, "//") ||
		strings.HasPrefix(href, "/")
}

func flatten(n *xhtml.Node) {
	if n.Parent == nil {
		return
	}
	var b strings.Builder
	for c := range allNodes(n) {
		if c.Type == xhtml.TextNode {
			b.WriteString(c.Data)
		}
	}
	parent := n.Parent
	text := &xhtml.Node{Type: xhtml.TextNode, Data: b.String()}
	parent.InsertBefore(text, n)
	parent.RemoveChild(n)
}

func attr(n *xhtml.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

func setAttr(n *xhtml.Node, name, value string) {
	for i := range n.Attr {
		if n.Attr[i].Key == name {
			n.Attr[i].Val = value
			return
		}
	}
	n.Attr = append(n.Attr, xhtml.Attribute{Key: name, Val: value})
}

func allNodes(root *xhtml.Node) func(func(*xhtml.Node) bool) {
	return func(yield func(*xhtml.Node) bool) {
		var walk func(n *xhtml.Node) bool
		walk = func(n *xhtml.Node) bool {
			if !yield(n) {
				return false
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if !walk(c) {
					return false
				}
			}
			return true
		}
		walk(root)
	}
}

// wrapHTML builds the final standalone document with embedded styles.
func wrapHTML(body, title string) string {
	css := cssBundle()
	return "<!DOCTYPE html>\n<html lang=\"ru\">\n<head>\n<meta charset=\"utf-8\">\n" +
		"<title>" + html.EscapeString(title) + "</title>\n" +
		"<style>" + css + "</style>\n</head>\n<body>\n" + body + "\n</body>\n</html>\n"
}

// cssBundle returns the embedded stylesheet.
func cssBundle() string {
	return assets.CSSBundle()
}

// countBrokenAnchors counts href="#..." links whose target id is missing.
func countBrokenAnchors(doc string) int {
	root, err := xhtml.Parse(strings.NewReader(doc))
	if err != nil {
		return 0
	}
	ids := map[string]bool{}
	for n := range allNodes(root) {
		for _, a := range n.Attr {
			if a.Key == "id" && a.Val != "" {
				ids[a.Val] = true
			}
		}
	}
	broken := 0
	for n := range allNodes(root) {
		if n.Data != "a" {
			continue
		}
		for _, a := range n.Attr {
			if a.Key == "href" && strings.HasPrefix(a.Val, "#") {
				target := strings.TrimPrefix(a.Val, "#")
				if pos := strings.IndexByte(target, ':'); pos >= 0 {
					continue // skip scheme-like fragments
				}
				decoded, err := url.PathUnescape(target)
				if err == nil {
					target = decoded
				}
				if target != "" && !ids[target] {
					broken++
				}
			}
		}
	}
	return broken
}

// WriteHTMLFile writes the book to path.
func (b *Book) WriteHTMLFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.HTML), 0o644)
}
