package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	model "github.com/dooray-go/dooray-sdk/openapi/model/calendar"
	"io"
	"net/http"
)

func (c *Calendar) GetCalendars(apikey string) (*model.GetCalendarsResponse, error) {
	return c.GetCalendarsCustomHTTPContext(context.Background(), apikey, c.httpClient)
}
func (c *Calendar) GetCalendarsContext(ctx context.Context, apikey string) (*model.GetCalendarsResponse, error) {
	return c.GetCalendarsCustomHTTPContext(ctx, apikey, c.httpClient)
}
func (c *Calendar) GetCalendarsCustomHTTP(apikey string, httpClient *http.Client) (*model.GetCalendarsResponse, error) {
	return c.GetCalendarsCustomHTTPContext(context.Background(), apikey, httpClient)
}
func (c *Calendar) GetCalendarsCustomHTTPContext(ctx context.Context, apikey string, httpClient *http.Client) (*model.GetCalendarsResponse, error) {
	url := c.endPoint + "/calendar/v1/calendars"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	req.URL.RawQuery = query.Encode()

	req.Header.Set("Authorization", "dooray-api "+apikey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		resp.Body.Close()
	}()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the response body into EventsResponse
	var calendarsResponse model.GetCalendarsResponse
	if err := json.Unmarshal(resBody, &calendarsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	// Store the raw JSON response
	calendarsResponse.RawJSON = string(resBody)

	return &calendarsResponse, nil
}
