package julesapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is a client for the Google Jules REST API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	maxRetries int
	baseDelay  time.Duration
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithMaxRetries sets the maximum number of retries for 429 and 5xx responses.
func WithMaxRetries(retries int) Option {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

// WithBaseDelay sets the initial backoff delay.
func WithBaseDelay(delay time.Duration) Option {
	return func(c *Client) {
		c.baseDelay = delay
	}
}

// NewClient creates a new Client.
func NewClient(baseURL, apiKey string, opts ...Option) *Client {
	c := &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		maxRetries: 3,
		baseDelay:  500 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// doRequest performs an HTTP request with retries on 429 and 5xx responses.
func (c *Client) doRequest(ctx context.Context, method, endpoint string, query url.Values, body any, target any) error {
	if c.apiKey == "" {
		return &APIError{
			StatusCode: 401,
			Message:    "Jules API key is required. Set JULES_API_KEY environment variable, configure it in config.toml, or pass --api-key flag.",
			ExitCode:   ExitAuthError,
		}
	}

	fullURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to encode request body: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}

		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}

		// Security: send API key strictly via header, never query parameter
		req.Header.Set("x-goog-api-key", c.apiKey)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			// Network error, back off and retry
			if attempt < c.maxRetries {
				time.Sleep(c.backoffDuration(attempt, 0))
				continue
			}
			return &APIError{
				StatusCode: 0,
				Message:    fmt.Sprintf("network error: %v", err),
				ExitCode:   ExitAPIError,
			}
		}

		respBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			if attempt < c.maxRetries {
				time.Sleep(c.backoffDuration(attempt, 0))
				continue
			}
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("failed to read response body: %v", err),
				ExitCode:   ExitAPIError,
			}
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if target != nil && len(respBody) > 0 {
				if err := json.Unmarshal(respBody, target); err != nil {
					return fmt.Errorf("failed to decode response: %w", err)
				}
			}
			return nil
		}

		// Check for retryable status codes (429 or 5xx)
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			var retryAfter time.Duration
			if raStr := resp.Header.Get("Retry-After"); raStr != "" {
				if seconds, parseErr := strconv.Atoi(raStr); parseErr == nil && seconds > 0 {
					retryAfter = time.Duration(seconds) * time.Second
				}
			}

			lastErr = NewAPIError(resp.StatusCode, extractErrorMessage(respBody))
			if attempt < c.maxRetries {
				time.Sleep(c.backoffDuration(attempt, retryAfter))
				continue
			}
			return lastErr
		}

		// Non-retryable error (e.g. 400, 401, 403, 404)
		return NewAPIError(resp.StatusCode, extractErrorMessage(respBody))
	}

	if lastErr != nil {
		return lastErr
	}
	return errors.New("request failed after retries")
}

func (c *Client) backoffDuration(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		return retryAfter
	}
	delay := c.baseDelay * (1 << attempt)
	return delay
}

func extractErrorMessage(respBody []byte) string {
	if len(respBody) == 0 {
		return "unknown error"
	}
	// Try parsing Google API standard error JSON: {"error": {"message": "..."}}
	var errResp struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(respBody, &errResp); err == nil {
		if errResp.Error.Message != "" {
			return errResp.Error.Message
		}
		if errResp.Message != "" {
			return errResp.Message
		}
	}
	return string(respBody)
}

