package project

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	model "github.com/dooray-go/dooray-sdk/openapi/model/project"
)

func (c *Project) GetProjects(apikey string, projectType string, scope string, state string) (*model.GetProjectsResponse, error) {
	return c.GetProjectsCustomHTTPContext(context.Background(), apikey, c.httpClient, projectType, scope, state)
}

func (c *Project) GetProjectsContext(ctx context.Context, apikey string, projectType string, scope string, state string) (*model.GetProjectsResponse, error) {
	return c.GetProjectsCustomHTTPContext(ctx, apikey, c.httpClient, projectType, scope, state)
}

func (c *Project) GetProjectsCustomHTTP(apikey string, httpClient *http.Client, projectType string, scope string, state string) (*model.GetProjectsResponse, error) {
	return c.GetProjectsCustomHTTPContext(context.Background(), apikey, httpClient, projectType, scope, state)
}

func (c *Project) GetProjectsCustomHTTPContext(ctx context.Context, apikey string, httpClient *http.Client, projectType string, scope string, state string) (*model.GetProjectsResponse, error) {
	url := c.endPoint + "/project/v1/projects"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()

	query.Add("member", "me")
	query.Add("size", "100")

	if projectType != "" {
		query.Add("type", projectType)
	}

	if scope != "" {
		query.Add("scope", scope)
	}

	if state != "" {
		query.Add("state", state)
	}

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

	var response model.GetProjectsResponse
	if err := json.Unmarshal(resBody, &response); err != nil {
		return nil, err
	}

	response.RawJSON = string(resBody)

	return &response, nil
}
