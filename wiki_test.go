package main

import "testing"

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
