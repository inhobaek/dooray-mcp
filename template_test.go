package main

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestTemplateToolsRegistration(t *testing.T) {
	s := newTestServer()
	token := "test-token"
	TemplateTools(s, &token)

	tools := s.ListTools()
	if _, ok := tools["dooray_template"]; !ok {
		t.Error("dooray_template tool not registered")
	}
}

func TestTemplateToolArguments(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	TemplateTools(s, &token)

	tools := s.ListTools()
	tool, ok := tools["dooray_template"]
	if !ok {
		t.Fatal("dooray_template tool not registered")
	}

	cases := []map[string]any{
		{"operation": "get_templates", "projectId": "12345"},
		{"operation": "get_templates", "projectId": "12345", "page": float64(0), "size": float64(50)},
		{"operation": "get_template", "projectId": "12345", "templateId": "67890"},
	}

	for _, args := range cases {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Name: "dooray_template", Arguments: args},
		}
		// API call fails with an invalid token; we only assert that argument
		// parsing does not panic and a result/error is returned cleanly.
		_, err := tool.Handler(context.Background(), req)
		if err != nil {
			t.Logf("handler returned expected error (API call failed) for %v: %v", args["operation"], err)
		}
	}
}

func TestTemplateToolMissingRequired(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	TemplateTools(s, &token)

	tool := s.ListTools()["dooray_template"]

	// get_template without templateId must return a tool error (not panic, not API call).
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "dooray_template",
			Arguments: map[string]any{"operation": "get_template", "projectId": "12345"},
		},
	}
	res, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("expected tool-level error result, got handler error: %v", err)
	}
	if res == nil || !res.IsError {
		t.Error("expected IsError result when templateId is missing for get_template")
	}
}
