package markup

import (
	"strings"
	"testing"
)

func TestSlug(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1. Общие положения", "1-общие-положения"},
		{"Органиграмма", "органиграмма"},
		{"Каталог ИТ-услуг", "каталог-ит-услуг"},
		{"Реестр поставщиков, договоров и датацентров", "реестр-поставщиков-договоров-и-датацентров"},
		{"  Hello   World! ", "hello-world"},
		{"A/B.C", "abc"},
	}
	for _, c := range cases {
		if got := Slug(c.in); got != c.want {
			t.Errorf("Slug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestAlertTransformation(t *testing.T) {
	r := NewRenderer(nil)
	html, _, err := r.Render("> [!NOTE]\n>\n> First link list item.\n")
	if err != nil {
		t.Fatal(err)
	}
	want := "markdown-alert markdown-alert-note"
	if !strings.Contains(html, want) {
		t.Fatalf("alert class %q not found in %q", want, html)
	}
	if strings.Contains(html, "[!NOTE]") {
		t.Fatalf("marker leaked into output: %q", html)
	}
}

func TestHeadingIDs(t *testing.T) {
	r := NewRenderer([]string{"reserved-id"})
	html, headings, err := r.Render("# Первый\n# Первый\n## Первый\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(headings) != 3 {
		t.Fatalf("got %d headings, want 3", len(headings))
	}
	ids := []string{headings[0].ID, headings[1].ID, headings[2].ID}
	if ids[0] != "первый" || ids[1] != "первый-1" || ids[2] != "первый-2" {
		t.Fatalf("unexpected ids %v", ids)
	}
	for _, want := range ids {
		if !strings.Contains(html, `id="`+want+`"`) {
			t.Fatalf("id %q missing in output", want)
		}
	}
}
