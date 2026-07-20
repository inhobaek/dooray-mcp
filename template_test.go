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
		{"operation": "create_template", "projectId": "12345", "templateName": "T", "subject": "S", "bodyContent": "B",
			"tagIds": "t1,t2", "toGroupIds": "g1", "ccMemberIds": "m1", "isDefault": false},
		{"operation": "update_template", "projectId": "12345", "templateId": "67890", "templateName": "T", "subject": "S", "bodyContent": "B",
			"guideContent": "guide", "priority": "normal"},
		{"operation": "delete_template", "projectId": "12345", "templateId": "67890"},
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

	// create_template without bodyContent must error before any API call.
	req2 := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "dooray_template",
			Arguments: map[string]any{"operation": "create_template", "projectId": "12345", "templateName": "T", "subject": "S"},
		},
	}
	res2, err := tool.Handler(context.Background(), req2)
	if err != nil {
		t.Fatalf("expected tool-level error result, got handler error: %v", err)
	}
	if res2 == nil || !res2.IsError {
		t.Error("expected IsError result when bodyContent is missing for create_template")
	}

	// update_template without templateId must error before any API call.
	req3 := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "dooray_template",
			Arguments: map[string]any{"operation": "update_template", "projectId": "12345", "templateName": "T", "subject": "S", "bodyContent": "B"},
		},
	}
	res3, err := tool.Handler(context.Background(), req3)
	if err != nil {
		t.Fatalf("expected tool-level error result, got handler error: %v", err)
	}
	if res3 == nil || !res3.IsError {
		t.Error("expected IsError result when templateId is missing for update_template")
	}

	// delete_template without templateId must error before any API call.
	req4 := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "dooray_template",
			Arguments: map[string]any{"operation": "delete_template", "projectId": "12345"},
		},
	}
	res4, err := tool.Handler(context.Background(), req4)
	if err != nil {
		t.Fatalf("expected tool-level error result, got handler error: %v", err)
	}
	if res4 == nil || !res4.IsError {
		t.Error("expected IsError result when templateId is missing for delete_template")
	}
}

func TestBuildTemplateRequest(t *testing.T) {
	// Missing required fields → error.
	if _, err := buildTemplateRequest(map[string]any{"templateName": "T"}); err == nil {
		t.Error("expected error when subject/bodyContent missing")
	}

	// Full build → member + group recipients, tags, guide, defaults.
	req, err := buildTemplateRequest(map[string]any{
		"templateName": "[DNS] DNS 추가/변경/삭제",
		"subject":      "DNS 추가/변경/삭제 요청",
		"bodyContent":  "## 구분\n",
		"guideContent": "가이드",
		"tagIds":       "tag1, tag2",
		"toGroupIds":   "grp1",
		"toMemberIds":  "mem1",
		"ccMemberIds":  "mem2",
		"priority":     "high",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.TemplateName == "" || req.Subject == "" || req.Body.Content == "" {
		t.Error("required fields not populated")
	}
	if !req.DueDateFlag {
		t.Error("dueDateFlag should default to true")
	}
	if req.Guide == nil || req.Guide.Content != "가이드" {
		t.Error("guide not populated")
	}
	if len(req.TagIDs) != 2 {
		t.Errorf("expected 2 tags, got %d", len(req.TagIDs))
	}
	if req.Users == nil || len(req.Users.To) != 2 || len(req.Users.Cc) != 1 {
		t.Fatalf("expected to=2 (group+member), cc=1; got %+v", req.Users)
	}
	// Verify group vs member recipient typing.
	var sawGroup, sawMember bool
	for _, r := range req.Users.To {
		switch r.Type {
		case "group":
			sawGroup = r.Group != nil && r.Group.ProjectMemberGroupID == "grp1"
		case "member":
			sawMember = r.Member != nil && r.Member.OrganizationMemberID == "mem1"
		}
	}
	if !sawGroup || !sawMember {
		t.Error("group/member recipients not marshaled correctly")
	}
}
