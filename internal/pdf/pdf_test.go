package pdf

import (
	"strings"
	"testing"
)

func TestFooterTemplateFor(t *testing.T) {
	tests := []struct {
		locale string
		wantEn bool
	}{
		{"", true},
		{"en", true},
		{"ru", false},
		{"fr", true},
	}
	for _, tt := range tests {
		got := FooterTemplateFor(tt.locale)
		if strings.Contains(got, "Page ") != tt.wantEn {
			t.Errorf("FooterTemplateFor(%q) English marker mismatch: %q", tt.locale, got)
		}
		if !strings.Contains(got, `<span class="pageNumber"></span>`) ||
			!strings.Contains(got, `<span class="totalPages"></span>`) {
			t.Errorf("FooterTemplateFor(%q) missing page placeholders: %q", tt.locale, got)
		}
	}
	if !strings.Contains(FooterTemplateFor("ru"), "Стр. ") {
		t.Errorf("FooterTemplateFor(\"ru\") should contain the Russian label")
	}
}
