package source

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

var (
	commentRe = regexp.MustCompile(`<!--.*?-->`)
	h1Re      = regexp.MustCompile(`(?m)^\s*#\s+(.+)$`)
)

// KV is an ordered frontmatter key/value pair with the raw value text.
type KV struct {
	Key   string
	Value string
}

// Doc is one markdown file included in the book.
type Doc struct {
	RelPath string // forward slashes, relative to the kb root
	AbsPath string
	Code    string // frontmatter `code`, may be empty
	Title   string // frontmatter `title`, may be empty
	H1      string // first H1 heading text, may be empty
	Front   []KV   // ordered frontmatter fields
	Body    string // markdown body without frontmatter or comments
}

// Scan resolves the include globs under root and returns the matched docs in
// the order the patterns appear in include. Within one pattern the matches
// are ordered by directory (alphabetical), then by file name with README.md
// first within each directory. A file matching several patterns appears only
// once, in the group of the first matching pattern.
func Scan(root string, include []string) ([]Doc, error) {
	seen := map[string]bool{}
	var rels []string
	for _, pattern := range include {
		pat := filepath.ToSlash(pattern)
		matches, err := doublestar.FilepathGlob(filepath.Join(root, filepath.FromSlash(pat)))
		if err != nil {
			return nil, err
		}
		var group []string
		for _, m := range matches {
			rel, err := filepath.Rel(root, m)
			if err != nil {
				continue
			}
			rel = filepath.ToSlash(rel)
			if seen[rel] {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(rel), ".md") {
				continue
			}
			seen[rel] = true
			group = append(group, rel)
		}
		SortRels(group)
		rels = append(rels, group...)
	}

	docs := make([]Doc, 0, len(rels))
	for _, rel := range rels {
		d, err := Load(root, rel)
		if err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}

// SortRels orders relative document paths by directory (alphabetical), then
// by file name with README.md first within each directory.
func SortRels(rels []string) {
	sort.SliceStable(rels, func(i, j int) bool {
		a, b := rels[i], rels[j]
		da, db := filepath.Dir(a), filepath.Dir(b)
		if da != db {
			return da < db
		}
		ra, rb := isReadme(a), isReadme(b)
		if ra != rb {
			return ra
		}
		return a < b
	})
}

// Load reads one markdown document by its path relative to root.
func Load(root, rel string) (Doc, error) {
	abs := filepath.Join(root, filepath.FromSlash(rel))
	data, err := os.ReadFile(abs)
	if err != nil {
		return Doc{}, err
	}
	front, body := SplitFrontmatter(data)
	kv := keyvals(front)
	d := Doc{
		RelPath: rel,
		AbsPath: abs,
		Code:    kv["code"],
		Title:   kv["title"],
		Front:   front,
		Body:    stripComments(string(body)),
	}
	d.H1 = firstH1(d.Body)
	return d, nil
}

// ClassifyPatterns splits include patterns into glob patterns (entries with
// meta characters such as `*`, `?`, `[`, `{`) and link-root directories
// (plain directory entries without a mask, which are never browsed for
// documents but act as allowed targets when following internal links).
func ClassifyPatterns(root string, include []string) (globs, linkRoots []string) {
	for _, pattern := range include {
		pat := filepath.ToSlash(filepath.Clean(pattern))
		if strings.ContainsAny(pat, "*?[{") {
			globs = append(globs, pat)
			continue
		}
		if strings.HasSuffix(pattern, "/") || isDir(filepath.Join(root, filepath.FromSlash(pat))) {
			linkRoots = append(linkRoots, pat)
			continue
		}
		globs = append(globs, pat)
	}
	return globs, linkRoots
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func isReadme(p string) bool {
	return strings.EqualFold(filepath.Base(p), "README.md")
}

// SplitFrontmatter splits a file into its ordered frontmatter fields and the
// remaining body. It returns nil for front when the file has no frontmatter.
func SplitFrontmatter(data []byte) ([]KV, string) {
	body := string(data)
	if !strings.HasPrefix(body, "---") {
		return nil, body
	}
	lines := strings.Split(body, "\n")
	if len(lines) < 3 {
		return nil, body
	}
	var front []string
	restStart := 0
	closed := false
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			restStart = i + 1
			closed = true
			break
		}
		front = append(front, lines[i])
	}
	if !closed {
		return nil, body
	}
	rest := strings.Join(lines[restStart:], "\n")
	if len(rest) > 0 && !strings.HasPrefix(rest, "\n") {
		rest = "\n" + rest
	}
	return parseFront(front), rest
}

func parseFront(lines []string) []KV {
	var out []KV
	for _, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), "#") || strings.TrimSpace(ln) == "" {
			continue
		}
		if idx := strings.Index(ln, ":"); idx >= 0 {
			key := strings.TrimSpace(ln[:idx])
			val := strings.TrimSpace(ln[idx+1:])
			val = strings.Trim(val, "\"'")
			out = append(out, KV{Key: key, Value: val})
		}
	}
	return out
}

func keyvals(kvs []KV) map[string]string {
	m := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		if _, ok := m[kv.Key]; !ok {
			m[kv.Key] = kv.Value
		}
	}
	return m
}

func firstH1(body string) string {
	m := h1Re.FindSubmatch([]byte(body))
	if m == nil {
		return ""
	}
	return strings.TrimSpace(string(m[1]))
}

// stripComments removes HTML comments. They are used in the source docs only
// as markers (such as `<!-- refs -->`) and would otherwise leak into output.
func stripComments(s string) string {
	return commentRe.ReplaceAllString(s, "")
}
