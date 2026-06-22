package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// TemplateTools registers the dooray_template tool, which exposes a project's
// work templates (업무 양식). The dooray-sdk (v0.4.1) does not cover the
// templates endpoints, so the handler calls the Dooray REST API directly via
// the getURL helper defined in project.go (same approach as get_logs).
//
//	GET /project/v1/projects/{project-id}/templates           list templates
//	GET /project/v1/projects/{project-id}/templates/{id}      single template (full body)
func TemplateTools(s *server.MCPServer, token *string) {
	tool := mcp.NewTool("dooray_template",
		mcp.WithDescription("read a project's work templates (업무 양식). 'get_templates': list all templates in a project. 'get_template': get a single template with full body/guide (requires templateId)."),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("The operation to perform. 'get_templates': list templates in a project (requires projectId). 'get_template': get a single template with full body (requires projectId, templateId)."),
			mcp.Enum("get_templates", "get_template"),
		),
		mcp.WithString("projectId",
			mcp.Required(),
			mcp.Description("project id. it can be obtained from the find_projects tool."),
		),
		mcp.WithString("templateId",
			mcp.Description("template id (required for get_template). it can be obtained from the get_templates tool."),
		),
		mcp.WithNumber("page",
			mcp.Description("page number for get_templates, default is 0"),
		),
		mcp.WithNumber("size",
			mcp.Description("number of templates per page for get_templates, default is 20, max is 100"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		op := request.GetArguments()["operation"].(string)
		projectId, _ := request.GetArguments()["projectId"].(string)
		if projectId == "" {
			return mcp.NewToolResultError("projectId is required"), nil
		}

		var result string
		switch op {
		case "get_templates":
			page := 0
			size := 20
			if v, ok := request.GetArguments()["page"]; ok {
				page = int(v.(float64))
			}
			if v, ok := request.GetArguments()["size"]; ok {
				size = int(v.(float64))
			}
			reqURL := fmt.Sprintf("%s/project/v1/projects/%s/templates?page=%d&size=%d", doorayAPIEndpoint, url.PathEscape(projectId), page, size)
			res, err := getURL(ctx, *token, reqURL)
			if err != nil {
				return nil, err
			}
			result = res
		case "get_template":
			templateId, _ := request.GetArguments()["templateId"].(string)
			if templateId == "" {
				return mcp.NewToolResultError("templateId is required for get_template"), nil
			}
			reqURL := fmt.Sprintf("%s/project/v1/projects/%s/templates/%s", doorayAPIEndpoint, url.PathEscape(projectId), url.PathEscape(templateId))
			res, err := getURL(ctx, *token, reqURL)
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
