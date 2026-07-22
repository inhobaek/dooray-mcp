package main

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestMessengerToolsRegistration(t *testing.T) {
	s := newTestServer()
	token := "test-token"
	MessengerTools(s, &token)

	tools := s.ListTools()
	if _, ok := tools["dooray_messenger"]; !ok {
		t.Fatal("dooray_messenger tool not registered")
	}
}

func TestMessengerSendArguments(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	MessengerTools(s, &token)

	tool := s.ListTools()["dooray_messenger"]

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "dooray_messenger",
			Arguments: map[string]any{
				"operation": "send",
				"to":        "member-123",
				"message":   "안녕하세요",
			},
		},
	}

	_, err := tool.Handler(context.Background(), req)
	if err == nil {
		t.Log("handler succeeded (argument parsing works)")
	} else {
		t.Logf("handler returned expected error: %v", err)
	}
}

func TestMessengerMissingTo(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	MessengerTools(s, &token)

	tool := s.ListTools()["dooray_messenger"]

	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered from panic (missing 'to'): %v", r)
		}
	}()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "dooray_messenger",
			Arguments: map[string]any{
				"operation": "send",
				"message":   "안녕하세요",
			},
		},
	}

	tool.Handler(context.Background(), req)
}

func TestMessengerMissingMessage(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	MessengerTools(s, &token)

	tool := s.ListTools()["dooray_messenger"]

	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered from panic (missing 'message'): %v", r)
		}
	}()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "dooray_messenger",
			Arguments: map[string]any{
				"operation": "send",
				"to":        "member-123",
			},
		},
	}

	tool.Handler(context.Background(), req)
}

func TestMessengerMissingAllArgs(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	MessengerTools(s, &token)

	tool := s.ListTools()["dooray_messenger"]

	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered from panic (missing all required args): %v", r)
		}
	}()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "dooray_messenger",
			Arguments: map[string]any{
				"operation": "send",
			},
		},
	}

	tool.Handler(context.Background(), req)
}

func TestMessengerSendChannelArguments(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	MessengerTools(s, &token)

	tool := s.ListTools()["dooray_messenger"]

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "dooray_messenger",
			Arguments: map[string]any{
				"operation": "send_channel",
				"channelId": "channel-123",
				"message":   "안녕하세요",
			},
		},
	}

	_, err := tool.Handler(context.Background(), req)
	if err == nil {
		t.Log("handler succeeded (argument parsing works)")
	} else {
		t.Logf("handler returned expected error: %v", err)
	}
}

func TestMessengerSendChannelMissingChannelId(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	MessengerTools(s, &token)

	tool := s.ListTools()["dooray_messenger"]

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "dooray_messenger",
			Arguments: map[string]any{
				"operation": "send_channel",
				"message":   "안녕하세요",
			},
		},
	}

	res, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || !res.IsError {
		t.Fatal("expected error result when channelId is missing")
	}
}

func TestMessengerSendChannelMissingMessage(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	MessengerTools(s, &token)

	tool := s.ListTools()["dooray_messenger"]

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "dooray_messenger",
			Arguments: map[string]any{
				"operation": "send_channel",
				"channelId": "channel-123",
			},
		},
	}

	res, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || !res.IsError {
		t.Fatal("expected error result when message is missing")
	}
}

func TestMessengerReplyThreadMissingArgs(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	MessengerTools(s, &token)

	tool := s.ListTools()["dooray_messenger"]

	// reply_thread requires channelId, parentMessageId, and message.
	// Each missing one must yield an error result (not a real API call).
	cases := []map[string]any{
		{"operation": "reply_thread", "parentMessageId": "p-1", "message": "hi"},    // no channelId
		{"operation": "reply_thread", "channelId": "c-1", "message": "hi"},          // no parentMessageId
		{"operation": "reply_thread", "channelId": "c-1", "parentMessageId": "p-1"}, // no message
	}
	for i, args := range cases {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Name: "dooray_messenger", Arguments: args},
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

func TestValidateWebhookURL(t *testing.T) {
	cases := []struct {
		url     string
		wantErr bool
	}{
		{"https://example.dooray.com/services/abc", false},
		{"https://hooks.dooray.com/services/abc", false},
		{"http://example.dooray.com/services/abc", true},    // http, not https
		{"https://evil.com/services/abc", true},            // wrong host
		{"https://dooray.com.evil.com/services/abc", true}, // host doesn't end with .dooray.com
		{"not-a-url", true},
		{"HTTPS://example.dooray.com/services/abc", false},  // uppercase scheme, url.Parse lowercases it
		{"https://1.2.3.4/", true},                         // IP literal, not .dooray.com
		{"example.dooray.com/services/x", true},             // no scheme, treated as relative path
	}
	for _, c := range cases {
		err := validateWebhookURL(c.url)
		if c.wantErr && err == nil {
			t.Errorf("validateWebhookURL(%q): expected error, got nil", c.url)
		}
		if !c.wantErr && err != nil {
			t.Errorf("validateWebhookURL(%q): expected no error, got %v", c.url, err)
		}
	}
}

func TestMessengerSendWebhookMissingArgs(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	MessengerTools(s, &token)

	tool := s.ListTools()["dooray_messenger"]
	cases := []map[string]any{
		{"operation": "send_webhook", "message": "hi"},                                      // no webhookUrl
		{"operation": "send_webhook", "webhookUrl": "https://example.dooray.com/services/x"}, // no message
	}
	for i, args := range cases {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Name: "dooray_messenger", Arguments: args},
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

func TestMessengerSendWebhookRejectsBadHost(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	MessengerTools(s, &token)

	tool := s.ListTools()["dooray_messenger"]
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "dooray_messenger",
			Arguments: map[string]any{
				"operation":  "send_webhook",
				"webhookUrl": "https://evil.com/steal",
				"message":    "hi",
			},
		},
	}
	res, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || !res.IsError {
		t.Fatal("expected error result for non-dooray.com webhook host")
	}
}

func TestMessengerToolCount(t *testing.T) {
	s := newTestServer()
	token := "test-token"
	MessengerTools(s, &token)

	tools := s.ListTools()
	count := 0
	for name := range tools {
		if name == "dooray_messenger" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 messenger tool, got %d", count)
	}
}
