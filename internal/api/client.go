package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const DefaultBaseURL = "https://yentronrhnmpewiyeqxd.supabase.co/functions/v1"

// Client communicates with the Specter API.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new API client.
func NewClient(token string) *Client {
	return &Client{
		BaseURL:    DefaultBaseURL,
		Token:      token,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) do(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("x-specter-token", c.Token)
	}

	return c.HTTPClient.Do(req)
}

func decodeResponse[T any](resp *http.Response) (*T, error) {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}
	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// Me fetches the authenticated user's profile.
func (c *Client) Me() (*User, error) {
	resp, err := c.do("GET", "/me", nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[User](resp)
}

// ListProjects fetches projects the authenticated user is a member of.
func (c *Client) ListProjects() ([]Project, error) {
	resp, err := c.do("GET", "/projects", nil)
	if err != nil {
		return nil, err
	}
	result, err := decodeResponse[[]Project](resp)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// DeleteProject deletes a project. Owner only.
func (c *Client) DeleteProject(project string) error {
	resp, err := c.do("DELETE", fmt.Sprintf("/projects/%s", project), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// CreateProject creates a new project on the service.
func (c *Client) CreateProject(slug, name string) (*Project, error) {
	body := map[string]string{"slug": slug, "name": name}
	resp, err := c.do("POST", "/projects", body)
	if err != nil {
		return nil, err
	}
	return decodeResponse[Project](resp)
}

// Push bulk-pushes documents to the service.
func (c *Client) Push(project string, req *SyncPushRequest) (*SyncPushResponse, error) {
	resp, err := c.do("POST", fmt.Sprintf("/sync/%s", project), req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusConflict {
		defer resp.Body.Close()
		var conflict ConflictError
		if err := json.NewDecoder(resp.Body).Decode(&conflict); err != nil {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API error (409): %s", string(body))
		}
		return nil, &conflict
	}
	return decodeResponse[SyncPushResponse](resp)
}

// Pull fetches documents changed since a given revision.
func (c *Client) Pull(project, branch string, sinceRevision *int) (*SyncPullResponse, error) {
	path := fmt.Sprintf("/sync/%s?branch=%s", project, url.QueryEscape(branch))
	if sinceRevision != nil {
		path += fmt.Sprintf("&since=%d", *sinceRevision)
	}
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[SyncPullResponse](resp)
}

// ListDocuments fetches all documents for a project and branch.
func (c *Client) ListDocuments(project, branch string) ([]Document, error) {
	path := fmt.Sprintf("/documents/%s?branch=%s", project, url.QueryEscape(branch))
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	result, err := decodeResponse[[]Document](resp)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// GetDocument fetches a single document.
func (c *Client) GetDocument(project, branch, docPath string) (*Document, error) {
	path := fmt.Sprintf("/documents/%s/%s?branch=%s",
		project, docPath, url.QueryEscape(branch))
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[Document](resp)
}

// ListProposals fetches proposals for a project.
func (c *Client) ListProposals(project, document, status string) ([]Proposal, error) {
	params := url.Values{}
	if document != "" {
		params.Set("document", document)
	}
	if status != "" {
		params.Set("status", status)
	}
	path := fmt.Sprintf("/proposals/%s", project)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	result, err := decodeResponse[[]Proposal](resp)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// CreateProposal creates a new proposal.
func (c *Client) CreateProposal(project string, proposal *Proposal) (*Proposal, error) {
	resp, err := c.do("POST", fmt.Sprintf("/proposals/%s", project), proposal)
	if err != nil {
		return nil, err
	}
	return decodeResponse[Proposal](resp)
}

// CreateInvite creates an invite for a project role. Owner only.
func (c *Client) CreateInvite(project, role string) (*Invite, error) {
	body := map[string]string{"project": project, "role": role}
	resp, err := c.do("POST", "/invites", body)
	if err != nil {
		return nil, err
	}
	return decodeResponse[Invite](resp)
}

// RedeemInvite redeems an invite code, adding the user as a project member.
func (c *Client) RedeemInvite(code string) (*InviteRedeemResponse, error) {
	body := map[string]string{"code": code}
	resp, err := c.do("POST", "/invites/redeem", body)
	if err != nil {
		return nil, err
	}
	return decodeResponse[InviteRedeemResponse](resp)
}

// ListDocumentHistory fetches revision history for a document.
func (c *Client) ListDocumentHistory(project, docPath, branch string, limit int) (*DocumentHistory, error) {
	params := url.Values{}
	if branch != "" {
		params.Set("branch", branch)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	path := fmt.Sprintf("/document-history/%s/%s/history", project, docPath)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[DocumentHistory](resp)
}

// GetDocumentRevision fetches a specific revision's full content.
func (c *Client) GetDocumentRevision(project, docPath string, revision int) (*DocumentRevision, error) {
	path := fmt.Sprintf("/document-history/%s/%s/history/%d", project, docPath, revision)
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[DocumentRevision](resp)
}

// GetDocumentDiff compares two revisions of a document.
func (c *Client) GetDocumentDiff(project, docPath, branch string, from, to int) (*DocumentDiff, error) {
	params := url.Values{}
	if branch != "" {
		params.Set("branch", branch)
	}
	params.Set("from", fmt.Sprintf("%d", from))
	if to > 0 {
		params.Set("to", fmt.Sprintf("%d", to))
	}
	path := fmt.Sprintf("/document-history/%s/%s/diff?%s", project, docPath, params.Encode())
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[DocumentDiff](resp)
}

// UpdateProposalStatus accepts or rejects a proposal.
func (c *Client) UpdateProposalStatus(project, id, status string) error {
	body := map[string]string{"status": status}
	resp, err := c.do("PATCH", fmt.Sprintf("/proposals/%s/%s", project, id), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}
