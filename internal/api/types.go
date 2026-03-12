package api

import "fmt"

// Document represents a synced markdown document.
type Document struct {
	Path        string `json:"path"`
	ContentMD   string `json:"content_md"`
	ContentHash string `json:"content_hash,omitempty"`
	Revision    int    `json:"revision,omitempty"`
}

// SyncPushRequest is the body for POST /sync/:owner/:slug.
type SyncPushRequest struct {
	Branch       string     `json:"branch"`
	BaseRevision string     `json:"base_revision,omitempty"`
	Documents    []Document `json:"documents"`
}

// SyncPushResponse is the response from POST /sync/:owner/:slug.
type SyncPushResponse struct {
	Created   []string `json:"created"`
	Updated   []string `json:"updated"`
	Unchanged []string `json:"unchanged"`
}

// SyncPullResponse is the response from GET /sync/:owner/:slug.
type SyncPullResponse struct {
	Documents []Document `json:"documents"`
	SyncedAt  string     `json:"synced_at"`
}

// Project represents a Specter project.
type Project struct {
	ID            string `json:"id"`
	Owner         string `json:"owner,omitempty"`
	Slug          string `json:"slug"`
	FullName      string `json:"full_name,omitempty"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch,omitempty"`
	Role          string `json:"role,omitempty"`
}

// Proposal represents a document proposal.
type Proposal struct {
	ID             string `json:"id,omitempty"`
	DocumentPath   string `json:"document_path"`
	Branch         string `json:"branch"`
	Type           string `json:"type"`
	AnchorContent  string `json:"anchor_content"`
	AnchorLineHint int    `json:"anchor_line_hint,omitempty"`
	Body           string `json:"body"`
	Status         string `json:"status,omitempty"`
}

// InviteCode represents a user's invite code.
type InviteCode struct {
	Code     string `json:"code"`
	Redeemed bool   `json:"redeemed"`
}

// User represents the authenticated user's profile.
type User struct {
	ID          string       `json:"id"`
	Email       string       `json:"email"`
	Username    string       `json:"username"`
	InviteCodes []InviteCode `json:"invite_codes"`
}

// ConflictDetail describes a single document conflict from a push.
type ConflictDetail struct {
	Path            string `json:"path"`
	ServerRevision  int    `json:"server_revision"`
	ServerUpdatedAt string `json:"server_updated_at"`
	ServerHash      string `json:"server_hash"`
}

// ConflictError is returned when a push is rejected due to conflicts (HTTP 409).
type ConflictError struct {
	Message   string           `json:"error"`
	Conflicts []ConflictDetail `json:"conflicts"`
	Created   []string         `json:"created"`
	Updated   []string         `json:"updated"`
	Unchanged []string         `json:"unchanged"`
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict: %s (%d conflicting document(s))", e.Message, len(e.Conflicts))
}
