package httpserver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTemplates(t *testing.T) {
	t.Run("LoadTemplates success", func(t *testing.T) {
		dir := t.TempDir()

		layoutPath := filepath.Join(dir, "layout.html")
		layoutContent := `<!DOCTYPE html><html><body>{{ block "content" . }}{{ end }}</body></html>`

		if err := os.WriteFile(layoutPath, []byte(layoutContent), 0o600); err != nil {
			t.Fatalf("failed to write layout.html: %v", err)
		}

		tpls, err := LoadTemplates(dir)
		if err != nil {
			t.Fatalf("LoadTemplates returned error: %v", err)
		}

		if tpls == nil {
			t.Fatalf("expected not nil Templates: %v", err)
		}
		if tpls.index == nil {
			t.Fatalf("expected index template to be parsed: %v", err)
		}
	})

	t.Run("LoadTemplates missing layout", func(t *testing.T) {
		dir := t.TempDir()

		tpls, err := LoadTemplates(dir)
		if err == nil {
			t.Fatal("expected missing layout.html, got nil")
		}
		if tpls != nil {
			t.Fatalf("expected nil templates; got error: %v", err)
		}
	})
}
