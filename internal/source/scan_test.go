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
