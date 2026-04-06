package gh

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// ghPRStatusResult represents the JSON output format from gh pr status --json.
type ghPRStatusResult struct {
	CurrentBranch *ghCurrentBranchPR `json:"currentBranch"`
	CreatedBy     []ghPRSummary      `json:"createdBy"`
	NeedsReview   []ghPRSummary      `json:"needsReview"`
}

// ghCurrentBranchPR represents the current branch PR in gh JSON output.
type ghCurrentBranchPR struct {
	Number            int             `json:"number"`
	Title             string          `json:"title"`
	HeadRefName       string          `json:"headRefName"`
	URL               string          `json:"url"`
	State             string          `json:"state"`
	ReviewDecision    string          `json:"reviewDecision,omitempty"`
	StatusCheckRollup []ghStatusCheck `json:"statusCheckRollup,omitempty"`
}

// ghPRSummary represents a brief PR summary in gh JSON output.
type ghPRSummary struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	HeadRefName string `json:"headRefName"`
	URL         string `json:"url"`
}

// PRStatusParser parses the output of 'gh pr status'.
type PRStatusParser struct {
	schema domain.Schema
}

// NewPRStatusParser creates a new PRStatusParser with the gh-pr-status schema.
func NewPRStatusParser() *PRStatusParser {
	return &PRStatusParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/gh-pr-status.json",
			"GitHub PR Status Output",
			"object",
			map[string]domain.PropertySchema{
				"current_branch":    {Type: "object", Description: "PR for current branch"},
				"created_by_you":    {Type: "array", Description: "PRs created by you"},
				"requesting_review": {Type: "array", Description: "PRs requesting your review"},
			},
			[]string{"created_by_you", "requesting_review"},
		),
	}
}

// Parse reads gh pr status JSON output and returns structured data.
func (p *PRStatusParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("read input: %w", err),
			"",
			0,
		), nil
	}

	raw := string(data)

	var ghResult ghPRStatusResult
	if err := json.Unmarshal(data, &ghResult); err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("parse JSON: %w", err),
			raw,
			0,
		), nil
	}

	// Convert to our domain types
	result := &PRStatusResult{
		CurrentBranch:    convertCurrentBranchPR(ghResult.CurrentBranch),
		CreatedByYou:     convertPRSummaries(ghResult.CreatedBy),
		RequestingReview: convertPRSummaries(ghResult.NeedsReview),
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for gh pr status output.
func (p *PRStatusParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *PRStatusParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "gh" {
		return false
	}
	if len(subcommands) < 2 {
		return false
	}
	return subcommands[0] == "pr" && subcommands[1] == "status"
}

// convertCurrentBranchPR converts gh current branch PR to our domain type.
func convertCurrentBranchPR(ghPR *ghCurrentBranchPR) *CurrentBranchPR {
	if ghPR == nil {
		return nil
	}

	pr := &CurrentBranchPR{
		Number:     ghPR.Number,
		Title:      ghPR.Title,
		HeadBranch: ghPR.HeadRefName,
		URL:        ghPR.URL,
		State:      ghPR.State,
	}

	// Convert review decision to human-readable status
	if ghPR.ReviewDecision != "" {
		pr.ReviewStatus = ghPR.ReviewDecision
	}

	// Derive check status from status check rollup
	if len(ghPR.StatusCheckRollup) > 0 {
		pr.CheckStatus = deriveCheckStatus(ghPR.StatusCheckRollup)
	}

	return pr
}

// convertPRSummaries converts gh PR summaries to our domain type.
func convertPRSummaries(ghSummaries []ghPRSummary) []PRSummary {
	summaries := make([]PRSummary, 0, len(ghSummaries))
	for _, s := range ghSummaries {
		summaries = append(summaries, PRSummary{
			Number:     s.Number,
			Title:      s.Title,
			HeadBranch: s.HeadRefName,
			URL:        s.URL,
		})
	}
	return summaries
}

// deriveCheckStatus determines overall check status from individual checks.
func deriveCheckStatus(checks []ghStatusCheck) string {
	if len(checks) == 0 {
		return ""
	}

	allSuccess := true
	for _, c := range checks {
		if c.Status != "SUCCESS" && c.Conclusion != "SUCCESS" {
			allSuccess = false
			break
		}
	}

	if allSuccess {
		return "All checks passing"
	}
	return "Some checks pending or failing"
}
