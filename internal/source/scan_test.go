package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSplitFrontmatter(t *testing.T) {
	data := "---\ntype: regulation\ncode: INF-R-001\nversion: \"1.0\"\n---\n\n# Title\n\ntext"
	front, body := SplitFrontmatter([]byte(data))
	if len(front) != 3 {
		t.Fatalf("got %d fields, want 3: %v", len(front), front)
	}
	if front[0] != (KV{Key: "type", Value: "regulation"}) {
		t.Errorf("front[0] = %+v", front[0])
	}
	if front[2].Value != "1.0" {
		t.Errorf("version should stay raw '1.0', got %q", front[2].Value)
	}
	if body == "" || !strings.Contains(body, "# Title") {
		t.Errorf("unexpected body %q", body)
	}
}

func TestScanOrdering(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "employees/README.md", "---\ntype: registry\n---\n# Registry\n")
	write(t, dir, "employees/zebra/README.md", "---\ncode: Z-001\n---\n# Z\n")
	write(t, dir, "employees/alpha.md", "---\ncode: A-001\n---\n# A\n")
	write(t, dir, "org/README.md", "---\ntype: registry\n---\n# Org\n")
	write(t, dir, "regulations/README.md", "---\ntype: index\n---\n# Reg\n")
	write(t, dir, "regulations/B/B-001.md", "---\ncode: B-001\n---\n# B1\n")
	write(t, dir, "regulations/A/A-001.md", "---\ncode: A-001\n---\n# A1\n")

	docs, err := Scan(dir, []string{"employees/**/*.md", "org/README.md", "regulations/**/*.md"})
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, d := range docs {
		got = append(got, d.RelPath)
	}
	want := []string{
		"employees/README.md",
		"employees/alpha.md",
		"employees/zebra/README.md",
		"org/README.md",
		"regulations/README.md",
		"regulations/A/A-001.md",
		"regulations/B/B-001.md",
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("order:\n got %v\nwant %v", got, want)
	}
}

func TestScanRespectsIncludeOrder(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "org/README.md", "# Org\n")
	write(t, dir, "employees/README.md", "# Employees\n")
	write(t, dir, "regulations/README.md", "# Regulations\n")
	write(t, dir, "regulations/B/B-001.md", "# B1\n")
	write(t, dir, "regulations/A/A-001.md", "# A1\n")
	write(t, dir, "employees/alpha.md", "# Alpha\n")

	docs, err := Scan(dir, []string{"regulations/**/*.md", "employees/**/*.md", "org/README.md"})
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, d := range docs {
		got = append(got, d.RelPath)
	}
	want := []string{
		"regulations/README.md",
		"regulations/A/A-001.md",
		"regulations/B/B-001.md",
		"employees/README.md",
		"employees/alpha.md",
		"org/README.md",
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("order:\n got %v\nwant %v", got, want)
	}
}

func TestScanDedupesOverlappingPatterns(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "team/alice/README.md", "# Alice\n")
	write(t, dir, "team/bob/README.md", "# Bob\n")

	docs, err := Scan(dir, []string{"team/alice/*.md", "team/**/*.md"})
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, d := range docs {
		got = append(got, d.RelPath)
	}
	if strings.Join(got, ",") != "team/alice/README.md,team/bob/README.md" {
		t.Fatalf("got %v, want first pattern to win and no duplicates", got)
	}
}

func TestClassifyPatterns(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "docs/org/README.md", "# Org\n")
	if err := os.MkdirAll(filepath.Join(dir, "docs", "team"), 0o755); err != nil {
		t.Fatal(err)
	}

	globs, linkRoots := ClassifyPatterns(dir, []string{
		"docs/**/*.md",
		"docs/org/README.md",
		"docs/team/",
		"docs/team",
		"missing/dir",
	})
	wantGlobs := []string{"docs/**/*.md", "docs/org/README.md", "missing/dir"}
	wantRoots := []string{"docs/team", "docs/team"}
	if strings.Join(globs, ",") != strings.Join(wantGlobs, ",") {
		t.Errorf("globs = %v, want %v", globs, wantGlobs)
	}
	if strings.Join(linkRoots, ",") != strings.Join(wantRoots, ",") {
		t.Errorf("linkRoots = %v, want %v", linkRoots, wantRoots)
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "regs/INF-R-001.md", "---\ncode: INF-R-001\ntitle: Инфраструктура\n---\n\n<!-- refs -->\n\n# Заголовок\n\nтекст\n")
	d, err := Load(dir, "regs/INF-R-001.md")
	if err != nil {
		t.Fatal(err)
	}
	if d.Code != "INF-R-001" {
		t.Errorf("Code = %q", d.Code)
	}
	if d.Title != "Инфраструктура" {
		t.Errorf("Title = %q", d.Title)
	}
	if d.H1 != "Заголовок" {
		t.Errorf("H1 = %q", d.H1)
	}
	if strings.Contains(d.Body, "<!-- refs -->") {
		t.Error("HTML comment leaked into Body")
	}
}

func TestExclude(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a/README.md", "# A\n")
	write(t, dir, "a/X.md", "# X\n")
	write(t, dir, "a/sub/Y.md", "# Y\n")
	write(t, dir, "a/sub/deep/Z.md", "# Z\n")
	write(t, dir, "b/B.md", "# B\n")

	docs, err := Scan(dir, []string{"a/**/*.md", "b/*.md"})
	if err != nil {
		t.Fatal(err)
	}
	got := rels(docs)
	if strings.Join(got, ",") != "a/README.md,a/X.md,a/sub/Y.md,a/sub/deep/Z.md,b/B.md" {
		t.Fatalf("unexpected scan result %v", got)
	}

	cases := []struct {
		name string
		excl []string
		want string
	}{
		{"file without suffix", []string{"a/X"}, "a/README.md,a/sub/Y.md,a/sub/deep/Z.md,b/B.md"},
		{"file with suffix", []string{"a/X.md"}, "a/README.md,a/sub/Y.md,a/sub/deep/Z.md,b/B.md"},
		{"folder", []string{"a/sub/"}, "a/README.md,a/X.md,b/B.md"},
		{"mask", []string{"**/Y.md"}, "a/README.md,a/X.md,a/sub/deep/Z.md,b/B.md"},
		{"mask with suffix", []string{"a/sub/*"}, "a/README.md,a/X.md,a/sub/deep/Z.md,b/B.md"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			filtered, err := Exclude(dir, append([]Doc(nil), docs...), c.excl)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Join(rels(filtered), ",") != c.want {
				t.Errorf("Exclude(%v) = %v, want %v", c.excl, rels(filtered), c.want)
			}
		})
	}
}

func rels(docs []Doc) []string {
	out := make([]string, 0, len(docs))
	for _, d := range docs {
		out = append(out, d.RelPath)
	}
	return out
}
