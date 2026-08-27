package project

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var testResponse = `{
    "header": {
        "isSuccessful": true,
        "resultCode": 0,
        "resultMessage": ""
    },
    "result": [{
       "id": "",
       "subject": "",
       "project": {
           "id": "",
           "code": ""
       },
       "taskNumber": "",
       "closed": false,
       "createdAt": "",
       "dueDate": "",
       "dueDateFlag": false,
       "updatedAt": "",
       "number": 1,
       "priority": "",
       "parent": {
           "id": "",
           "number": "",
           "subject": ""
       },
       "workflowClass": "working",
       "milestone": {
           "id": "",
           "name": ""
       },
       "tags": [{
           "id": ""
       }],
       "users": {
           "from": {
               "type": "member",
               "member": {
                   "organizationmemberid": ""
               }
           },
           "to": [{
               "type": "member",

               "member": {
                   "organizationMemberId": ""
               },
               "workflow": {
                 "id": "1",
                 "name": "등록"
               }
           },{
               "type": "emailUser",
               "emailUser": {
                   "emailAddress": "",
                   "name": ""
               },
               "workflow": {
                 "id": "1",
                 "name": "등록"
               }
           }],
           "cc": [{
               "type": "group",
               "group": {
                   "projectMemberGroupId": "",
                   "members": [{
                       "organizationMemberId": ""
                   }]
               }
           }]
       },
      "workflow": {
        "id": "",
        "name": ""
      }
    }],
    "totalCount": 10
}`

func TestProject_GetPosts(t *testing.T) {
	projectId := "1234567890"
	tomemberIds := "1234567890,0987654321"

	var actualProjectId string
	var actualToMemberIds string

	mux := http.NewServeMux()
	mux.HandleFunc("/project/v1/projects/{projectId}/posts", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		actualProjectId = r.PathValue("projectId")

		query := r.URL.Query()
		actualToMemberIds = query.Get("toMemberIds")

		response := []byte(testResponse)
		rw.Write(response)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	actualResponse, err := NewProject(server.URL).GetPosts("dooray-api-key", projectId, tomemberIds, "")
	if err != nil {
		t.Errorf("Expected not to receive error: %s", err)
	}

	if !reflect.DeepEqual(testResponse, actualResponse.RawJSON) {
		t.Errorf("Response did not match\nwant: %#v\n got: %#v", projectId, actualProjectId)
	}

	if !reflect.DeepEqual(projectId, actualProjectId) {
		t.Errorf("projeceId did not match\nwant: %#v\n got: %#v", projectId, actualProjectId)
	}

	if !reflect.DeepEqual(tomemberIds, actualToMemberIds) {
		t.Errorf("tomemberIds did not match\nwant: %#v\n got: %#v", tomemberIds, actualToMemberIds)
	}
}

func TestProject_GetPostsWithOptions(t *testing.T) {
	projectId := "1234567890"

	var capturedQuery map[string]string

	mux := http.NewServeMux()
	mux.HandleFunc("/project/v1/projects/{projectId}/posts", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()
		capturedQuery = make(map[string]string)
		for key := range query {
			capturedQuery[key] = query.Get(key)
		}

		rw.Write([]byte(testResponse))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	intPtr := func(v int) *int { return &v }

	opts := GetPostsOptions{
		Page:                intPtr(1),
		Size:                intPtr(50),
		FromEmailAddress:    "test@example.com",
		FromMemberIds:       "member1",
		ToMemberIds:         "member2,member3",
		ToMemberSize:        intPtr(1),
		CcMemberIds:         "cc1",
		TagIds:              "tag1,tag2",
		ParentPostId:        "parent1",
		PostNumber:          "42",
		PostWorkflowClasses: "registered,working",
		PostWorkflowIds:     "wf1,wf2",
		MilestoneIds:        "ms1",
		Subjects:            "test subject",
		CreatedAt:           "today",
		UpdatedAt:           "thisweek",
		DueAt:               "next-7d",
		Order:               "-createdAt",
	}

	_, err := NewProject(server.URL).GetPostsWithOptions("dooray-api-key", projectId, opts)
	if err != nil {
		t.Fatalf("Expected not to receive error: %s", err)
	}

	expected := map[string]string{
		"page":                "1",
		"size":                "50",
		"fromEmailAddress":    "test@example.com",
		"fromMemberIds":       "member1",
		"toMemberIds":         "member2,member3",
		"toMemberSize":        "1",
		"ccMemberIds":         "cc1",
		"tagIds":              "tag1,tag2",
		"parentPostId":        "parent1",
		"postNumber":          "42",
		"postWorkflowClasses": "registered,working",
		"postWorkflowIds":     "wf1,wf2",
		"milestoneIds":        "ms1",
		"subjects":            "test subject",
		"createdAt":           "today",
		"updatedAt":           "thisweek",
		"dueAt":               "next-7d",
		"order":               "-createdAt",
	}

	for key, want := range expected {
		got, ok := capturedQuery[key]
		if !ok {
			t.Errorf("missing query parameter %q", key)
			continue
		}
		if got != want {
			t.Errorf("query parameter %q: want %q, got %q", key, want, got)
		}
	}
}

func TestProject_GetPostsWithOptions_EmptyOptions(t *testing.T) {
	projectId := "1234567890"
	var queryString string

	mux := http.NewServeMux()
	mux.HandleFunc("/project/v1/projects/{projectId}/posts", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		queryString = r.URL.RawQuery
		rw.Write([]byte(testResponse))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	_, err := NewProject(server.URL).GetPostsWithOptions("dooray-api-key", projectId, GetPostsOptions{})
	if err != nil {
		t.Fatalf("Expected not to receive error: %s", err)
	}

	if queryString != "" {
		t.Errorf("expected empty query string for empty options, got %q", queryString)
	}
}