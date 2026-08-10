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
