package providers_test

import (
	"testing"

	"github.com/abskrj/velane/services/control-plane/internal/providers"
)

func TestResolveKey(t *testing.T) {
	key, ok := providers.ResolveKey("GitHub")
	if !ok || key != "github" {
		t.Fatalf("ResolveKey(GitHub) = %q, %v; want github, true", key, ok)
	}

	key, ok = providers.ResolveKey("google sheets")
	if !ok || key != "google-sheets" {
		t.Fatalf("ResolveKey(google sheets) = %q, %v; want google-sheets, true", key, ok)
	}

	_, ok = providers.ResolveKey("nonexistent-xyz")
	if ok {
		t.Fatal("expected unknown provider")
	}
}

func TestDocsURL(t *testing.T) {
	if u := providers.DocsURL("github"); u == "" {
		t.Fatal("expected docs URL for github")
	}
	if u := providers.DocsURL("figma"); u != "https://nango.dev/docs/api-integrations/figma" {
		t.Fatalf("figma docs URL = %q", u)
	}
}
