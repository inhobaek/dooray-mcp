package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSlimChannelLogs(t *testing.T) {
	raw := `{"header":{"resultCode":0},"result":[
	 {"id":"1","seq":5,"type":"WEBHOOK","sender":{"type":"member","member":{"organizationMemberId":"9"}},"customName":"Endpoint","customIconUrl":"http://x/pepe.png","sentAt":"2026-09-05T01:00:00+09:00","text":"hi","attachments":[{"color":"orange","title":"T","titleLink":"http://l","text":"body"}],"flags":{}},
	 {"id":"2","type":"MEMBER","sender":{"type":"member","member":{"organizationMemberId":"8","name":"백인호"}},"sentAt":"2026-09-01T01:00:00+09:00","text":"old"},
	 {"id":"3","type":"NORMAL","sender":{"type":"member","member":{"organizationMemberId":"7"}},"sentAt":"2026-09-01T02:00:00+09:00","text":"nameless"}]}`
	out := slimChannelLogs(raw, "2026-09-04T20:00:00+09:00", nil)
	var env map[string]any
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatal(err)
	}
	res := env["result"].([]any)
	if len(res) != 1 {
		t.Fatalf("since filter: want 1, got %d", len(res))
	}
	m := res[0].(map[string]any)
	if m["sender"] != "Endpoint" || m["text"] != "hi" {
		t.Fatalf("bad slim: %v", m)
	}
	if strings.Contains(out, "customIconUrl") || strings.Contains(out, "titleLink") || strings.Contains(out, "header") {
		t.Fatalf("noise kept: %s", out)
	}
	all := slimChannelLogs(raw, "", func(id string) string { return "N" + id })
	if !strings.Contains(all, `"sender":"N7"`) || !strings.Contains(all, `"sender":"백인호"`) {
		t.Fatalf("resolver should fill only nameless senders: %s", all)
	}
	if slimChannelLogs(raw, "", nil)[0] != '{' || len(slimJSON("not json", nil)) != 8 {
		t.Fatal("fallback")
	}
}

func TestSlimPosts(t *testing.T) {
	raw := `{"header":{},"totalCount":1,"result":[{"id":"p1","number":7,"taskNumber":"X/7","subject":"s","closed":false,"priority":"none","createdAt":"c","updatedAt":"u","dueDate":null,"workflowClass":"registered","workflow":{"id":"w","name":"대기"},"project":{"id":"proj"},"tags":[{"id":"t1"}],"fileIdList":[],
	 "users":{"from":{"type":"member","member":{"organizationMemberId":"1","name":"A"}},"to":[{"type":"group","group":{"code":"G/eng","members":[{"name":"B"},{"name":"C"}]}}],"cc":[{"type":"member","member":{"name":"D"}}]}}]}`
	out := slimPosts(raw)
	var env map[string]any
	json.Unmarshal([]byte(out), &env)
	m := env["result"].([]any)[0].(map[string]any)
	if m["from"] != "A" || m["to"].([]any)[0] != "G/eng" || m["cc"].([]any)[0] != "D" || m["projectId"] != "proj" || m["tagIds"].([]any)[0] != "t1" {
		t.Fatalf("bad slim: %s", out)
	}
	if env["totalCount"].(float64) != 1 || strings.Contains(out, "members") || strings.Contains(out, "fileIdList") {
		t.Fatalf("noise kept: %s", out)
	}
}

func TestSlimChannelsProjects(t *testing.T) {
	ch := slimChannels(`{"result":[{"id":"c","title":"t","type":"private","status":"normal","users":{"participants":[{},{}]},"organization":{"id":"o"}}]}`)
	if !strings.Contains(ch, `"participantCount":2`) || strings.Contains(ch, "organization") {
		t.Fatal(ch)
	}
	pr := slimProjects(`{"result":[{"id":"1","code":"P","name":"N","organization":{"id":"o"},"state":"active"}]}`)
	if strings.Contains(pr, "organization") || !strings.Contains(pr, `"code":"P"`) {
		t.Fatal(pr)
	}
}

func TestSlimLogsAndRateLimit(t *testing.T) {
	raw := `{"result":[{"id":"a","createdAt":"2026-09-05T10:00:00+09:00","body":{"content":"x"}},{"id":"b","createdAt":"2026-09-01T10:00:00+09:00"}]}`
	out := slimLogs(raw, "2026-09-04T00:00:00+09:00")
	if !strings.Contains(out, `"id":"a"`) || strings.Contains(out, `"id":"b"`) || !strings.Contains(out, `"content":"x"`) {
		t.Fatal(out)
	}
	if !emptyDespiteTotal(`{"totalCount":5,"result":[]}`) || emptyDespiteTotal(`{"totalCount":0,"result":[]}`) || emptyDespiteTotal(`{"totalCount":1,"result":[{}]}`) {
		t.Fatal("rate limit detection")
	}
}
