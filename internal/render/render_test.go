package render

import (
	"strings"
	"testing"
	"time"

	"github.com/OWNER/jules-tui/internal/julesapi"
)

func TestPlanRendering(t *testing.T) {
	plan := &julesapi.Plan{
		Steps: []julesapi.PlanStep{
			{Index: 0, Title: "Step One", Description: "Details for step 1"},
			{Index: 1, Title: "Step Two", Description: "Details for step 2"},
		},
	}

	out := Plan(plan)
	if !strings.Contains(out, "[1] Step One") || !strings.Contains(out, "Details for step 1") {
		t.Errorf("unexpected plan output: %s", out)
	}
	if !strings.Contains(out, "[2] Step Two") {
		t.Errorf("unexpected step 2 in plan output: %s", out)
	}
}

func TestDiffRendering(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	diffContent := "--- a/file.go\n+++ b/file.go\n@@ -1,2 +1,3 @@\n-old line\n+new line\n context line"
	out := Diff(diffContent)
	if !strings.Contains(out, "+new line") || !strings.Contains(out, "-old line") {
		t.Errorf("expected diff lines, got: %s", out)
	}
}

func TestActivityRenderingVariants(t *testing.T) {
	now := time.Now()

	// 1. Plan generated
	actPlan := julesapi.Activity{
		ID:         "act-1",
		Originator: "agent",
		CreateTime: &now,
		PlanGenerated: &julesapi.PlanGenerated{
			Plan: &julesapi.Plan{
				Steps: []julesapi.PlanStep{{Index: 0, Title: "Analyze"}},
			},
		},
	}
	out := Activity(actPlan, false, false, 80)
	if !strings.Contains(out, "Plan Generated") || !strings.Contains(out, "Analyze") {
		t.Errorf("expected plan generated in activity output: %s", out)
	}

	// 2. Agent message with markdown
	actMsg := julesapi.Activity{
		ID:         "act-2",
		Originator: "agent",
		AgentMessaged: &julesapi.AgentMessaged{
			AgentMessage: "Here is **bold text** and `code`",
		},
	}
	outMsg := Activity(actMsg, false, false, 80)
	if !strings.Contains(outMsg, "Agent Message") {
		t.Errorf("expected Agent Message in output: %s", outMsg)
	}

	// 3. Artifact diff collapsed vs expanded
	actArtifact := julesapi.Activity{
		ID:         "act-3",
		Originator: "agent",
		Artifacts: []julesapi.Artifact{
			{
				ChangeSet: &julesapi.ChangeSet{
					GitPatch: &julesapi.GitPatch{
						UnidiffPatch: "+line 1\n+line 2\n",
					},
				},
			},
		},
	}
	outCollapsed := Activity(actArtifact, false, false, 80)
	if !strings.Contains(outCollapsed, "press 'd' to expand") {
		t.Errorf("expected collapsed diff prompt, got: %s", outCollapsed)
	}
	outExpanded := Activity(actArtifact, true, false, 80)
	if !strings.Contains(outExpanded, "+line 1") {
		t.Errorf("expected expanded diff content, got: %s", outExpanded)
	}

	// 4. Unknown activity raw json fallback
	actUnknown := julesapi.Activity{
		ID:      "act-4",
		RawJSON: `{"customEvent":"surprise"}`,
	}
	outRaw := Activity(actUnknown, false, true, 80)
	if !strings.Contains(outRaw, "surprise") {
		t.Errorf("expected raw json output, got: %s", outRaw)
	}
}
