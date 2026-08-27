package calendar

import "github.com/dooray-go/dooray-sdk/utils"

// Member represents a member in the event.
type Member struct {
	OrganizationMemberId string `json:"organizationMemberId,omitempty"`
}

// EmailUser represents an email user in the event.
type EmailUser struct {
	EmailAddress string `json:"emailAddress,omitempty"`
	Name         string `json:"name,omitempty"`
}

// Recipient represents a recipient (to or cc) in the event.
type Recipient struct {
	Type      string     `json:"type"`
	Member    *Member    `json:"member,omitempty"`
	EmailUser *EmailUser `json:"emailUser,omitempty"`
}

// Users represents the users involved in the event.
type Users struct {
	To []Recipient `json:"to,omitempty"`
	Cc []Recipient `json:"cc,omitempty"`
}

// Body represents the body of the event.
type Body struct {
	MimeType string `json:"mimeType"`
	Content  string `json:"content"`
}

// RecurrenceRule represents the recurrence rule for the event.
type RecurrenceRule struct {
	Frequency    string `json:"frequency"`
	Interval     int    `json:"interval"`
	Until        string `json:"until"`
	Byday        string `json:"byday"`
	Bymonth      string `json:"bymonth"`
	Bymonthday   string `json:"bymonthday"`
	TimezoneName string `json:"timezoneName"`
}

// Alarm represents an alarm setting for the event.
type Alarm struct {
	Action  string `json:"action"`
	Trigger string `json:"trigger"`
}

// PersonalSettings represents personal settings for the event.
type PersonalSettings struct {
	Alarms []Alarm `json:"alarms"`
	Busy   bool    `json:"busy"`
	Class  string  `json:"class"`
}

// EventRequest represents the payload for creating a calendar event.
type EventRequest struct {
	Users            Users             `json:"users"`
	Subject          string            `json:"subject"`
	Body             Body              `json:"body"`
	StartedAt        utils.JsonTime    `json:"startedAt"`
	EndedAt          utils.JsonTime    `json:"endedAt"`
	WholeDayFlag     bool              `json:"wholeDayFlag"`
	Location         string            `json:"location"`
	RecurrenceRule   *RecurrenceRule   `json:"recurrenceRule,omitempty"`
	PersonalSettings *PersonalSettings `json:"personalSettings,omitempty"`
}
