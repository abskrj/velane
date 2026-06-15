package nangodocs_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abskrj/velane/services/control-plane/internal/nangodocs"
)

func TestFetchMarkdownFromURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ".md") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte("# Test provider\n\nNango integration guide with OAuth setup and common API gotchas for agents."))
	}))
	defer srv.Close()

	body, err := nangodocs.FetchMarkdown(context.Background(), srv.URL+"/docs/api-integrations/test-provider.md", "test-provider")
	if err != nil {
		t.Fatalf("FetchMarkdown: %v", err)
	}
	if !strings.Contains(body, "Test provider") {
		t.Fatalf("unexpected body: %q", body)
	}
}
