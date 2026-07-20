package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	model "github.com/dooray-go/dooray-sdk/openapi/model/project"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// templateRecipient is a to/cc entry on a template. Like posts, a recipient is
// either a member (organizationMemberId) or a group (projectMemberGroupId).
type templateRecipient struct {
	Type   string             `json:"type"`             // "member" | "group"
	Member *model.PostMember  `json:"member,omitempty"` // type=member
	Group  *templateGroupRef  `json:"group,omitempty"`  // type=group
}

type templateGroupRef struct {
	ProjectMemberGroupID string `json:"projectMemberGroupId"`
}

type templateUsers struct {
	To []templateRecipient `json:"to,omitempty"`
	Cc []templateRecipient `json:"cc,omitempty"`
}

type templateTagRef struct {
	ID string `json:"id"`
}

// templateRequest is the POST/PUT body for a work template. It mirrors the shape
// returned by GET .../templates/{id}. PUT is full-replacement, so all fields the
// caller wants to keep must be resent (fetch current via get_template first).
type templateRequest struct {
	TemplateName string                 `json:"templateName"`
	Subject      string                 `json:"subject"`
	Body         model.PostBody         `json:"body"`
	Guide        *model.PostBody        `json:"guide,omitempty"`
	TagIDs       []templateTagRef       `json:"tags,omitempty"`
	Users        *templateUsers         `json:"users,omitempty"`
	DueDateFlag  bool                   `json:"dueDateFlag"`
	IsDefault    bool                   `json:"isDefault"`
	Priority     string                 `json:"priority,omitempty"`
}

// parseRecipients turns a comma-separated id list into member-type recipients.
func parseMemberRecipients(csv string) []templateRecipient {
	var out []templateRecipient
	for _, id := range strings.Split(csv, ",") {
		if id = strings.TrimSpace(id); id != "" {
			out = append(out, templateRecipient{
				Type:   "member",
				Member: &model.PostMember{OrganizationMemberID: id},
			})
		}
	}
	return out
}

// parseGroupRecipients turns a comma-separated group id list into group-type recipients.
func parseGroupRecipients(csv string) []templateRecipient {
	var out []templateRecipient
	for _, id := range strings.Split(csv, ",") {
		if id = strings.TrimSpace(id); id != "" {
			out = append(out, templateRecipient{
				Type:  "group",
				Group: &templateGroupRef{ProjectMemberGroupID: id},
			})
		}
	}
	return out
}

// buildTemplateRequest assembles a templateRequest from the tool arguments,
// shared by create_template and update_template.
func buildTemplateRequest(args map[string]any) (templateRequest, error) {
	templateName, _ := args["templateName"].(string)
	subject, _ := args["subject"].(string)
	bodyContent, _ := args["bodyContent"].(string)
	if templateName == "" || subject == "" || bodyContent == "" {
		return templateRequest{}, fmt.Errorf("templateName, subject, bodyContent are required")
	}

	bodyMime, _ := args["bodyMimeType"].(string)
	if bodyMime == "" {
		bodyMime = "text/x-markdown"
	}

	req := templateRequest{
		TemplateName: templateName,
		Subject:      subject,
		Body:         model.PostBody{MimeType: bodyMime, Content: bodyContent},
		DueDateFlag:  true, // observed default on every existing template
	}

	if guide, _ := args["guideContent"].(string); guide != "" {
		gm, _ := args["guideMimeType"].(string)
		if gm == "" {
			gm = "text/x-markdown"
		}
		req.Guide = &model.PostBody{MimeType: gm, Content: guide}
	}
	if v, ok := args["dueDateFlag"].(bool); ok {
		req.DueDateFlag = v
	}
	if v, ok := args["isDefault"].(bool); ok {
		req.IsDefault = v
	}
	if v, _ := args["priority"].(string); v != "" {
		req.Priority = v
	}
	if v, _ := args["tagIds"].(string); v != "" {
		for _, id := range strings.Split(v, ",") {
			if id = strings.TrimSpace(id); id != "" {
				req.TagIDs = append(req.TagIDs, templateTagRef{ID: id})
			}
		}
	}

	users := &templateUsers{}
	hasUsers := false
	if v, _ := args["toMemberIds"].(string); v != "" {
		users.To = append(users.To, parseMemberRecipients(v)...)
		hasUsers = true
	}
	if v, _ := args["toGroupIds"].(string); v != "" {
		users.To = append(users.To, parseGroupRecipients(v)...)
		hasUsers = true
	}
	if v, _ := args["ccMemberIds"].(string); v != "" {
		users.Cc = append(users.Cc, parseMemberRecipients(v)...)
		hasUsers = true
	}
	if v, _ := args["ccGroupIds"].(string); v != "" {
		users.Cc = append(users.Cc, parseGroupRecipients(v)...)
		hasUsers = true
	}
	if hasUsers {
		req.Users = users
	}

	return req, nil
}

