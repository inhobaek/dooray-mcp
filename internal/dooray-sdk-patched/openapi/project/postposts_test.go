package project

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	model "github.com/dooray-go/dooray-sdk/openapi/model/project"
	"github.com/dooray-go/dooray-sdk/utils"
)

func TestCreatePost(t *testing.T) {
	projectID := "1234567890"
	expectedAPIKey := "test-api-key"

	// Mock server to simulate the API response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the HTTP method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Check the headers
		expectedAuth := "dooray-api " + expectedAPIKey
		if r.Header.Get("Authorization") != expectedAuth {
			t.Errorf("Expected Authorization header '%s', got %s", expectedAuth, r.Header.Get("Authorization"))
		}

		// Check the Content-Type header
		if r.Header.Get("Content-Type") != "application/json;charset=utf-8" {
			t.Errorf("Expected Content-Type 'application/json;charset=utf-8', got %s", r.Header.Get("Content-Type"))
		}

		// Check the URL path
		expectedPath := "/project/v1/projects/" + projectID + "/posts"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got %s", expectedPath, r.URL.Path)
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		defer r.Body.Close()

		// Verify the request body can be unmarshaled
		var postRequest model.PostRequest
		if err := json.Unmarshal(body, &postRequest); err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}

		// Check some fields from the request
		if postRequest.Subject != "Test Post" {
			t.Errorf("Expected subject 'Test Post', got %s", postRequest.Subject)
		}

		if postRequest.Body.Content != "This is a test post." {
			t.Errorf("Expected body content 'This is a test post.', got %s", postRequest.Body.Content)
		}

		// Respond with a mock response
		response := `{
			"header": {
				"isSuccessful": true,
				"resultCode": 0,
				"resultMessage": ""
			},
			"result": {
				"id": "987654321"
			}
		}`
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	}))
	defer mockServer.Close()

	// Create a Project instance with the mock server endpoint
	project := NewProject(mockServer.URL)

	// Create due date
	dueDate := utils.NewJsonTime(time.Now().Add(24 * time.Hour))

	// Create a sample PostRequest
	postRequest := model.PostRequest{
		Subject: "Test Post",
		Body: model.PostBody{
			MimeType: "text/html",
			Content:  "This is a test post.",
		},
		Users: &model.PostUsers{
			To: []model.PostRecipient{
				{
					Type: "member",
					Member: &model.PostMember{
						OrganizationMemberID: "member123",
					},
				},
			},
		},
		DueDate:  &dueDate,
		Priority: "normal",
		TagIDs:   []string{"tag1", "tag2"},
	}

	// Call the CreatePost method
	response, err := project.CreatePost(expectedAPIKey, projectID, postRequest)
	if err != nil {
		t.Fatalf("CreatePost returned an error: %v", err)
	}

	// Verify the response
	if response.Header.IsSuccessful != true {
		t.Errorf("Expected isSuccessful to be true, got %v", response.Header.IsSuccessful)
	}

	if response.Result.ID != "987654321" {
		t.Errorf("Expected result ID '987654321', got %s", response.Result.ID)
	}

	if response.RawJSON == "" {
		t.Error("Expected RawJSON to be populated, got empty string")
	}
}

func TestCreatePost_ErrorResponse(t *testing.T) {
	projectID := "1234567890"

	// Mock server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid request"}`))
	}))
	defer mockServer.Close()

	// Create a Project instance with the mock server endpoint
	project := NewProject(mockServer.URL)

	// Create a sample PostRequest
	postRequest := model.PostRequest{
		Subject: "Test Post",
		Body: model.PostBody{
			MimeType: "text/html",
			Content:  "This is a test post.",
		},
	}

	// Call the CreatePost method
	_, err := project.CreatePost("test-api-key", projectID, postRequest)
	if err == nil {
		t.Fatal("Expected CreatePost to return an error, but got nil")
	}

	// Verify error message contains status information
	expectedErrSubstring := "failed to create post"
	if !contains(err.Error(), expectedErrSubstring) {
		t.Errorf("Expected error to contain '%s', got: %s", expectedErrSubstring, err.Error())
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}