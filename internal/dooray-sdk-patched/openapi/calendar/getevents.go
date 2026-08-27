package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	model "github.com/dooray-go/dooray-sdk/openapi/model/calendar"
	"github.com/dooray-go/dooray-sdk/utils"
)

func (c *Calendar) GetEvents(apikey string, calendars string, timeMin time.Time, timeMax time.Time) (*model.EventsResponse, error) {
	return c.GetEventsCustomHTTPContext(context.Background(), apikey, c.httpClient, calendars, timeMin, timeMax)
}
func (c *Calendar) GetEventsContext(ctx context.Context, apikey string, calendars string, timeMin time.Time, timeMax time.Time) (*model.EventsResponse, error) {
	return c.GetEventsCustomHTTPContext(ctx, apikey, c.httpClient, calendars, timeMin, timeMax)
}
func (c *Calendar) GetEventsCustomHTTP(apikey string, httpClient *http.Client, calendars string, timeMin time.Time, timeMax time.Time) (*model.EventsResponse, error) {
	return c.GetEventsCustomHTTPContext(context.Background(), apikey, httpClient, calendars, timeMin, timeMax)
}

// GetEventsCustomHTTPContext sends a GET request to retrieve calendar events and returns a structured response.
func (c *Calendar) GetEventsCustomHTTPContext(ctx context.Context, apikey string, httpClient *http.Client, calendars string,
	timeMin time.Time, timeMax time.Time) (*model.EventsResponse, error) {
	url := c.endPoint + "/calendar/v1/calendars/*/events"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()

	if calendars != "" {
		query.Add("calendars", calendars)
	}

	query.Add("timeMin", utils.FormatTimeToISO8601(timeMin))
	query.Add("timeMax", utils.FormatTimeToISO8601(timeMax))

	req.URL.RawQuery = query.Encode()

	req.Header.Set("Authorization", "dooray-api "+apikey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the response body into EventsResponse
	var eventsResponse model.EventsResponse
	if err := json.Unmarshal(resBody, &eventsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	// Store the raw JSON response
	eventsResponse.RawJSON = string(resBody)

	return &eventsResponse, nil
}
