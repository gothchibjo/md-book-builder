package book

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gothchibjo/md-book-builder/internal/config"
)

func writeKB(t *testing.T, dir string) {
	t.Helper()
	files := map[string]string{
		"regulations/README.md": "---\ntype: index\n---\n# Регламенты\n",
		"regulations/INF/INF-R-001.md": `---
type: regulation
code: INF-R-001
title: Управление инфраструктурой
---

# 1. Общие положения

См. [ITG-P-001]. Также [реестр сотрудников].

## Термины

Подробное описание в [INF-R-002].

<!-- refs -->

[ITG-P-001]: ../ITG/ITG-P-001.md
[реестр сотрудников]: ../../employees/README.md
[INF-R-002]: ./INF-R-002.md
`,
		"regulations/INF/INF-R-002.md": `---
type: regulation
code: INF-R-002
title: Мониторинг
---

# Инфраструктура — мониторинг

Внутренняя ссылка: [INF-R-001] и якорь [раздел](#1-общие-положения).

<!-- refs -->

[INF-R-001]: ./INF-R-001.md
`,
		"employees/README.md": `---
type: registry
---
# Реестр сотрудников

[Иванов Пётр]

<!-- refs -->

[Иванов Пётр]: ./i-petrov/README.md
`,
	}
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBuildAndLinkResolution(t *testing.T) {
	dir := t.TempDir()
	writeKB(t, dir)

	cfg := &config.Config{
		Title:     "Example",
		KBRoot:    dir,
		Out:       filepath.Join(dir, "out.pdf"),
		TOCLevels: []int{1, 2},
		Include:   []string{"regulations/**/*.md", "employees/README.md"},
	}
	b, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}

	st := doc.Stats
	if st.Documents != 4 {
		t.Errorf("Documents = %d, want 4", st.Documents)
	}
	if st.FrontTables != 4 {
		t.Errorf("FrontTables = %d, want 4", st.FrontTables)
	}
	if st.TOCEntries != 4 {
		t.Errorf("TOCEntries = %d, want 4", st.TOCEntries)
	}
	// INF-R-001: [ITG-P-001] outside -> flattened, [реестр сотрудников] and
	// [INF-R-002] inside -> internal. INF-R-002: [INF-R-001] -> internal.
	// employees/README: [Иванов Пётр] outside -> flattened.
	if st.InternalLinks != 3 {
		t.Errorf("InternalLinks = %d, want 3", st.InternalLinks)
	}
	if st.Flattened != 2 {
		t.Errorf("Flattened = %d, want 2", st.Flattened)
	}
	if st.BrokenAnchors != 0 {
		t.Errorf("BrokenAnchors = %d, want 0", st.BrokenAnchors)
	}

	// Order: cover first, then TOC, then the document body.
	cover := strings.Index(doc.HTML, `<article class="book-cover">`)
	toc := strings.Index(doc.HTML, `<nav class="book-toc">`)
	main := strings.Index(doc.HTML, `<main class="markdown-body">`)
	if !(cover >= 0 && cover < toc && toc < main) {
		t.Errorf("expected cover before TOC before body, got cover=%d toc=%d main=%d", cover, toc, main)
	}

	// The flattened reference keeps its label text but is no longer a link.
	if !strings.Contains(doc.HTML, "ITG-P-001") {
		t.Error("flattened label text missing from output")
	}
	if r := strings.Count(doc.HTML, `href="#inf-r-001"`); r < 1 {
		t.Errorf("expected internal link to inf-r-001, got %d", r)
	}
	// With toc_levels [1,2], the H2 'Термины' appears in the TOC.
	if toc := tocText(doc.HTML); !strings.Contains(toc, "Термины") {
		t.Error("H2 heading missing from TOC with toc_levels [1,2]")
	}
}

func tocText(docHTML string) string {
	start := strings.Index(docHTML, `<nav class="book-toc">`)
	if start < 0 {
		return ""
	}
	start += len(`<nav class="book-toc">`)
	end := strings.Index(docHTML[start:], "</nav>")
	if end < 0 {
		return ""
	}
	return docHTML[start : start+end]
}

func TestTOCLevelsDocumentsOnly(t *testing.T) {
	dir := t.TempDir()
	writeKB(t, dir)
	cfg := &config.Config{
		Title:     "Example",
		KBRoot:    dir,
		Out:       filepath.Join(dir, "out.pdf"),
		TOCLevels: []int{1},
		Include:   []string{"regulations/INF/INF-R-001.md"},
	}
	b, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(tocText(doc.HTML), "Термины") {
		t.Error("H2 heading leaked into TOC with toc_levels [1]")
	}
}

