package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDefaultsAndResolution(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "book.yaml")
	content := "title: Example\nkb_root: kb\nout: build/out.pdf\ninclude:\n  - docs/**/*.md\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.TOCLevels) != 1 || c.TOCLevels[0] != 1 {
		t.Fatalf("TOCLevels should default to [1], got %v", c.TOCLevels)
	}
	wantRoot := filepath.Join(dir, "kb")
	if c.KBRoot != wantRoot {
		t.Errorf("KBRoot = %q, want %q", c.KBRoot, wantRoot)
	}
	if c.Out != filepath.Join(dir, "build", "out.pdf") {
		t.Errorf("Out = %q", c.Out)
	}
}

func TestLoadRequiresFields(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "book.yaml")
	if err := os.WriteFile(cfgPath, []byte("title: X\nkb_root: kb\ninclude: [a.md]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(cfgPath); err == nil || !strings.Contains(err.Error(), "out is required") {
		t.Fatalf("expected 'out is required' error, got %v", err)
	}
}
