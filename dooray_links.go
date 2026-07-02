package main

import (
	"encoding/json"
	"regexp"
	"sync"

	"github.com/dooray-go/dooray-sdk/openapi/project"
)

// plainDoorayURL matches a bare Dooray task/page URL that is NOT already
// wrapped in a markdown link, e.g. https://nhnent.dooray.com/project/tasks/123
// or https://nhnent.dooray.com/project/pages/123. Wrapping these as
// [text](dooray://{orgId}/tasks/{id} "state") makes the Dooray editor render
// them as inline link cards instead of showing the raw URL.
var plainDoorayURL = regexp.MustCompile(`(^|[^(\]])(https://[a-zA-Z0-9-]+\.dooray\.com/project/(tasks|pages)/(\d+))`)

var (
	orgIDOnce   sync.Once
	cachedOrgID string
)

// organizationID returns the organization id for the token's account,
// fetched once via find_projects and cached for the process lifetime
// (a single MCP server instance talks to exactly one Dooray organization).
func organizationID(token string) string {
	orgIDOnce.Do(func() {
		res, err := project.NewDefaultProject().GetProjects(token, "public", "private", "active")
		if err != nil {
			return
		}
		var parsed struct {
			Result []struct {
				Organization struct {
					ID string `json:"id"`
				} `json:"organization"`
			} `json:"result"`
		}
		if json.Unmarshal([]byte(res.RawJSON), &parsed) != nil {
			return
		}
		for _, p := range parsed.Result {
			if p.Organization.ID != "" {
				cachedOrgID = p.Organization.ID
				break
			}
		}
	})
	return cachedOrgID
}

// linkifyDoorayURLs rewrites bare Dooray task/page URLs found in markdown
// content into dooray:// inline link cards, e.g.:
//
//	https://nhnent.dooray.com/project/tasks/123
//	-> [https://nhnent.dooray.com/project/tasks/123](dooray://{orgId}/tasks/123)
//
// URLs already inside a markdown link (preceded by '(' or ']') are left as-is.
// No-op if the organization id can't be resolved.
func linkifyDoorayURLs(token, mimeType, content string) string {
	if mimeType != "" && mimeType != "text/x-markdown" {
		return content
	}
	orgID := organizationID(token)
	if orgID == "" {
		return content
	}
	return linkifyContent(orgID, content)
}

func linkifyContent(orgID, content string) string {
	return plainDoorayURL.ReplaceAllString(content, "${1}[${2}](dooray://"+orgID+"/${3}/${4})")
}
