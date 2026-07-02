package main

import "testing"

func TestLinkifyContent(t *testing.T) {
	const orgID = "1387695619080878080"
	cases := map[string]string{
		"https://nhnent.dooray.com/project/tasks/123":                         "[https://nhnent.dooray.com/project/tasks/123](dooray://" + orgID + "/tasks/123)",
		"https://nhnent.dooray.com/project/pages/456":                         "[https://nhnent.dooray.com/project/pages/456](dooray://" + orgID + "/pages/456)",
		"see [link](https://nhnent.dooray.com/project/tasks/123) for details": "see [link](https://nhnent.dooray.com/project/tasks/123) for details",
		"no urls here": "no urls here",
		"a https://nhnent.dooray.com/project/tasks/1 b https://nhnent.dooray.com/project/tasks/2": "a [https://nhnent.dooray.com/project/tasks/1](dooray://" + orgID + "/tasks/1) b [https://nhnent.dooray.com/project/tasks/2](dooray://" + orgID + "/tasks/2)",
	}
	for in, want := range cases {
		if got := linkifyContent(orgID, in); got != want {
			t.Errorf("linkifyContent(%q) = %q, want %q", in, got, want)
		}
	}
}
