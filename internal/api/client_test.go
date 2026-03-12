package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Push(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/sync/hiasinho/my-project" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-specter-token") != "test-token" {
			t.Errorf("missing or wrong token header")
		}

		var req SyncPushRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Branch != "main" {
			t.Errorf("expected branch 'main', got %q", req.Branch)
		}
		if len(req.Documents) != 1 {
			t.Errorf("expected 1 document, got %d", len(req.Documents))
		}

		json.NewEncoder(w).Encode(SyncPushResponse{
			Created:   []string{"specs/api.md"},
			Updated:   []string{},
			Unchanged: []string{},
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL

	result, err := client.Push("hiasinho/my-project", &SyncPushRequest{
		Branch: "main",
		Documents: []Document{
			{Path: "specs/api.md", ContentMD: "# API"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Created) != 1 {
		t.Errorf("expected 1 created, got %d", len(result.Created))
	}
}

func TestClient_Pull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/sync/hiasinho/my-project" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("branch") != "main" {
			t.Errorf("expected branch=main, got %q", r.URL.Query().Get("branch"))
		}

		json.NewEncoder(w).Encode(SyncPullResponse{
			Documents: []Document{
				{Path: "specs/api.md", ContentMD: "# API", ContentHash: "abc", Revision: 5},
			},
			SyncedAt: "2025-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL

	result, err := client.Pull("hiasinho/my-project", "main", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Documents) != 1 {
		t.Errorf("expected 1 document, got %d", len(result.Documents))
	}
	if result.Documents[0].Revision != 5 {
		t.Errorf("expected revision 5, got %d", result.Documents[0].Revision)
	}
}

func TestClient_PullWithSince(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		since := r.URL.Query().Get("since")
		if since != "3" {
			t.Errorf("expected since=3, got %q", since)
		}
		json.NewEncoder(w).Encode(SyncPullResponse{
			Documents: []Document{},
			SyncedAt:  "2025-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL

	sinceRev := 3
	_, err := client.Pull("hiasinho/my-project", "main", &sinceRev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ListDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/documents/hiasinho/my-project" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]Document{
			{Path: "specs/api.md", ContentMD: "# API", ContentHash: "abc", Revision: 1},
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL

	docs, err := client.ListDocuments("hiasinho/my-project", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("expected 1 document, got %d", len(docs))
	}
}

func TestClient_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid token"}`))
	}))
	defer server.Close()

	client := NewClient("bad-token")
	client.BaseURL = server.URL

	_, err := client.Push("hiasinho/my-project", &SyncPushRequest{Branch: "main"})
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestClient_PushConflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Conflicts detected",
			"conflicts": []map[string]any{
				{"path": "specs/api.md", "server_revision": 5, "server_updated_at": "2025-01-01T00:00:00Z", "server_hash": "abc123"},
			},
			"created":   []string{},
			"updated":   []string{},
			"unchanged": []string{},
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL

	_, err := client.Push("hiasinho/my-project", &SyncPushRequest{Branch: "main"})
	if err == nil {
		t.Fatal("expected error for 409 response")
	}

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected ConflictError, got %T: %v", err, err)
	}
	if len(conflict.Conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(conflict.Conflicts))
	}
	if conflict.Conflicts[0].Path != "specs/api.md" {
		t.Errorf("expected conflict path 'specs/api.md', got %q", conflict.Conflicts[0].Path)
	}
}

func TestClient_PushWithBaseRevision(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req SyncPushRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.BaseRevision != "2025-01-01T00:00:00Z" {
			t.Errorf("expected base_revision '2025-01-01T00:00:00Z', got %q", req.BaseRevision)
		}
		json.NewEncoder(w).Encode(SyncPushResponse{
			Created:   []string{},
			Updated:   []string{"specs/api.md"},
			Unchanged: []string{},
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL

	result, err := client.Push("hiasinho/my-project", &SyncPushRequest{
		Branch:       "main",
		BaseRevision: "2025-01-01T00:00:00Z",
		Documents:    []Document{{Path: "specs/api.md", ContentMD: "# API"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Updated) != 1 {
		t.Errorf("expected 1 updated, got %d", len(result.Updated))
	}
}