// ListSessions lists sessions with optional pageSize, pageToken, and state filter.
func (c *Client) ListSessions(ctx context.Context, pageSize int, pageToken string, stateFilter SessionState) (*ListSessionsResponse, error) {
	q := url.Values{}
	if pageSize > 0 {
		q.Set("pageSize", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		q.Set("pageToken", pageToken)
	}

	var resp ListSessionsResponse
	if err := c.doRequest(ctx, http.MethodGet, "/v1alpha/sessions", q, nil, &resp); err != nil {
		return nil, err
	}

	if stateFilter != "" {
		filtered := make([]Session, 0, len(resp.Sessions))
		for _, s := range resp.Sessions {
			if strings.EqualFold(string(s.State), string(stateFilter)) {
				filtered = append(filtered, s)
			}
		}
		resp.Sessions = filtered
	}

	return &resp, nil
}

// GetSession retrieves a single session by ID or resource name.
func (c *Client) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	sessionID = normalizeResourceID("sessions", sessionID)
	endpoint := fmt.Sprintf("/v1alpha/sessions/%s", sessionID)
	var session Session
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, nil, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// CreateSession creates a new session.
func (c *Client) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {
	var session Session
	if err := c.doRequest(ctx, http.MethodPost, "/v1alpha/sessions", nil, req, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// SendMessage sends a user message to an active session.
func (c *Client) SendMessage(ctx context.Context, sessionID, message string) error {
	sessionID = normalizeResourceID("sessions", sessionID)
	endpoint := fmt.Sprintf("/v1alpha/sessions/%s:sendMessage", sessionID)
	req := SendMessageRequest{Prompt: message}
	var resp SendMessageResponse
	return c.doRequest(ctx, http.MethodPost, endpoint, nil, req, &resp)
}

// ApprovePlan approves a pending plan in a session.
func (c *Client) ApprovePlan(ctx context.Context, sessionID string) error {
	sessionID = normalizeResourceID("sessions", sessionID)
	endpoint := fmt.Sprintf("/v1alpha/sessions/%s:approvePlan", sessionID)
	req := ApprovePlanRequest{}
	var resp ApprovePlanResponse
	return c.doRequest(ctx, http.MethodPost, endpoint, nil, req, &resp)
}

// ListSources lists connected sources with pagination.
func (c *Client) ListSources(ctx context.Context, pageSize int, pageToken string) (*ListSourcesResponse, error) {
	q := url.Values{}
	if pageSize > 0 {
		q.Set("pageSize", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		q.Set("pageToken", pageToken)
	}

	var resp ListSourcesResponse
	if err := c.doRequest(ctx, http.MethodGet, "/v1alpha/sources", q, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ResolveSource resolves a repository reference into a source resource name.
// Strategy:
// 1. If repo starts with "sources/", pass through as raw source name.
// 2. Otherwise list sources (paginated) and match githubRepo owner and repo (case-insensitive).
// 3. If no match is found, return error indicating repository is not connected.
// 4. If listing sources fails, fallback to derived "sources/github-<owner>-<repo>" with a warning.
func (c *Client) ResolveSource(ctx context.Context, repo string, warnFn func(format string, args ...any)) (string, error) {
	if strings.HasPrefix(repo, "sources/") {
		return repo, nil
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid repository format %q; expected 'owner/name' or 'sources/...'", repo)
	}
	owner, name := parts[0], parts[1]

	// Paginate through sources to find match
	var pageToken string
	for {
		resp, err := c.ListSources(ctx, 100, pageToken)
		if err != nil {
			fallback := fmt.Sprintf("sources/github-%s-%s", owner, name)
			if warnFn != nil {
				warnFn("Warning: failed to query sources list (%v); falling back to derived source name %s\n", err, fallback)
			}
			return fallback, nil
		}

		for _, src := range resp.Sources {
			if src.GitHubRepo != nil {
				if strings.EqualFold(src.GitHubRepo.Owner, owner) && strings.EqualFold(src.GitHubRepo.Repo, name) {
					return src.Name, nil
				}
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return "", fmt.Errorf("repository %q is not connected to Jules (run Jules web UI to connect it or pass raw source resource name)", repo)
}

func normalizeResourceID(prefix, id string) string {
	id = strings.TrimPrefix(id, prefix+"/")
	return strings.TrimPrefix(id, "/")
}

// ListActivities lists activities for a session with optional pagination and createTime filter.
func (c *Client) ListActivities(ctx context.Context, sessionID string, pageSize int, pageToken string, createTime string) (*ListActivitiesResponse, error) {
	sessionID = normalizeResourceID("sessions", sessionID)
	endpoint := fmt.Sprintf("/v1alpha/sessions/%s/activities", sessionID)

	q := url.Values{}
	if pageSize > 0 {
		q.Set("pageSize", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		q.Set("pageToken", pageToken)
	}
	if createTime != "" {
		q.Set("createTime", createTime)
	}

	var rawResp struct {
		Activities    []json.RawMessage `json:"activities"`
		NextPageToken string            `json:"nextPageToken"`
	}

	if err := c.doRequest(ctx, http.MethodGet, endpoint, q, nil, &rawResp); err != nil {
		return nil, err
	}

	resp := &ListActivitiesResponse{
		NextPageToken: rawResp.NextPageToken,
		Activities:    make([]Activity, 0, len(rawResp.Activities)),
	}

	for _, rawItem := range rawResp.Activities {
		var act Activity
		if err := json.Unmarshal(rawItem, &act); err != nil {
			// Tolerant decoding: keep raw JSON even if parsing some fields fails
			act = Activity{
				ID:          "unknown",
				Description: "Unable to parse activity",
			}
		}
		act.RawJSON = string(rawItem)
		resp.Activities = append(resp.Activities, act)
	}

	return resp, nil
}
