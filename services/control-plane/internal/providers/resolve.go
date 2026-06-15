package providers

import (
	"strings"
)

// ResolveKey maps a user-supplied provider label to the canonical catalog slug.
// Returns the slug and whether it is known to the static catalog or bundled docs.
func ResolveKey(input string) (string, bool) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return "", false
	}
	key := strings.ToLower(raw)

	if Get(key) != nil {
		return key, true
	}
	for _, c := range Catalog {
		if c.UniqueKey == key || strings.EqualFold(c.Name, raw) {
			return c.UniqueKey, true
		}
	}

	// Common aliases: "Google Sheets" → google-sheets
	hyphenated := strings.ReplaceAll(key, " ", "-")
	hyphenated = strings.ReplaceAll(hyphenated, "_", "-")
	for _, c := range Catalog {
		if c.UniqueKey == hyphenated {
			return c.UniqueKey, true
		}
	}

	return key, false
}

// DocsURL returns the Nango documentation URL for a provider slug.
func DocsURL(providerKey string) string {
	if doc := Get(providerKey); doc != nil && doc.DocsURL != "" {
		return doc.DocsURL
	}
	return "https://nango.dev/docs/api-integrations/" + providerKey
}
