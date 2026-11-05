package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/OWNER/jules-tui/internal/config"
	"github.com/OWNER/jules-tui/internal/julesapi"
	"github.com/OWNER/jules-tui/internal/notify"
	"github.com/OWNER/jules-tui/internal/poller"
	"github.com/OWNER/jules-tui/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	exitCode := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	os.Exit(exitCode)
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return runTUI("", "", stderr)
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "tui":
		return runTUIWithArgs(subArgs, stderr)
	case "sessions":
		return runSessions(subArgs, stdout, stderr)
	case "show":
		return runShow(subArgs, stdout, stderr)
	case "send":
		return runSend(subArgs, stdin, stdout, stderr)
	case "approve":
		return runApprove(subArgs, stdout, stderr)
	case "new":
		return runNew(subArgs, stdin, stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return julesapi.ExitSuccess
	default:
		fmt.Fprintf(stderr, "Unknown command: %s\n\n", subcommand)
		printUsage(stderr)
		return julesapi.ExitUsageError
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `jules-tui - CLI client for Google Jules

Usage:
  jules-tui [command] [arguments]

If no command is provided, the interactive TUI home screen launches.

Commands:
  tui               Launch interactive terminal UI
  sessions [--json] [--state STATE] [--page-size N]
      List sessions for the authenticated user

  show <session-id> [--json]
      Show metadata and outputs of a session

  send <session-id> (--message-file <path|-> | --message "text")
      Send a message to an active session

  approve <session-id>
      Approve a pending plan for a session

  new --repo <owner/name> --branch <branch> (--prompt-file <path|-> | --prompt "text") [--require-approval] [--title "title"] [--auto-pr] [--json]
      Create a new session

Global flags (applicable per command):
  --config <path>   Path to configuration TOML file
  --api-key <key>   Google Jules API key override

Exit codes:
  0 Success
  1 General error
  2 Usage error
  3 Auth error (401/403 or missing API key)
  4 Resource not found (404)
  5 API / Network error
`)
}

// splitPositionalAndFlags reorganizes args so flags come before positional arguments.
func splitPositionalAndFlags(args []string) ([]string, []string) {
	var flags []string
	var positionals []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flagName := strings.TrimLeft(arg, "-")
				switch flagName {
				case "json", "require-approval", "auto-pr", "h", "help":
					// boolean
				default:
					flags = append(flags, args[i+1])
					i++
				}
			}
		} else {
			positionals = append(positionals, arg)
		}
	}
	return flags, positionals
}

func createClient(configPath, flagAPIKey string) (*julesapi.Client, error) {
	cfg, err := config.Load(configPath, flagAPIKey)
	if err != nil {
		return nil, err
	}
	if cfg.APIKey == "" {
		return nil, &julesapi.APIError{
			StatusCode: 401,
			Message:    "API key is missing. Set JULES_API_KEY environment variable, configure it in config.toml, or pass --api-key.",
			ExitCode:   julesapi.ExitAuthError,
		}
	}
	return julesapi.NewClient(cfg.BaseURL, cfg.APIKey), nil
}

func handleError(err error, stderr io.Writer) int {
	var apiErr *julesapi.APIError
	if errors.As(err, &apiErr) {
		fmt.Fprintf(stderr, "Error: %s\n", apiErr.Message)
		if apiErr.ExitCode != 0 {
			return apiErr.ExitCode
		}
		return julesapi.ExitAPIError
	}
	fmt.Fprintf(stderr, "Error: %v\n", err)
	return julesapi.ExitGeneralError
}

func runTUIWithArgs(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var configPath string
	var apiKey string

	fs.StringVar(&configPath, "config", "", "Path to config TOML file")
	fs.StringVar(&apiKey, "api-key", "", "API key override")

	flags, _ := splitPositionalAndFlags(args)
	if err := fs.Parse(flags); err != nil {
		return julesapi.ExitUsageError
	}

	return runTUI(configPath, apiKey, stderr)
}

