package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abskrj/velane/services/mcp-server/internal/controlplane"
	"github.com/abskrj/velane/services/mcp-server/internal/protocol"
	"github.com/abskrj/velane/services/mcp-server/internal/server"
	"github.com/abskrj/velane/services/mcp-server/internal/tools"
)

func TestInitialize(t *testing.T) {
	srv := server.New(tools.NewRegistry(controlplane.New("http://localhost:1")))
	resp := srv.HandleRequest(context.Background(), "Bearer test", protocol.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("result should not be nil")
	}
	raw, _ := json.Marshal(resp.Result)
	if !strings.Contains(string(raw), "protocolVersion") {
		t.Fatalf("expected initialize response to include protocolVersion: %s", string(raw))
	}
	for _, capability := range []string{"tools", "resources", "prompts"} {
		if !strings.Contains(string(raw), capability) {
			t.Fatalf("expected initialize response to include %s capability: %s", capability, string(raw))
		}
	}
}

func TestToolsList(t *testing.T) {
	srv := server.New(tools.NewRegistry(controlplane.New("http://localhost:1")))
	resp := srv.HandleRequest(context.Background(), "Bearer test", protocol.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	raw, _ := json.Marshal(resp.Result)
	if !strings.Contains(string(raw), "list_workflows") {
		t.Fatalf("expected tool list to include list_workflows: %s", string(raw))
	}
}

func TestToolsCallListWorkflows(t *testing.T) {
	cp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test" {
			http.Error(w, `{"error":"bad auth"}`, http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/v1/snippets" {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"sn_1","slug":"hello"}]`))
	}))
	defer cp.Close()

	srv := server.New(tools.NewRegistry(controlplane.New(cp.URL)))
	params := map[string]any{
		"name":      "list_workflows",
		"arguments": map[string]any{},
	}
	pb, _ := json.Marshal(params)
	resp := srv.HandleRequest(context.Background(), "Bearer test", protocol.Request{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params:  pb,
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	if !strings.Contains(string(b), "structuredContent") {
		t.Fatalf("expected structuredContent in result: %s", string(b))
	}
	var parsed struct {
		StructuredContent map[string]any `json:"structuredContent"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if parsed.StructuredContent == nil {
		t.Fatalf("structuredContent should be an object: %s", string(b))
	}
	if _, ok := parsed.StructuredContent["workflows"]; !ok {
		t.Fatalf("expected workflows key in structuredContent: %s", string(b))
	}
}

func TestHandleJSONRPCEndpoint(t *testing.T) {
	cp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer cp.Close()

	srv := server.New(tools.NewRegistry(controlplane.New(cp.URL)))
	httpSrv := httptest.NewServer(srv.Router())
	defer httpSrv.Close()

	reqBody := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`
	req, _ := http.NewRequest(http.MethodPost, httpSrv.URL+"/mcp", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer test")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d; want 200", resp.StatusCode)
	}
}

func TestHandleJSONRPCInitializedNotification(t *testing.T) {
	srv := server.New(tools.NewRegistry(controlplane.New("http://localhost:1")))
	httpSrv := httptest.NewServer(srv.Router())
	defer httpSrv.Close()

	reqBody := `{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`
	req, _ := http.NewRequest(http.MethodPost, httpSrv.URL+"/mcp", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d; want 204", resp.StatusCode)
	}
}

func TestResourcesAndPromptsListAreEmpty(t *testing.T) {
	srv := server.New(tools.NewRegistry(controlplane.New("http://localhost:1")))
	for _, method := range []string{"resources/list", "prompts/list"} {
		resp := srv.HandleRequest(context.Background(), "Bearer test", protocol.Request{
			JSONRPC: "2.0",
			ID:      1,
			Method:  method,
		})
		if resp.Error != nil {
			t.Fatalf("%s returned error: %+v", method, resp.Error)
		}
		raw, _ := json.Marshal(resp.Result)
		if !strings.Contains(string(raw), strings.TrimSuffix(method, "/list")) {
			t.Fatalf("%s returned unexpected result: %s", method, string(raw))
		}
	}
}

func TestResourcesReadRuntimeContract(t *testing.T) {
	srv := server.New(tools.NewRegistry(controlplane.New("http://localhost:1")))
	params := map[string]any{"uri": "velane://runtime/contract"}
	pb, _ := json.Marshal(params)
	resp := srv.HandleRequest(context.Background(), "Bearer test", protocol.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "resources/read",
		Params:  pb,
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	raw, _ := json.Marshal(resp.Result)
	if !strings.Contains(string(raw), "Velane Runtime Contract") {
		t.Fatalf("expected runtime contract content: %s", string(raw))
	}
}

func TestResourcesReadWorkflowCatalogTruncates(t *testing.T) {
	cp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/snippets" {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"id":"1","slug":"one","name":"One","language":"bun","code":"should-not-leak"},
			{"id":"2","slug":"two","name":"Two","language":"python","code":"should-not-leak"}
		]`))
	}))
	defer cp.Close()

	srv := server.New(tools.NewRegistry(controlplane.New(cp.URL)))
	params := map[string]any{"uri": "velane://workflows"}
	pb, _ := json.Marshal(params)
	resp := srv.HandleRequest(context.Background(), "Bearer test", protocol.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "resources/read",
		Params:  pb,
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	raw, _ := json.Marshal(resp.Result)
	if strings.Contains(string(raw), "should-not-leak") {
		t.Fatalf("workflow catalog should not include code: %s", string(raw))
	}
	if !strings.Contains(string(raw), "one") || !strings.Contains(string(raw), "two") {
		t.Fatalf("expected compact workflow metadata: %s", string(raw))
	}
}

func TestPromptsGet(t *testing.T) {
	srv := server.New(tools.NewRegistry(controlplane.New("http://localhost:1")))
	params := map[string]any{
		"name": "create_integration_workflow",
		"arguments": map[string]any{
			"provider": "github",
			"goal":     "create an issue",
		},
	}
	pb, _ := json.Marshal(params)
	resp := srv.HandleRequest(context.Background(), "Bearer test", protocol.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "prompts/get",
		Params:  pb,
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	raw, _ := json.Marshal(resp.Result)
	if !strings.Contains(string(raw), "list_connections") || !strings.Contains(string(raw), "github") {
		t.Fatalf("expected workflow prompt content: %s", string(raw))
	}
}
