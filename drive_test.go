package main

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestDriveToolsRegistration(t *testing.T) {
	s := newTestServer()
	token := "test-token"
	DriveTools(s, &token)

	tools := s.ListTools()
	if _, ok := tools["dooray_drive"]; !ok {
		t.Fatal("dooray_drive tool not registered")
	}
}

func TestDriveFindDrivesMissingQueryString(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	DriveTools(s, &token)

	tool := s.ListTools()["dooray_drive"]
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "dooray_drive",
			Arguments: map[string]any{"operation": "find_drives"},
		},
	}
	res, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || !res.IsError {
		t.Fatal("expected error result when queryString is missing")
	}
}

func TestDriveFindFoldersMissingArgs(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	DriveTools(s, &token)

	tool := s.ListTools()["dooray_drive"]
	cases := []map[string]any{
		{"operation": "find_folders", "queryString": "type=folder"},       // no driveId
		{"operation": "find_folders", "driveId": "drive-1"},               // no queryString
	}
	for i, args := range cases {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Name: "dooray_drive", Arguments: args},
		}
		res, err := tool.Handler(context.Background(), req)
		if err != nil {
			t.Fatalf("case %d: unexpected error: %v", i, err)
		}
		if res == nil || !res.IsError {
			t.Fatalf("case %d: expected error result for missing required arg", i)
		}
	}
}

func TestDriveUploadFileMissingArgs(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	DriveTools(s, &token)

	tool := s.ListTools()["dooray_drive"]
	cases := []map[string]any{
		{"operation": "upload_file", "parentId": "p-1", "filePath": "/tmp/x.txt"},              // no driveId
		{"operation": "upload_file", "driveId": "d-1", "filePath": "/tmp/x.txt"},                // no parentId
		{"operation": "upload_file", "driveId": "d-1", "parentId": "p-1"},                       // no filePath
	}
	for i, args := range cases {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Name: "dooray_drive", Arguments: args},
		}
		res, err := tool.Handler(context.Background(), req)
		if err != nil {
			t.Fatalf("case %d: unexpected error: %v", i, err)
		}
		if res == nil || !res.IsError {
			t.Fatalf("case %d: expected error result for missing required arg", i)
		}
	}
}

func TestDriveUploadContentMissingArgs(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	DriveTools(s, &token)

	tool := s.ListTools()["dooray_drive"]
	cases := []map[string]any{
		{"operation": "upload_content", "parentId": "p-1", "fileName": "a.md", "content": "hi"}, // no driveId
		{"operation": "upload_content", "driveId": "d-1", "fileName": "a.md", "content": "hi"},  // no parentId
		{"operation": "upload_content", "driveId": "d-1", "parentId": "p-1", "content": "hi"},   // no fileName
		{"operation": "upload_content", "driveId": "d-1", "parentId": "p-1", "fileName": "a.md"}, // no content
	}
	for i, args := range cases {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Name: "dooray_drive", Arguments: args},
		}
		res, err := tool.Handler(context.Background(), req)
		if err != nil {
			t.Fatalf("case %d: unexpected error: %v", i, err)
		}
		if res == nil || !res.IsError {
			t.Fatalf("case %d: expected error result for missing required arg", i)
		}
	}
}