func runTUI(configPath, apiKey string, stderr io.Writer) int {
	client, err := createClient(configPath, apiKey)
	if err != nil {
		return handleError(err, stderr)
	}

	store, err := notify.NewStore("")
	if err != nil {
		return handleError(err, stderr)
	}
	notifier := notify.NewNotifier(store)
	p := poller.New(client, notifier, poller.DefaultConfig())

	model := tui.NewModel(p)
	program := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(stderr, "Error running TUI: %v\n", err)
		return julesapi.ExitGeneralError
	}
	return julesapi.ExitSuccess
}

func runSessions(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("sessions", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var jsonOutput bool
	var stateStr string
	var pageSize int
	var configPath string
	var apiKey string

	fs.BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	fs.StringVar(&stateStr, "state", "", "Filter by session state (e.g. QUEUED, PLANNING, AWAITING_PLAN_APPROVAL, COMPLETED)")
	fs.IntVar(&pageSize, "page-size", 30, "Number of sessions to return (1-100)")
	fs.StringVar(&configPath, "config", "", "Path to config TOML file")
	fs.StringVar(&apiKey, "api-key", "", "API key override")

	flags, _ := splitPositionalAndFlags(args)
	if err := fs.Parse(flags); err != nil {
		return julesapi.ExitUsageError
	}

	client, err := createClient(configPath, apiKey)
	if err != nil {
		return handleError(err, stderr)
	}

	resp, err := client.ListSessions(context.Background(), pageSize, "", julesapi.SessionState(stateStr))
	if err != nil {
		return handleError(err, stderr)
	}

	if jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(resp); err != nil {
			return handleError(err, stderr)
		}
		return julesapi.ExitSuccess
	}

	if len(resp.Sessions) == 0 {
		fmt.Fprintln(stdout, "No sessions found.")
		return julesapi.ExitSuccess
	}

	w := tabwriter.NewWriter(stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATE\tTITLE\tUPDATED")
	for _, s := range resp.Sessions {
		id := s.ID
		if id == "" {
			id = strings.TrimPrefix(s.Name, "sessions/")
		}
		title := s.Title
		if title == "" {
			title = "(untitled)"
		}
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		updated := "-"
		if s.UpdateTime != nil {
			updated = s.UpdateTime.Format(time.RFC3339)
		} else if s.CreateTime != nil {
			updated = s.CreateTime.Format(time.RFC3339)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, s.State, title, updated)
	}
	_ = w.Flush()
	return julesapi.ExitSuccess
}

func runShow(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var jsonOutput bool
	var configPath string
	var apiKey string

	fs.BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	fs.StringVar(&configPath, "config", "", "Path to config TOML file")
	fs.StringVar(&apiKey, "api-key", "", "API key override")

	flags, positionals := splitPositionalAndFlags(args)
	if err := fs.Parse(flags); err != nil {
		return julesapi.ExitUsageError
	}

	if len(positionals) < 1 {
		fmt.Fprintln(stderr, "Error: session ID is required. Usage: jules-tui show <session-id> [--json]")
		return julesapi.ExitUsageError
	}

	sessionID := positionals[0]
	client, err := createClient(configPath, apiKey)
	if err != nil {
		return handleError(err, stderr)
	}

	session, err := client.GetSession(context.Background(), sessionID)
	if err != nil {
		return handleError(err, stderr)
	}

	if jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(session); err != nil {
			return handleError(err, stderr)
		}
		return julesapi.ExitSuccess
	}

	id := session.ID
	if id == "" {
		id = strings.TrimPrefix(session.Name, "sessions/")
	}
	fmt.Fprintf(stdout, "Session ID: %s\n", id)
	if session.Title != "" {
		fmt.Fprintf(stdout, "Title:      %s\n", session.Title)
	}
	fmt.Fprintf(stdout, "State:      %s\n", session.State)
	if session.URL != "" {
		fmt.Fprintf(stdout, "URL:        %s\n", session.URL)
	}
	if session.Prompt != "" {
		fmt.Fprintf(stdout, "Prompt:     %s\n", session.Prompt)
	}
	if session.CreateTime != nil {
		fmt.Fprintf(stdout, "Created:    %s\n", session.CreateTime.Format(time.RFC3339))
	}
	if session.UpdateTime != nil {
		fmt.Fprintf(stdout, "Updated:    %s\n", session.UpdateTime.Format(time.RFC3339))
	}

	if len(session.Outputs) > 0 {
		fmt.Fprintln(stdout, "\nOutputs:")
		for _, out := range session.Outputs {
			if out.PullRequest != nil {
				fmt.Fprintf(stdout, "  Pull Request: %s\n", out.PullRequest.URL)
				if out.PullRequest.Title != "" {
					fmt.Fprintf(stdout, "    Title: %s\n", out.PullRequest.Title)
				}
				if out.PullRequest.Description != "" {
					fmt.Fprintf(stdout, "    Desc:  %s\n", out.PullRequest.Description)
				}
			}
		}
	}

	return julesapi.ExitSuccess
}

