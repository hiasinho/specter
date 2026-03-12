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

// Push bulk-pushes documents to the service.
func (c *Client) Push(project string, req *SyncPushRequest) (*SyncPushResponse, error) {
	resp, err := c.do("POST", fmt.Sprintf("/projects/%s/sync", url.PathEscape(project)), req)
	if err != nil {
		return nil, err
	}
	return decodeResponse[SyncPushResponse](resp)
}

// Pull fetches documents changed since a given revision.
func (c *Client) Pull(project, branch string, sinceRevision *int) (*SyncPullResponse, error) {
	path := fmt.Sprintf("/projects/%s/sync?branch=%s", url.PathEscape(project), url.QueryEscape(branch))
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
	path := fmt.Sprintf("/projects/%s/documents?branch=%s", url.PathEscape(project), url.QueryEscape(branch))
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
	path := fmt.Sprintf("/projects/%s/documents/%s?branch=%s",
		url.PathEscape(project), url.PathEscape(docPath), url.QueryEscape(branch))
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
	path := fmt.Sprintf("/projects/%s/proposals", url.PathEscape(project))
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
	resp, err := c.do("POST", fmt.Sprintf("/projects/%s/proposals", url.PathEscape(project)), proposal)
	if err != nil {
		return nil, err
	}
	return decodeResponse[Proposal](resp)
}

// UpdateProposalStatus accepts or rejects a proposal.
func (c *Client) UpdateProposalStatus(project, id, status string) error {
	body := map[string]string{"status": status}
	resp, err := c.do("PATCH", fmt.Sprintf("/projects/%s/proposals/%s", url.PathEscape(project), url.PathEscape(id)), body)
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
