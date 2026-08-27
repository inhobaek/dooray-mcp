package calendar

import "github.com/dooray-go/dooray-sdk/openapi/model"

// Calendar represents the calendar information in the event.
type GetEventsCalendar struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Tenant represents the tenant information in the event.
type Tenant struct {
	ID string `json:"id"`
}

// Conferencing represents the conferencing information in the event.
type Conferencing struct {
	Key         string `json:"key"`
	ServiceType string `json:"serviceType"`
	URL         string `json:"url"`
}

// Me represents the user's participation information in the event.
type Me struct {
	Type   string `json:"type"`
	Member struct {
		OrganizationMemberId string `json:"organizationMemberId"`
		EmailAddress         string `json:"emailAddress"`
		Name                 string `json:"name"`
	} `json:"member"`
	Status   string `json:"status"`
	UserType string `json:"userType"`
}

// Event represents a single event in the API response.
type Event struct {
	ID                string        `json:"id"`
	MasterScheduleID  string        `json:"masterScheduleId"`
	GetEventsCalendar Calendar      `json:"calendar"`
	Project           *Calendar     `json:"project,omitempty"`
	RecurrenceID      *string       `json:"recurrenceId,omitempty"`
	StartedAt         *string       `json:"startedAt,omitempty"`
	EndedAt           *string       `json:"endedAt,omitempty"`
	DueDate           *string       `json:"dueDate,omitempty"`
	Location          *string       `json:"location,omitempty"`
	Subject           string        `json:"subject"`
	CreatedAt         string        `json:"createdAt"`
	UpdatedAt         string        `json:"updatedAt"`
	Category          string        `json:"category"`
	Users             *struct{}     `json:"users,omitempty"`
	Me                Me            `json:"me"`
	Tenant            Tenant        `json:"tenant"`
	UID               string        `json:"uid"`
	RecurrenceType    string        `json:"recurrenceType"`
	Conferencing      *Conferencing `json:"conferencing,omitempty"`
}

// EventsResponse represents the full API response for GetEvents.
type EventsResponse struct {
	Header  model.ResponseHeader `json:"header"`
	Result  []Event              `json:"result"`
	RawJSON string               `json:"-"` // Raw JSON response for debugging or logging
}
