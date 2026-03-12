package api

import "fmt"

// Document represents a synced markdown document.
type Document struct {
	Path        string `json:"path"`
	ContentMD   string `json:"content_md"`
	ContentHash string `json:"content_hash,omitempty"`
	Revision    int    `json:"revision,omitempty"`
}

// PushDocument contains only the fields sent when pushing documents to the API.
type PushDocument struct {
	Path      string `json:"path"`
	ContentMD string `json:"content_md"`
}

// SyncPushRequest is the body for POST /sync/:owner/:slug.
type SyncPushRequest struct {
	Branch       string         `json:"branch"`
	BaseRevision string         `json:"base_revision,omitempty"`
	Documents    []PushDocument `json:"documents"`
}

// SyncPushResponse is the response from POST /sync/:owner/:slug.
type SyncPushResponse struct {
	Created   []string `json:"created"`
	Updated   []string `json:"updated"`
	Unchanged []string `json:"unchanged"`
	SyncedAt  string   `json:"synced_at"`
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

// Invite represents the response from creating an invite.
type Invite struct {
	Code string `json:"code"`
	Role string `json:"role"`
}

// InviteRedeemResponse represents the response from redeeming an invite.
type InviteRedeemResponse struct {
	Project Project `json:"project"`
	Role    string  `json:"role"`
}

// Author represents a user in history responses.
type Author struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// RevisionMeta is a single entry in a document's revision history.
type RevisionMeta struct {
	Revision    int    `json:"revision"`
	ContentHash string `json:"content_hash"`
	Author      Author `json:"author"`
	CreatedAt   string `json:"created_at"`
}

// DocumentHistory is the response from listing a document's revision history.
type DocumentHistory struct {
	Path      string         `json:"path"`
	Revisions []RevisionMeta `json:"revisions"`
}

// DocumentRevision is the response from fetching a specific revision.
type DocumentRevision struct {
	Path        string `json:"path"`
	Revision    int    `json:"revision"`
	ContentMD   string `json:"content_md"`
	ContentHash string `json:"content_hash"`
	Author      Author `json:"author"`
	CreatedAt   string `json:"created_at"`
}

// DocumentDiff is the response from comparing two revisions.
type DocumentDiff struct {
	Path         string `json:"path"`
	FromRevision int    `json:"from_revision"`
	ToRevision   int    `json:"to_revision"`
	Diff         string `json:"diff"`
}

// RegisterResponse is the response from POST /register.
type RegisterResponse struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Username    string   `json:"username"`
	Token       string   `json:"token"`
	InviteCodes []string `json:"invite_codes"`
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