// TemplateTools registers the dooray_template tool, which manages a project's
// work templates (업무 양식). The dooray-sdk (v0.4.1) does not cover the
// templates endpoints, so the handler calls the Dooray REST API directly via the
// HTTP helpers defined in project.go (same approach as get_logs / set_workflow).
//
//	GET  /project/v1/projects/{project-id}/templates           list templates
//	GET  /project/v1/projects/{project-id}/templates/{id}      single template (full body)
//	POST   /project/v1/projects/{project-id}/templates           create template
//	PUT    /project/v1/projects/{project-id}/templates/{id}      update template (full-replacement)
//	DELETE /project/v1/projects/{project-id}/templates/{id}      delete template
func TemplateTools(s *server.MCPServer, token *string) {
	tool := mcp.NewTool("dooray_template",
		mcp.WithDescription("read and manage a project's work templates (업무 양식). 'get_templates': list templates. 'get_template': single template with full body/guide. 'create_template': create a new template. 'update_template': FULL-REPLACEMENT edit of a template (resend every field; fetch current via get_template first or fields are cleared). 'delete_template': delete a template (requires templateId — irreversible)."),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("The operation to perform. 'get_templates' (requires projectId). 'get_template' (requires projectId, templateId). 'create_template' (requires projectId, templateName, subject, bodyContent). 'update_template' (requires projectId, templateId, templateName, subject, bodyContent — PUT is full-replacement). 'delete_template' (requires projectId, templateId — irreversible)."),
			mcp.Enum("get_templates", "get_template", "create_template", "update_template", "delete_template"),
		),
		mcp.WithString("projectId",
			mcp.Required(),
			mcp.Description("project id. it can be obtained from the find_projects tool."),
		),
		mcp.WithString("templateId",
			mcp.Description("template id (required for get_template and update_template). it can be obtained from the get_templates tool."),
		),
		// paging (get_templates)
		mcp.WithNumber("page",
			mcp.Description("page number for get_templates, default is 0"),
		),
		mcp.WithNumber("size",
			mcp.Description("number of templates per page for get_templates, default is 20, max is 100"),
		),
		// write fields (create_template / update_template)
		mcp.WithString("templateName",
			mcp.Description("template display name (required for create/update_template). e.g. '[DNS] DNS 추가/변경/삭제'"),
		),
		mcp.WithString("subject",
			mcp.Description("default post subject the template produces (required for create/update_template). may contain tokens like [${department}]."),
		),
		mcp.WithString("bodyContent",
			mcp.Description("template body content (required for create/update_template)."),
		),
		mcp.WithString("bodyMimeType",
			mcp.Description("body mime type: 'text/x-markdown' (default) or 'text/html'."),
		),
		mcp.WithString("guideContent",
			mcp.Description("optional guide (작성 가이드) content shown beside the template."),
		),
		mcp.WithString("guideMimeType",
			mcp.Description("guide mime type: 'text/x-markdown' (default) or 'text/html'."),
		),
		mcp.WithString("tagIds",
			mcp.Description("tag ids to attach, comma separated."),
		),
		mcp.WithString("toMemberIds",
			mcp.Description("default assignee organizationMemberIds, comma separated (type=member)."),
		),
		mcp.WithString("toGroupIds",
			mcp.Description("default assignee projectMemberGroupIds, comma separated (type=group)."),
		),
		mcp.WithString("ccMemberIds",
			mcp.Description("default cc organizationMemberIds, comma separated (type=member)."),
		),
		mcp.WithString("ccGroupIds",
			mcp.Description("default cc projectMemberGroupIds, comma separated (type=group)."),
		),
		mcp.WithBoolean("dueDateFlag",
			mcp.Description("whether the template enables a due date. default true (matches existing templates)."),
		),
		mcp.WithBoolean("isDefault",
			mcp.Description("whether this is the project's default template. default false."),
		),
		mcp.WithString("priority",
			mcp.Description("default priority: none | low | normal | high | urgent. default none."),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		op, _ := args["operation"].(string)
		projectId, _ := args["projectId"].(string)
		if projectId == "" {
			return mcp.NewToolResultError("projectId is required"), nil
		}

		var result string
		switch op {
		case "get_templates":
			page := 0
			size := 20
			if v, ok := args["page"]; ok {
				page = int(v.(float64))
			}
			if v, ok := args["size"]; ok {
				size = int(v.(float64))
			}
			reqURL := fmt.Sprintf("%s/project/v1/projects/%s/templates?page=%d&size=%d", doorayAPIEndpoint, url.PathEscape(projectId), page, size)
			res, err := getURL(ctx, *token, reqURL)
			if err != nil {
				return nil, err
			}
			result = res
		case "get_template":
			templateId, _ := args["templateId"].(string)
			if templateId == "" {
				return mcp.NewToolResultError("templateId is required for get_template"), nil
			}
			reqURL := fmt.Sprintf("%s/project/v1/projects/%s/templates/%s", doorayAPIEndpoint, url.PathEscape(projectId), url.PathEscape(templateId))
			res, err := getURL(ctx, *token, reqURL)
			if err != nil {
				return nil, err
			}
			result = res
		case "create_template":
			body, err := buildTemplateRequest(args)
			if err != nil {
				return mcp.NewToolResultError(err.Error() + " for create_template"), nil
			}
			payload, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			reqURL := fmt.Sprintf("%s/project/v1/projects/%s/templates", doorayAPIEndpoint, url.PathEscape(projectId))
			res, err := postJSON(ctx, *token, reqURL, payload)
			if err != nil {
				return nil, err
			}
			result = res
		case "update_template":
			templateId, _ := args["templateId"].(string)
			if templateId == "" {
				return mcp.NewToolResultError("templateId is required for update_template"), nil
			}
			body, err := buildTemplateRequest(args)
			if err != nil {
				return mcp.NewToolResultError(err.Error() + " for update_template (PUT is full-replacement; resend every field fetched via get_template or they are cleared)"), nil
			}
			payload, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			reqURL := fmt.Sprintf("%s/project/v1/projects/%s/templates/%s", doorayAPIEndpoint, url.PathEscape(projectId), url.PathEscape(templateId))
			res, err := putJSON(ctx, *token, reqURL, payload)
			if err != nil {
				return nil, err
			}
			result = res
		case "delete_template":
			templateId, _ := args["templateId"].(string)
			if templateId == "" {
				return mcp.NewToolResultError("templateId is required for delete_template"), nil
			}
			reqURL := fmt.Sprintf("%s/project/v1/projects/%s/templates/%s", doorayAPIEndpoint, url.PathEscape(projectId), url.PathEscape(templateId))
			res, err := deleteURL(ctx, *token, reqURL)
			if err != nil {
				return nil, err
			}
			result = res
		default:
			return mcp.NewToolResultError("unknown operation: " + op), nil
		}
		return mcp.NewToolResultText(result), nil
	})
}
