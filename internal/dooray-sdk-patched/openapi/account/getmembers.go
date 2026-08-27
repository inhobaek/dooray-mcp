package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dooray-go/dooray-sdk/openapi/model/account"
	model "github.com/dooray-go/dooray-sdk/openapi/model/account"
)

func (a *Account) GetMembers(apikey string, name string, userCode string) (*model.GetMembersResponse, error) {
	return a.GetMembersCustomHTTPContext(context.Background(), apikey, a.httpClient, name, userCode)
}

func (a *Account) GetMembersContext(ctx context.Context, apikey string, name string, userCode string) (*model.GetMembersResponse, error) {
	return a.GetMembersCustomHTTPContext(ctx, apikey, a.httpClient, name, userCode)
}

func (a *Account) GetMembersCustomHTTP(apikey string, httpClient *http.Client, name string, userCode string) (*model.GetMembersResponse, error) {
	return a.GetMembersCustomHTTPContext(context.Background(), apikey, httpClient, name, userCode)
}

func (a *Account) GetMembersCustomHTTPContext(ctx context.Context, apikey string, httpClient *http.Client, name string, userCode string) (*model.GetMembersResponse, error) {

	url := a.endPoint + "/common/v1/members"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed new request: %w", err)
	}

	query := req.URL.Query()

	if name != "" {
		query.Add("name", name)
	}

	if userCode != "" {
		query.Add("userCode", userCode)
	}

	req.URL.RawQuery = query.Encode()

	req.Header.Set("Authorization", fmt.Sprintf("dooray-api %s", apikey))

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response body into GetMembersResponse
	var membersResponse account.GetMembersResponse
	if err := json.Unmarshal(body, &membersResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	// Store the raw JSON response
	membersResponse.RawJSON = string(body)

	return &membersResponse, nil
}
