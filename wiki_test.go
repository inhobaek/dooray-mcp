package main

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestPageIDFromInput(t *testing.T) {
	cases := map[string]string{
		"3000000000000000001": "3000000000000000001",
		"https://example.dooray.com/project/pages/3000000000000000001": "3000000000000000001",
		"dooray://1234567890123456789/pages/3000000000000000002":      "3000000000000000002", // last numeric segment wins
		"  3000000000000000002  ":                                     "3000000000000000002", // trimmed
	}
	for in, want := range cases {
		if got := pageIDFromInput(in); got != want {
			t.Errorf("pageIDFromInput(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWikiInsertMarkdown(t *testing.T) {
	ok := `{"header":{"isSuccessful":true},"result":{"attachFileId":"2000000000000000001","name":"a.png"}}`
	got := wikiInsertMarkdown(ok, "4000000000000000001")
	want := "본문 삽입: ![a.png](/wikis/4000000000000000001/files/2000000000000000001)"
	if got != want {
		t.Errorf("wikiInsertMarkdown ok = %q, want %q", got, want)
	}

	// missing attachFileId → hint, not a broken markdown link
	bad := wikiInsertMarkdown(`{"result":{}}`, "w1")
	if !strings.Contains(bad, "실패") {
		t.Errorf("wikiInsertMarkdown missing-id should hint failure, got %q", bad)
	}
	// malformed JSON → hint, no panic
	if got := wikiInsertMarkdown("not json", "w1"); !strings.Contains(got, "실패") {
		t.Errorf("wikiInsertMarkdown bad-json should hint failure, got %q", got)
	}
}

// delete_file/delete_page 2단계 확인 가드: 필수 인자 누락과 위조 토큰은 네트워크를 타지 않고 에러여야 한다.
func TestWikiDeleteGuards(t *testing.T) {
	s := newTestServer()
	token := "invalid-token"
	WikiTools(s, &token)

	tool, ok := s.ListTools()["dooray_wiki"]
	if !ok {
		t.Fatal("dooray_wiki tool not registered")
	}

	call := func(op string, args map[string]any) *mcp.CallToolResult {
		args["operation"] = op
		res, err := tool.Handler(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{Name: "dooray_wiki", Arguments: args},
		})
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}
		return res
	}

	if res := call("delete_file", map[string]any{"wikiId": "1", "pageId": "2"}); !res.IsError {
		t.Error("delete_file: expected error when fileId is missing")
	}
	if res := call("delete_file", map[string]any{"wikiId": "1", "pageId": "2", "fileId": "3", "confirmToken": "deadbeef"}); !res.IsError {
		t.Error("delete_file: expected error for a confirmToken that was never issued")
	}
	if res := call("delete_page", map[string]any{"wikiId": "1"}); !res.IsError {
		t.Error("delete_page: expected error when pageId is missing")
	}
	if res := call("delete_page", map[string]any{"wikiId": "1", "pageId": "2", "confirmToken": "deadbeef"}); !res.IsError {
		t.Error("delete_page: expected error for a confirmToken that was never issued")
	}
}
