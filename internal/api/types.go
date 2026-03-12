package api

// Document represents a synced markdown document.
type Document struct {
	Path        string `json:"path"`
	ContentMD   string `json:"content_md"`
	ContentHash string `json:"content_hash,omitempty"`
	Revision    int    `json:"revision,omitempty"`
}

// SyncPushRequest is the body for POST /projects/:slug/sync.
type SyncPushRequest struct {
	Branch    string     `json:"branch"`
	Documents []Document `json:"documents"`
}

// SyncPushResponse is the response from POST /projects/:slug/sync.
type SyncPushResponse struct {
	Created   []string `json:"created"`
	Updated   []string `json:"updated"`
	Unchanged []string `json:"unchanged"`
}

// SyncPullResponse is the response from GET /projects/:slug/sync.
type SyncPullResponse struct {
	Documents []Document `json:"documents"`
	SyncedAt  string     `json:"synced_at"`
}

// Project represents a Specter project.
type Project struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// Proposal represents a document proposal.
type Proposal struct {
	ID            string `json:"id,omitempty"`
	DocumentPath  string `json:"document_path"`
	Branch        string `json:"branch"`
	Type          string `json:"type"`
	AnchorContent string `json:"anchor_content"`
	AnchorLineHint int   `json:"anchor_line_hint,omitempty"`
	Body          string `json:"body"`
	Status        string `json:"status,omitempty"`
}
