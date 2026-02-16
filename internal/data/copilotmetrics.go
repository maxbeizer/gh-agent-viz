package data

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// CopilotOrgMetrics represents aggregate Copilot usage for an org
type CopilotOrgMetrics struct {
	Date               string `json:"date"`
	TotalActiveUsers   int    `json:"total_active_users"`
	TotalEngagedUsers  int    `json:"total_engaged_users"`
	CodeCompletions    *CopilotCodeCompletions `json:"copilot_ide_code_completions"`
	IDEChat            *CopilotIDEChat         `json:"copilot_ide_chat"`
	DotcomChat         *CopilotDotcomChat      `json:"copilot_dotcom_chat"`
	DotcomPullRequests *CopilotDotcomPRs       `json:"copilot_dotcom_pull_requests"`
}

// CopilotCodeCompletions holds code completion metrics
type CopilotCodeCompletions struct {
	TotalEngagedUsers int `json:"total_engaged_users"`
}

// CopilotIDEChat holds IDE chat metrics
type CopilotIDEChat struct {
	TotalEngagedUsers int `json:"total_engaged_users"`
}

// CopilotDotcomChat holds GitHub.com chat metrics
type CopilotDotcomChat struct {
	TotalEngagedUsers int `json:"total_engaged_users"`
}

// CopilotDotcomPRs holds PR summary metrics
type CopilotDotcomPRs struct {
	TotalEngagedUsers int `json:"total_engaged_users"`
}

// OrgMetricsResult represents the result of fetching org metrics
type OrgMetricsResult struct {
	Available bool                // whether metrics are accessible
	Metrics   []CopilotOrgMetrics // daily metrics (most recent first)
	Error     string              // user-facing reason if unavailable
}

// FetchOrgMetrics attempts to fetch Copilot metrics for the given org.
// Returns Available=false silently when the user lacks admin access (403)
// or the endpoint doesn't exist (404). Only surfaces real errors.
func FetchOrgMetrics(org string) OrgMetricsResult {
	if strings.TrimSpace(org) == "" {
		return OrgMetricsResult{Available: false}
	}

	endpoint := fmt.Sprintf("/orgs/%s/copilot/metrics?per_page=7", org)
	cmd := exec.Command("gh", "api", endpoint, "--jq", ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := strings.TrimSpace(string(output))
		// 403 = no admin access, 404 = endpoint/org not found — both silently hide
		if strings.Contains(outputStr, "403") || strings.Contains(outputStr, "404") ||
			strings.Contains(outputStr, "Not Found") || strings.Contains(outputStr, "admin rights") {
			return OrgMetricsResult{Available: false}
		}
		return OrgMetricsResult{Available: false, Error: outputStr}
	}

	var metrics []CopilotOrgMetrics
	if err := json.Unmarshal(output, &metrics); err != nil {
		return OrgMetricsResult{Available: false, Error: "failed to parse metrics response"}
	}

	if len(metrics) == 0 {
		return OrgMetricsResult{Available: false}
	}

	return OrgMetricsResult{Available: true, Metrics: metrics}
}