func writeLinkRootKB(t *testing.T, dir string) {
	t.Helper()
	files := map[string]string{
		"regulations/README.md": "---\ntype: index\n---\n# Регламенты\n",
		"regulations/INF/INF-R-001.md": `---
type: regulation
code: INF-R-001
title: Управление инфраструктурой
---

# Общие положения

См. [ITG-P-001] и [Z-001].

<!-- refs -->

[ITG-P-001]: ../../itg/ITG-P-001.md
[Z-001]: ../../zzz/Z-001.md
`,
		"itg/ITG-P-001.md": `---
type: regulation
code: ITG-P-001
---

# Политика 1

Продолжение в [ITG-P-002].

<!-- refs -->

[ITG-P-002]: ./ITG-P-002.md
`,
		"itg/ITG-P-002.md": `---
type: regulation
code: ITG-P-002
---

# Политика 2

Продолжение в [ITG-P-003].

<!-- refs -->

[ITG-P-003]: ./ITG-P-003.md
`,
		"itg/ITG-P-003.md": `---
type: regulation
code: ITG-P-003
---

# Политика 3
`,
		"zzz/Z-001.md": "---\ntype: regulation\ncode: Z-001\n---\n# Чужой документ\n",
	}
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func boolPtr(b bool) *bool { return &b }

func TestLinkRootCollectionTransitive(t *testing.T) {
	dir := t.TempDir()
	writeLinkRootKB(t, dir)

	cfg := &config.Config{
		Title:     "Example",
		KBRoot:    dir,
		Out:       filepath.Join(dir, "out.pdf"),
		TOCLevels: []int{1},
		Include:   []string{"regulations/**/*.md", "itg/"},
	}
	b, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}

	st := doc.Stats
	if st.Documents != 5 {
		t.Errorf("Documents = %d, want 5", st.Documents)
	}
	if st.InternalLinks != 3 {
		t.Errorf("InternalLinks = %d, want 3", st.InternalLinks)
	}
	if st.Flattened != 1 {
		t.Errorf("Flattened = %d, want 1", st.Flattened)
	}
	if st.BrokenAnchors != 0 {
		t.Errorf("BrokenAnchors = %d, want 0", st.BrokenAnchors)
	}
	if strings.Contains(doc.HTML, `href="#z-001"`) {
		t.Error("link to zzz/Z-001 must stay flattened (not a link-root)")
	}

	// Pulled documents are appended after the glob matches, sorted by path.
	order := []string{"regulations-readme", "inf-r-001", "itg-p-001", "itg-p-002", "itg-p-003"}
	idx := make([]int, 0, len(order))
	for _, id := range order {
		i := strings.Index(doc.HTML, `<section class="book-doc" id="`+id+`"`)
		if i < 0 {
			t.Fatalf("section %q missing from output", id)
		}
		idx = append(idx, i)
	}
	for i := 1; i < len(idx); i++ {
		if idx[i] < idx[i-1] {
			t.Fatalf("sections out of order: %v", order)
		}
	}
}

func TestLinkRootCollectionDirectOnly(t *testing.T) {
	dir := t.TempDir()
	writeLinkRootKB(t, dir)

	cfg := &config.Config{
		Title:           "Example",
		KBRoot:          dir,
		Out:             filepath.Join(dir, "out.pdf"),
		TOCLevels:       []int{1},
		Include:         []string{"regulations/**/*.md", "itg/"},
		TransitiveLinks: boolPtr(false),
	}
	b, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}

	st := doc.Stats
	if st.Documents != 3 {
		t.Errorf("Documents = %d, want 3", st.Documents)
	}
	if st.InternalLinks != 1 {
		t.Errorf("InternalLinks = %d, want 1", st.InternalLinks)
	}
	if st.Flattened != 2 {
		t.Errorf("Flattened = %d, want 2", st.Flattened)
	}
}

func TestNoTOCAndNoCover(t *testing.T) {
	dir := t.TempDir()
	writeKB(t, dir)
	cfg := &config.Config{
		Title:     "Example",
		KBRoot:    dir,
		Out:       filepath.Join(dir, "out.pdf"),
		TOCLevels: []int{1},
		TOC:       boolPtr(false),
		Cover:     boolPtr(false),
		Include:   []string{"regulations/INF/INF-R-001.md"},
	}
	b, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(doc.HTML, `<nav class="book-toc">`) {
		t.Error("TOC rendered despite toc: false")
	}
	if strings.Contains(doc.HTML, `<article class="book-cover">`) {
		t.Error("cover rendered despite cover: false")
	}
	if !strings.Contains(doc.HTML, `<main class="markdown-body">`) {
		t.Error("document body missing")
	}
	if doc.Stats.TOCEntries != 0 {
		t.Errorf("TOCEntries = %d, want 0", doc.Stats.TOCEntries)
	}
}

func TestExcludeAfterCollection(t *testing.T) {
	dir := t.TempDir()
	writeLinkRootKB(t, dir)

	cfg := &config.Config{
		Title:     "Example",
		KBRoot:    dir,
		Out:       filepath.Join(dir, "out.pdf"),
		TOCLevels: []int{1},
		Include:   []string{"regulations/**/*.md", "itg/"},
		Exclude:   []string{"itg/ITG-P-002"},
	}
	b, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}

	st := doc.Stats
	if st.Documents != 4 {
		t.Errorf("Documents = %d, want 4", st.Documents)
	}
	if strings.Contains(doc.HTML, `<section class="book-doc" id="itg-p-002"`) {
		t.Error("excluded document itg/ITG-P-002.md still in the book")
	}
	if st.InternalLinks != 1 {
		t.Errorf("InternalLinks = %d, want 1", st.InternalLinks)
	}
	if st.Flattened != 2 {
		t.Errorf("Flattened = %d, want 2", st.Flattened)
	}
}
