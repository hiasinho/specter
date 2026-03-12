package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Push(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/sync/my-project" {
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

	result, err := client.Push("my-project", &SyncPushRequest{
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
		if r.URL.Path != "/sync/my-project" {
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

	result, err := client.Pull("my-project", "main", nil)
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
	_, err := client.Pull("my-project", "main", &sinceRev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ListDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/documents/my-project" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]Document{
			{Path: "specs/api.md", ContentMD: "# API", ContentHash: "abc", Revision: 1},
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL

	docs, err := client.ListDocuments("my-project", "main")
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

	_, err := client.Push("my-project", &SyncPushRequest{Branch: "main"})
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}
