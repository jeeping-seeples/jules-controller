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

// PlanStep represents a step in a plan.
type PlanStep struct {
	ID          string `json:"id,omitempty"`
	Index       int    `json:"index,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// Plan represents a sequence of steps.
type Plan struct {
	ID         string     `json:"id,omitempty"`
	Steps      []PlanStep `json:"steps,omitempty"`
	CreateTime *time.Time `json:"createTime,omitempty"`
}

// PlanGenerated represents a plan generation event.
type PlanGenerated struct {
	Plan *Plan `json:"plan,omitempty"`
}

// PlanApproved represents a plan approval event.
type PlanApproved struct {
	PlanID string `json:"planId,omitempty"`
}

// UserMessaged represents a message from the user.
type UserMessaged struct {
	UserMessage string `json:"userMessage,omitempty"`
}

// AgentMessaged represents a message from Jules.
type AgentMessaged struct {
	AgentMessage string `json:"agentMessage,omitempty"`
}

// ProgressUpdated represents a progress status update.
type ProgressUpdated struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// SessionCompleted represents session completion.
type SessionCompleted struct{}

// SessionFailed represents session failure.
type SessionFailed struct {
	Reason string `json:"reason,omitempty"`
}

// GitPatch represents git diff patch details.
type GitPatch struct {
	BaseCommitID           string `json:"baseCommitId,omitempty"`
	UnidiffPatch           string `json:"unidiffPatch,omitempty"`
	SuggestedCommitMessage string `json:"suggestedCommitMessage,omitempty"`
}

// ChangeSet represents a change set artifact.
type ChangeSet struct {
	Source   string    `json:"source,omitempty"`
	GitPatch *GitPatch `json:"gitPatch,omitempty"`
}

// BashOutput represents bash command output artifact.
type BashOutput struct {
	Command  string `json:"command,omitempty"`
	Output   string `json:"output,omitempty"`
	ExitCode int    `json:"exitCode,omitempty"`
}

// Media represents a media file artifact.
type Media struct {
	MimeType string `json:"mimeType,omitempty"`
	Data     string `json:"data,omitempty"`
}

// Artifact represents an output artifact produced by an activity.
type Artifact struct {
	ChangeSet  *ChangeSet  `json:"changeSet,omitempty"`
	BashOutput *BashOutput `json:"bashOutput,omitempty"`
	Media      *Media      `json:"media,omitempty"`
}

// Activity represents an event within a session.
type Activity struct {
	Name             string            `json:"name,omitempty"`
	ID               string            `json:"id,omitempty"`
	Originator       string            `json:"originator,omitempty"`
	Description      string            `json:"description,omitempty"`
	CreateTime       *time.Time        `json:"createTime,omitempty"`
	Artifacts        []Artifact        `json:"artifacts,omitempty"`
	PlanGenerated    *PlanGenerated    `json:"planGenerated,omitempty"`
	PlanApproved     *PlanApproved     `json:"planApproved,omitempty"`
	UserMessaged     *UserMessaged     `json:"userMessaged,omitempty"`
	AgentMessaged    *AgentMessaged    `json:"agentMessaged,omitempty"`
	ProgressUpdated  *ProgressUpdated  `json:"progressUpdated,omitempty"`
	SessionCompleted *SessionCompleted `json:"sessionCompleted,omitempty"`
	SessionFailed    *SessionFailed    `json:"sessionFailed,omitempty"`

	// RawJSON stores raw activity JSON for unknown activity types
	RawJSON string `json:"-"`
}

// ListActivitiesResponse represents the response from GET /v1alpha/sessions/{id}/activities.
type ListActivitiesResponse struct {
	Activities    []Activity `json:"activities,omitempty"`
	NextPageToken string     `json:"nextPageToken,omitempty"`
}
