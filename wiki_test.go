package main

import (
	"strings"
	"testing"
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
