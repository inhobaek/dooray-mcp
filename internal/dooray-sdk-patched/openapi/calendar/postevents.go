package calendar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	model "github.com/dooray-go/dooray-sdk/openapi/model/calendar"
)

func (c *Calendar) CreateEvent(apikey string, calendarID string, event model.EventRequest) (*model.EventResponse, error) {
	return c.CreateEventCustomHTTPContext(context.Background(), apikey, c.httpClient, calendarID, event)
}
func (c *Calendar) CreateEventContext(ctx context.Context, apikey string, calendarID string, event model.EventRequest) (*model.EventResponse, error) {
	return c.CreateEventCustomHTTPContext(ctx, apikey, c.httpClient, calendarID, event)
}
func (c *Calendar) CreateEventCustomHTTP(apikey string, httpClient *http.Client, calendarID string, event model.EventRequest) (*model.EventResponse, error) {
	return c.CreateEventCustomHTTPContext(context.Background(), apikey, httpClient, calendarID, event)
}

// CreateEvent sends a POST request to create a calendar event.
func (c *Calendar) CreateEventCustomHTTPContext(ctx context.Context, apikey string, httpClient *http.Client, calendarID string, event model.EventRequest) (*model.EventResponse, error) {
	url := fmt.Sprintf("%s/calendar/v1/calendars/%s/events", c.endPoint, calendarID)

	// Serialize the event request to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("Authorization", "dooray-api "+apikey)

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create event, status: %s, body: %s", resp.Status, string(body))
	}

	// Parse the response body into EventResponse
	var eventResponse model.EventResponse
	if err := json.Unmarshal(body, &eventResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	// Store the raw JSON response
	eventResponse.RawJSON = string(body)

	return &eventResponse, nil
}
