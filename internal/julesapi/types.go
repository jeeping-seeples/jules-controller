package julesapi

import "time"

// SessionState represents the state of a Jules session.
type SessionState string

const (
	StateUnspecified          SessionState = "STATE_UNSPECIFIED"
	StateQueued               SessionState = "QUEUED"
	StatePlanning             SessionState = "PLANNING"
	StateAwaitingPlanApproval SessionState = "AWAITING_PLAN_APPROVAL"
	StateAwaitingUserFeedback SessionState = "AWAITING_USER_FEEDBACK"
	StateInProgress           SessionState = "IN_PROGRESS"
	StatePaused               SessionState = "PAUSED"
	StateCompleted            SessionState = "COMPLETED"
	StateFailed               SessionState = "FAILED"
)

// AutomationMode represents automation options for a session.
type AutomationMode string

const (
	AutomationModeUnspecified AutomationMode = "AUTOMATION_MODE_UNSPECIFIED"
	AutoCreatePR              AutomationMode = "AUTO_CREATE_PR"
)

// GitHubRepo contains details about a GitHub repository.
type GitHubRepo struct {
	Owner         string         `json:"owner,omitempty"`
	Repo          string         `json:"repo,omitempty"`
	IsPrivate     bool           `json:"isPrivate,omitempty"`
	DefaultBranch *GitHubBranch  `json:"defaultBranch,omitempty"`
	Branches      []GitHubBranch `json:"branches,omitempty"`
}

// GitHubBranch represents a GitHub branch.
type GitHubBranch struct {
	DisplayName string `json:"displayName,omitempty"`
}

// GitHubRepoContext contains context for using a GitHub repo.
type GitHubRepoContext struct {
	StartingBranch string `json:"startingBranch"`
}

// SourceContext defines the source repository and branch context.
type SourceContext struct {
	Source            string             `json:"source"`
	GitHubRepoContext *GitHubRepoContext `json:"githubRepoContext,omitempty"`
}

// PullRequest represents a pull request produced by a session.
type PullRequest struct {
	URL         string `json:"url,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// SessionOutput represents an output from a completed session.
type SessionOutput struct {
	PullRequest *PullRequest `json:"pullRequest,omitempty"`
}

// Session represents a Google Jules session resource.
type Session struct {
	Name                string          `json:"name,omitempty"`
	ID                  string          `json:"id,omitempty"`
	Prompt              string          `json:"prompt,omitempty"`
	Title               string          `json:"title,omitempty"`
	State               SessionState    `json:"state,omitempty"`
	URL                 string          `json:"url,omitempty"`
	SourceContext       *SourceContext  `json:"sourceContext,omitempty"`
	RequirePlanApproval bool            `json:"requirePlanApproval,omitempty"`
	AutomationMode      AutomationMode  `json:"automationMode,omitempty"`
	Outputs             []SessionOutput `json:"outputs,omitempty"`
	CreateTime          *time.Time      `json:"createTime,omitempty"`
	UpdateTime          *time.Time      `json:"updateTime,omitempty"`
}

// ListSessionsResponse represents the response from GET /v1alpha/sessions.
type ListSessionsResponse struct {
	Sessions      []Session `json:"sessions,omitempty"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
}

// CreateSessionRequest represents POST /v1alpha/sessions request body.
type CreateSessionRequest struct {
	Prompt              string         `json:"prompt"`
	Title               string         `json:"title,omitempty"`
	SourceContext       *SourceContext `json:"sourceContext,omitempty"`
	RequirePlanApproval bool           `json:"requirePlanApproval,omitempty"`
	AutomationMode      AutomationMode `json:"automationMode,omitempty"`
}

// SendMessageRequest represents POST /v1alpha/sessions/{id}:sendMessage request body.
type SendMessageRequest struct {
	Prompt string `json:"prompt"`
}

// SendMessageResponse represents POST /v1alpha/sessions/{id}:sendMessage response body.
type SendMessageResponse struct{}

// ApprovePlanRequest represents POST /v1alpha/sessions/{id}:approvePlan request body.
type ApprovePlanRequest struct{}

// ApprovePlanResponse represents POST /v1alpha/sessions/{id}:approvePlan response body.
type ApprovePlanResponse struct{}

// Source represents a connected code source.
type Source struct {
	Name       string      `json:"name,omitempty"`
	ID         string      `json:"id,omitempty"`
	GitHubRepo *GitHubRepo `json:"githubRepo,omitempty"`
}

// ListSourcesResponse represents GET /v1alpha/sources response body.
type ListSourcesResponse struct {
	Sources       []Source `json:"sources,omitempty"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
}
