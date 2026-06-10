package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/abskrj/velane/services/mcp-server/internal/controlplane"
)

const (
	workflowCatalogLimit   = 50
	connectionCatalogLimit = 50
)

type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type Content struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

type Registry struct {
	client *controlplane.Client
}

func NewRegistry(client *controlplane.Client) *Registry {
	return &Registry{client: client}
}

func (r *Registry) List() []Resource {
	return []Resource{
		{
			URI:         "velane://runtime/contract",
			Name:        "Velane runtime contract",
			Description: "How to write workflows, use integrations, invoke safely, and interpret stdout/stderr.",
			MimeType:    "text/markdown",
		},
		{
			URI:         "velane://workflows",
			Name:        "Workflow catalog",
			Description: "Compact first page of workflows for this tenant. Use get_workflow for code and versions.",
			MimeType:    "application/json",
		},
		{
			URI:         "velane://connections",
			Name:        "Connected integrations",
			Description: "Compact first page of OAuth integrations connected for this tenant.",
			MimeType:    "application/json",
		},
	}
}

func (r *Registry) Read(ctx context.Context, authHeader, uri string) ([]Content, error) {
	switch strings.TrimSpace(uri) {
	case "velane://runtime/contract":
		return []Content{{
			URI:      uri,
			MimeType: "text/markdown",
			Text:     runtimeContract,
		}}, nil
	case "velane://workflows", "velane://snippets":
		text, err := r.readWorkflowCatalog(ctx, authHeader)
		if err != nil {
			return nil, err
		}
		return []Content{{URI: uri, MimeType: "application/json", Text: text}}, nil
	case "velane://connections":
		text, err := r.readConnectionCatalog(ctx, authHeader)
		if err != nil {
			return nil, err
		}
		return []Content{{URI: uri, MimeType: "application/json", Text: text}}, nil
	default:
		return nil, fmt.Errorf("unknown resource: %s", uri)
	}
}

func (r *Registry) readWorkflowCatalog(ctx context.Context, authHeader string) (string, error) {
	var workflows []map[string]any
	if err := r.client.Get(ctx, authHeader, "/v1/snippets", &workflows); err != nil {
		return "", err
	}

	total := len(workflows)
	if len(workflows) > workflowCatalogLimit {
		workflows = workflows[:workflowCatalogLimit]
	}

	items := make([]map[string]any, 0, len(workflows))
	for _, wf := range workflows {
		items = append(items, compact(wf, "id", "slug", "name", "language", "created_at"))
	}

	return marshalPretty(map[string]any{
		"items": items,
		"page": map[string]any{
			"limit":     workflowCatalogLimit,
			"returned":  len(items),
			"total":     total,
			"truncated": total > len(items),
		},
		"next_steps": []string{
			"Use get_workflow with workflow_id to read code, versions, and active environments.",
			"Use list_workflows when you need the raw tool response.",
		},
	})
}

func (r *Registry) readConnectionCatalog(ctx context.Context, authHeader string) (string, error) {
	var connections []map[string]any
	path := "/v1/connections" + controlplane.Query(map[string]string{"limit": fmt.Sprintf("%d", connectionCatalogLimit)})
	if err := r.client.Get(ctx, authHeader, path, &connections); err != nil {
		return "", err
	}

	items := make([]map[string]any, 0, len(connections))
	for _, conn := range connections {
		items = append(items, compact(conn, "id", "provider", "alias", "status", "created_at", "updated_at"))
	}

	return marshalPretty(map[string]any{
		"items": items,
		"page": map[string]any{
			"limit":    connectionCatalogLimit,
			"returned": len(items),
		},
		"next_steps": []string{
			"Use list_connections with provider/limit/offset to filter or paginate.",
			"Use get_integration_docs with provider before writing integration-heavy workflow code.",
		},
	})
}

func compact(src map[string]any, keys ...string) map[string]any {
	out := map[string]any{}
	for _, key := range keys {
		if value, ok := src[key]; ok {
			out[key] = value
		}
	}
	return out
}

func marshalPretty(value any) (string, error) {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal resource: %w", err)
	}
	return string(b), nil
}

const runtimeContract = `# Velane Runtime Contract

Velane workflows are tenant-scoped functions that run in Bun/TypeScript or Python.

## Handler shape

Bun workflows export a default async function or named handler:

` + "```ts" + `
export default async function handler(input: Record<string, unknown>) {
  return { ok: true, input }
}
` + "```" + `

Python workflows define handler:

` + "```python" + `
def handler(input: dict) -> dict:
    return {"ok": True, "input": input}
` + "```" + `

## Integrations

Use connected OAuth integrations through the built-in integration helper. Do not embed credentials in workflow code.

Bun:

` + "```ts" + `
import { integration } from '@velane/integrations'
const github = integration('github')
const user = await github.get('/user')
` + "```" + `

Python:

` + "```python" + `
from velane.integrations import integration
github = integration("github")
user = github.get("/user")
` + "```" + `

## Environments and logs

- Validate in dev first, then promote to staging or prod.
- Debug logs from print/console.log are live-only and are forwarded only in dev.
- Generator chunks and final results are forwarded in every environment.
- Stored invocation records include status, output, error, stderr, duration, and mode.

## MCP workflow

Prefer this order for integration-heavy work:

1. Read velane://connections or call list_connections.
2. Call get_integration_docs(provider).
3. Create or update a draft workflow.
4. Invoke in dev and inspect logs/output.
5. Publish the exact validated version_number.
`
