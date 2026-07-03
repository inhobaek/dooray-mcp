package main

import (
	"strings"
	"testing"
)

func TestPageIDFromInput(t *testing.T) {
	cases := map[string]string{
		"3205786621378362771": "3205786621378362771",
		"https://nhnent.dooray.com/project/pages/3205786621378362771": "3205786621378362771",
		"dooray://1387695619080878080/pages/3169680032924002758":      "3169680032924002758", // last numeric segment wins
		"  3169680032924002758  ":                                     "3169680032924002758", // trimmed
	}
	for in, want := range cases {
		if got := pageIDFromInput(in); got != want {
			t.Errorf("pageIDFromInput(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWikiInsertMarkdown(t *testing.T) {
	ok := `{"header":{"isSuccessful":true},"result":{"attachFileId":"2541304532468051951","name":"a.png"}}`
	got := wikiInsertMarkdown(ok, "4365528512344859383")
	want := "본문 삽입: ![a.png](/wikis/4365528512344859383/files/2541304532468051951)"
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
