package calendar

import (
	"encoding/json"
	"github.com/dooray-go/dooray-sdk/openapi/model"
)

// CalendarMe represents the "me" field in the calendar response.
type CalendarMe struct {
	Default bool   `json:"default"` // "true" or "false"
	Color   string `json:"color"`
	Listed  bool   `json:"listed"`  // "true" or "false"
	Checked bool   `json:"checked"` // "true" or "false"
	Role    string `json:"role"`    // e.g., "owner"
	Order   int    `json:"order"`
}

// Calendar represents a single calendar in the response.
type Calendar struct {
	ID                        string     `json:"id"`
	Name                      string     `json:"name"`
	Type                      string     `json:"type"` // "private", "project", "subscription"
	CreatedAt                 string     `json:"createdAt"`
	OwnerOrganizationMemberID string     `json:"ownerOrganizationMemberId"`
	ProjectID                 *string    `json:"projectId,omitempty"` // Only for project calendars
	Me                        CalendarMe `json:"me"`
}

// GetCalendarsResponse represents the full API response for retrieving calendars.
type GetCalendarsResponse struct {
	Header     model.ResponseHeader `json:"header"`
	Result     []Calendar           `json:"result"`
	TotalCount json.Number          `json:"totalCount"`
	RawJSON    string               `json:"-"` // 원본 JSON 응답 (디버깅 또는 로깅용)
}
