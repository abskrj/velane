package nangodocs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	nangoDocsOrigin = "https://nango.dev"
	maxMarkdownSize = 512 << 10 // 512 KiB
	cacheTTL        = 24 * time.Hour
)

type cacheEntry struct {
	body      string
	fetchedAt time.Time
}

var (
	cacheMu sync.RWMutex
	cache   = map[string]cacheEntry{}
	client  = &http.Client{Timeout: 20 * time.Second}
)

// FetchMarkdown retrieves Nango's public provider documentation as markdown.
// It uses the docs URL from Nango's provider API when available, then falls back
// to the canonical api-integrations / integrations/all paths documented in
// https://nango.dev/docs/llms.txt
func FetchMarkdown(ctx context.Context, docsURL, providerSlug string) (string, error) {
	for _, url := range candidateURLs(docsURL, providerSlug) {
		if body, ok := cached(url); ok {
			return body, nil
		}
		body, err := fetchURL(ctx, url)
		if err != nil || strings.TrimSpace(body) == "" {
			continue
		}
		putCache(url, body)
		return body, nil
	}
	return "", fmt.Errorf("nango docs not found for provider %q", providerSlug)
}

func candidateURLs(docsURL, slug string) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}

	if docsURL != "" {
		add(docsURL)
		if !strings.HasSuffix(docsURL, ".md") {
			add(docsURL + ".md")
		}
	}

	add(fmt.Sprintf("%s/docs/api-integrations/%s.md", nangoDocsOrigin, slug))
	add(fmt.Sprintf("%s/docs/integrations/all/%s.md", nangoDocsOrigin, slug))
	add(fmt.Sprintf("%s/docs/api-integrations/%s", nangoDocsOrigin, slug))
	add(fmt.Sprintf("%s/docs/integrations/all/%s", nangoDocsOrigin, slug))
	return out
}

func fetchURL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/markdown, text/plain, */*")
	req.Header.Set("User-Agent", "velane-control-plane/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxMarkdownSize))
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(string(body))
	if len(text) < 80 {
		return "", fmt.Errorf("response too short")
	}
	return text, nil
}

func cached(url string) (string, bool) {
	cacheMu.RLock()
	entry, ok := cache[url]
	cacheMu.RUnlock()
	if !ok || time.Since(entry.fetchedAt) > cacheTTL {
		return "", false
	}
	return entry.body, true
}

func putCache(url, body string) {
	cacheMu.Lock()
	cache[url] = cacheEntry{body: body, fetchedAt: time.Now()}
	cacheMu.Unlock()
}
