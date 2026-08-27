package calendar

import "github.com/dooray-go/dooray-sdk/openapi/model"

// EventResponseResult represents the result part of the API response.
type EventResponseResult struct {
	ID string `json:"id"` // 생성된 event ID
}

// EventResponse represents the full API response.
type EventResponse struct {
	Header  model.ResponseHeader      `json:"header"`
	Result  EventResponseResult `json:"result"`
	RawJSON string              `json:"-"` // Raw JSON response for debugging or logging
}
