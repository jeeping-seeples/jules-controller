package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/OWNER/jules-tui/internal/julesapi"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Markdown renders markdown content using Glamour or plain text if NO_COLOR is set.
func Markdown(content string, wordWrap int) string {
	if wordWrap <= 0 {
		wordWrap = 80
	}

	if os.Getenv("NO_COLOR") != "" {
		return content
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(wordWrap),
	)
	if err != nil {
		return content
	}

	out, err := r.Render(content)
	if err != nil {
		return content
	}
	return strings.TrimRight(out, "\n")
}

// Plan renders plan steps in structured format.
func Plan(plan *julesapi.Plan) string {
	if plan == nil || len(plan.Steps) == 0 {
		return "  (No plan steps recorded)"
	}

	var b strings.Builder
	for i, step := range plan.Steps {
		stepNum := step.Index + 1
		if stepNum <= 0 {
			stepNum = i + 1
		}
		title := step.Title
		if title == "" {
			title = "(untitled step)"
		}
		b.WriteString(fmt.Sprintf("  [%d] %s\n", stepNum, title))
		if step.Description != "" {
			lines := strings.Split(step.Description, "\n")
			for _, line := range lines {
				if line != "" {
					b.WriteString(fmt.Sprintf("      %s\n", line))
				}
			}
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// Diff renders a unified diff with coloring for additions and deletions.
func Diff(unidiff string) string {
	if unidiff == "" {
		return "  (Empty diff)"
	}

	noColor := os.Getenv("NO_COLOR") != ""
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379"))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75"))
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("#61AFEF"))

	var b strings.Builder
	lines := strings.Split(unidiff, "\n")
	for _, line := range lines {
		if noColor {
			b.WriteString(line + "\n")
			continue
		}

		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			b.WriteString(green.Render(line) + "\n")
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			b.WriteString(red.Render(line) + "\n")
		} else if strings.HasPrefix(line, "@@") {
			b.WriteString(cyan.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// Activity renders a single Jules activity into formatted human readable text.
func Activity(act julesapi.Activity, showDiffs bool, showRawJSON bool, wrapWidth int) string {
	if showRawJSON && act.RawJSON != "" {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, []byte(act.RawJSON), "  ", "  "); err == nil {
			return fmt.Sprintf("  Activity %s (Raw JSON):\n  %s", act.ID, pretty.String())
		}
		return fmt.Sprintf("  Activity %s (Raw JSON):\n  %s", act.ID, act.RawJSON)
	}

	var b strings.Builder
	timestamp := "-"
	if act.CreateTime != nil {
		timestamp = act.CreateTime.Format("15:04:05")
	}

	originator := act.Originator
	if originator == "" {
		originator = "system"
	}

	b.WriteString(fmt.Sprintf("[%s] %s (%s):\n", timestamp, strings.ToUpper(originator), act.ID))

	handled := false

	if act.PlanGenerated != nil && act.PlanGenerated.Plan != nil {
		handled = true
		b.WriteString("  Plan Generated:\n")
		b.WriteString(Plan(act.PlanGenerated.Plan))
		b.WriteString("\n")
	}

	if act.PlanApproved != nil {
		handled = true
		planID := act.PlanApproved.PlanID
		if planID == "" {
			planID = "approved"
		}
		b.WriteString(fmt.Sprintf("  Plan %s was approved.\n", planID))
	}

	if act.AgentMessaged != nil {
		handled = true
		rendered := Markdown(act.AgentMessaged.AgentMessage, wrapWidth)
		b.WriteString("  Agent Message:\n")
		for _, line := range strings.Split(rendered, "\n") {
			b.WriteString("    " + line + "\n")
		}
	}

	if act.UserMessaged != nil {
		handled = true
		b.WriteString(fmt.Sprintf("  User: %s\n", act.UserMessaged.UserMessage))
	}

	if act.ProgressUpdated != nil {
		handled = true
		b.WriteString(fmt.Sprintf("  Progress: %s\n", act.ProgressUpdated.Title))
		if act.ProgressUpdated.Description != "" {
			b.WriteString(fmt.Sprintf("    %s\n", act.ProgressUpdated.Description))
		}
	}

	if act.SessionCompleted != nil {
		handled = true
		b.WriteString("  Session completed successfully.\n")
	}

	if act.SessionFailed != nil {
		handled = true
		b.WriteString(fmt.Sprintf("  Session failed: %s\n", act.SessionFailed.Reason))
	}

	// Artifacts
	for i, artifact := range act.Artifacts {
		if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
			patch := artifact.ChangeSet.GitPatch
			b.WriteString(fmt.Sprintf("  Artifact #%d: Code Changes\n", i+1))
			if patch.SuggestedCommitMessage != "" {
				b.WriteString(fmt.Sprintf("    Suggested commit: %s\n", patch.SuggestedCommitMessage))
			}
			if patch.BaseCommitID != "" {
				b.WriteString(fmt.Sprintf("    Base commit: %s\n", patch.BaseCommitID))
			}
			if showDiffs && patch.UnidiffPatch != "" {
				b.WriteString("    Diff:\n")
				renderedDiff := Diff(patch.UnidiffPatch)
				for _, line := range strings.Split(renderedDiff, "\n") {
					b.WriteString("      " + line + "\n")
				}
			} else if patch.UnidiffPatch != "" {
				lines := strings.Count(patch.UnidiffPatch, "\n")
				b.WriteString(fmt.Sprintf("    Diff: %d lines (press 'd' to expand)\n", lines))
			}
		}

		if artifact.BashOutput != nil {
			b.WriteString(fmt.Sprintf("  Artifact #%d: Bash Output ($ %s) [exit: %d]\n", i+1, artifact.BashOutput.Command, artifact.BashOutput.ExitCode))
			if artifact.BashOutput.Output != "" {
				for _, line := range strings.Split(artifact.BashOutput.Output, "\n") {
					b.WriteString("    " + line + "\n")
				}
			}
		}

		if artifact.Media != nil {
			b.WriteString(fmt.Sprintf("  Artifact #%d: Media (%s, %d bytes)\n", i+1, artifact.Media.MimeType, len(artifact.Media.Data)))
		}
	}

	// If no known event field was populated, output description or raw JSON fallback
	if !handled && len(act.Artifacts) == 0 {
		if act.Description != "" {
			b.WriteString(fmt.Sprintf("  %s\n", act.Description))
		}
		if act.RawJSON != "" {
			b.WriteString(fmt.Sprintf("  Raw: %s\n", act.RawJSON))
		}
	}

	return strings.TrimRight(b.String(), "\n")
}
