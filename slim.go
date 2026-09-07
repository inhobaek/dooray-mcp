package main

import "encoding/json"

// 응답 슬림화. Dooray 원본 JSON에서 LLM이 실제로 쓰는 필드만 남긴다.
// 실측(2026-09-07, 5세션): 채널 로그의 89%가 attachments(그중 titleLink 63%), 업무 목록의 53%가 users 그룹 멤버 나열이었다.
// 파싱 실패 시 원본을 그대로 돌려준다. raw=true 인자로 원본을 받을 수 있다.

func slimJSON(raw string, fn func(item map[string]any) map[string]any) string {
	var env map[string]any
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		return raw
	}
	items, ok := env["result"].([]any)
	if !ok {
		return raw
	}
	out := make([]any, 0, len(items))
	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		if s := fn(m); s != nil {
			out = append(out, s)
		}
	}
	env["result"] = out
	delete(env, "header")
	b, err := json.Marshal(env)
	if err != nil {
		return raw
	}
	return string(b)
}

func pick(m map[string]any, keys ...string) map[string]any {
	out := map[string]any{}
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			out[k] = v
		}
	}
	return out
}

func str(v any, path ...string) string {
	for _, p := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return ""
		}
		v = m[p]
	}
	s, _ := v.(string)
	return s
}

// 담당자 항목을 "이름" 또는 "그룹코드" 한 줄로. 그룹 멤버 나열은 버린다.
func userName(u any) string {
	if n := str(u, "member", "name"); n != "" {
		return n
	}
	if c := str(u, "group", "code"); c != "" {
		return c
	}
	return str(u, "member", "organizationMemberId")
}

func userNames(v any) []string {
	arr, _ := v.([]any)
	out := make([]string, 0, len(arr))
	for _, u := range arr {
		if n := userName(u); n != "" {
			out = append(out, n)
		}
	}
	return out
}

// slimChannelLogs: since(RFC3339, 문자열 비교로 충분)보다 오래된 메시지는 버린다.
// ponytail: sentAt 오프셋이 채널마다 다르면 비교가 어긋난다. 실측은 모두 +09:00.
func slimChannelLogs(raw, since string) string {
	return slimJSON(raw, func(m map[string]any) map[string]any {
		sentAt := str(m["sentAt"])
		if since != "" && sentAt < since {
			return nil
		}
		s := pick(m, "id", "type", "sentAt", "text")
		if n := str(m["customName"]); n != "" {
			s["sender"] = n
		} else {
			s["sender"] = userName(m["sender"])
		}
		if atts, ok := m["attachments"].([]any); ok && len(atts) > 0 {
			out := make([]any, 0, len(atts))
			for _, a := range atts {
				if am, ok := a.(map[string]any); ok {
					out = append(out, pick(am, "title", "text")) // titleLink은 Grafana URL이 전체의 63%라 raw=true에서만
				}
			}
			s["attachments"] = out
		}
		return s
	})
}

func slimPosts(raw string) string {
	return slimJSON(raw, func(m map[string]any) map[string]any {
		s := pick(m, "id", "number", "taskNumber", "subject", "closed", "priority",
			"createdAt", "updatedAt", "dueDate", "startedAt", "workflowClass", "workflow")
		s["projectId"] = str(m["project"], "id")
		if tags, ok := m["tags"].([]any); ok && len(tags) > 0 {
			ids := make([]string, 0, len(tags))
			for _, t := range tags {
				ids = append(ids, str(t, "id"))
			}
			s["tagIds"] = ids
		}
		if users, ok := m["users"].(map[string]any); ok {
			s["from"] = userName(users["from"])
			s["to"] = userNames(users["to"])
			if cc := userNames(users["cc"]); len(cc) > 0 {
				s["cc"] = cc
			}
		}
		return s
	})
}

func slimChannels(raw string) string {
	return slimJSON(raw, func(m map[string]any) map[string]any {
		s := pick(m, "id", "title", "type", "status", "updatedAt")
		if p, ok := m["users"].(map[string]any); ok {
			if arr, ok := p["participants"].([]any); ok {
				s["participantCount"] = len(arr)
			}
		}
		return s
	})
}

func slimProjects(raw string) string {
	return slimJSON(raw, func(m map[string]any) map[string]any {
		return pick(m, "id", "code", "name", "type", "scope", "state", "description")
	})
}