func runSend(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("send", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var messageFile string
	var messageText string
	var configPath string
	var apiKey string

	fs.StringVar(&messageFile, "message-file", "", "File containing message to send (use '-' for stdin)")
	fs.StringVar(&messageText, "message", "", "Message text string")
	fs.StringVar(&configPath, "config", "", "Path to config TOML file")
	fs.StringVar(&apiKey, "api-key", "", "API key override")

	flags, positionals := splitPositionalAndFlags(args)
	if err := fs.Parse(flags); err != nil {
		return julesapi.ExitUsageError
	}

	if len(positionals) < 1 {
		fmt.Fprintln(stderr, "Error: session ID is required. Usage: jules-tui send <session-id> (--message-file <path|-> | --message \"text\")")
		return julesapi.ExitUsageError
	}

	if (messageFile == "" && messageText == "") || (messageFile != "" && messageText != "") {
		fmt.Fprintln(stderr, "Error: exactly one of --message-file or --message must be provided")
		return julesapi.ExitUsageError
	}

	var prompt string
	if messageText != "" {
		prompt = messageText
	} else if messageFile == "-" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "Error reading message from stdin: %v\n", err)
			return julesapi.ExitGeneralError
		}
		prompt = string(data)
	} else {
		data, err := os.ReadFile(messageFile)
		if err != nil {
			fmt.Fprintf(stderr, "Error reading message file: %v\n", err)
			return julesapi.ExitGeneralError
		}
		prompt = string(data)
	}

	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		fmt.Fprintln(stderr, "Error: message content cannot be empty")
		return julesapi.ExitUsageError
	}

	sessionID := positionals[0]
	client, err := createClient(configPath, apiKey)
	if err != nil {
		return handleError(err, stderr)
	}

	if err := client.SendMessage(context.Background(), sessionID, prompt); err != nil {
		return handleError(err, stderr)
	}

	fmt.Fprintf(stdout, "Message successfully sent to session %s\n", sessionID)
	return julesapi.ExitSuccess
}

