package project

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	model "github.com/dooray-go/dooray-sdk/openapi/model/project"
)

func (c *Project) CreatePost(apikey string, projectID string, post model.PostRequest) (*model.PostResponse, error) {
	return c.CreatePostCustomHTTPContext(context.Background(), apikey, c.httpClient, projectID, post)
}

func (c *Project) CreatePostContext(ctx context.Context, apikey string, projectID string, post model.PostRequest) (*model.PostResponse, error) {
	return c.CreatePostCustomHTTPContext(ctx, apikey, c.httpClient, projectID, post)
}

func (c *Project) CreatePostCustomHTTP(apikey string, httpClient *http.Client, projectID string, post model.PostRequest) (*model.PostResponse, error) {
	return c.CreatePostCustomHTTPContext(context.Background(), apikey, httpClient, projectID, post)
}

// CreatePost sends a POST request to create a project post.
func (c *Project) CreatePostCustomHTTPContext(ctx context.Context, apikey string, httpClient *http.Client, projectID string, post model.PostRequest) (*model.PostResponse, error) {
	url := fmt.Sprintf("%s/project/v1/projects/%s/posts", c.endPoint, projectID)

	// Serialize the post request to JSON
	payload, err := json.Marshal(post)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal post request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
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
		return nil, fmt.Errorf("failed to create post, status: %s, body: %s", resp.Status, string(body))
	}

	// Parse the response body into PostResponse
	var postResponse model.PostResponse
	if err := json.Unmarshal(body, &postResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	// Store the raw JSON response
	postResponse.RawJSON = string(body)

	return &postResponse, nil
}