func runApprove(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("approve", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var configPath string
	var apiKey string

	fs.StringVar(&configPath, "config", "", "Path to config TOML file")
	fs.StringVar(&apiKey, "api-key", "", "API key override")

	flags, positionals := splitPositionalAndFlags(args)
	if err := fs.Parse(flags); err != nil {
		return julesapi.ExitUsageError
	}

	if len(positionals) < 1 {
		fmt.Fprintln(stderr, "Error: session ID is required. Usage: jules-tui approve <session-id>")
		return julesapi.ExitUsageError
	}

	sessionID := positionals[0]
	client, err := createClient(configPath, apiKey)
	if err != nil {
		return handleError(err, stderr)
	}

	if err := client.ApprovePlan(context.Background(), sessionID); err != nil {
		return handleError(err, stderr)
	}

	fmt.Fprintf(stdout, "Plan for session %s approved successfully\n", sessionID)
	return julesapi.ExitSuccess
}

func runNew(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var repo string
	var branch string
	var promptFile string
	var promptText string
	var title string
	var requireApproval bool
	var autoPR bool
	var jsonOutput bool
	var configPath string
	var apiKey string

	fs.StringVar(&repo, "repo", "", "Repository in owner/name format or sources/... resource name")
	fs.StringVar(&branch, "branch", "", "Starting branch name (e.g. main)")
	fs.StringVar(&promptFile, "prompt-file", "", "File containing prompt text (use '-' for stdin)")
	fs.StringVar(&promptText, "prompt", "", "Prompt text string")
	fs.StringVar(&title, "title", "", "Optional title for session")
	fs.BoolVar(&requireApproval, "require-approval", false, "Require approval before executing plan")
	fs.BoolVar(&autoPR, "auto-pr", false, "Automatically create PR when changes are ready")
	fs.BoolVar(&jsonOutput, "json", false, "Output created session in JSON format")
	fs.StringVar(&configPath, "config", "", "Path to config TOML file")
	fs.StringVar(&apiKey, "api-key", "", "API key override")

	flags, _ := splitPositionalAndFlags(args)
	if err := fs.Parse(flags); err != nil {
		return julesapi.ExitUsageError
	}

	if repo == "" {
		fmt.Fprintln(stderr, "Error: --repo is required")
		return julesapi.ExitUsageError
	}
	if branch == "" {
		fmt.Fprintln(stderr, "Error: --branch is required")
		return julesapi.ExitUsageError
	}
	if (promptFile == "" && promptText == "") || (promptFile != "" && promptText != "") {
		fmt.Fprintln(stderr, "Error: exactly one of --prompt-file or --prompt must be provided")
		return julesapi.ExitUsageError
	}

	var prompt string
	if promptText != "" {
		prompt = promptText
	} else if promptFile == "-" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "Error reading prompt from stdin: %v\n", err)
			return julesapi.ExitGeneralError
		}
		prompt = string(data)
	} else {
		data, err := os.ReadFile(promptFile)
		if err != nil {
			fmt.Fprintf(stderr, "Error reading prompt file: %v\n", err)
			return julesapi.ExitGeneralError
		}
		prompt = string(data)
	}

	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		fmt.Fprintln(stderr, "Error: prompt content cannot be empty")
		return julesapi.ExitUsageError
	}

	client, err := createClient(configPath, apiKey)
	if err != nil {
		return handleError(err, stderr)
	}

	warnFn := func(format string, a ...any) {
		fmt.Fprintf(stderr, format, a...)
	}

	sourceName, err := client.ResolveSource(context.Background(), repo, warnFn)
	if err != nil {
		return handleError(err, stderr)
	}

	autoMode := julesapi.AutomationModeUnspecified
	if autoPR {
		autoMode = julesapi.AutoCreatePR
	}

	req := &julesapi.CreateSessionRequest{
		Prompt: prompt,
		Title:  title,
		SourceContext: &julesapi.SourceContext{
			Source: sourceName,
			GitHubRepoContext: &julesapi.GitHubRepoContext{
				StartingBranch: branch,
			},
		},
		RequirePlanApproval: requireApproval,
		AutomationMode:      autoMode,
	}

	session, err := client.CreateSession(context.Background(), req)
	if err != nil {
		return handleError(err, stderr)
	}

	if jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(session); err != nil {
			return handleError(err, stderr)
		}
		return julesapi.ExitSuccess
	}

	id := session.ID
	if id == "" {
		id = strings.TrimPrefix(session.Name, "sessions/")
	}
	fmt.Fprintf(stdout, "Created session %s (%s)\n", id, session.State)
	if session.URL != "" {
		fmt.Fprintf(stdout, "URL: %s\n", session.URL)
	}

	return julesapi.ExitSuccess
}
