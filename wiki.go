package main

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// pageIDFromInput accepts either a raw wiki page id or a Dooray page URL and
// returns the page id. Dooray page URLs look like:
//
//	https://nhnent.dooray.com/project/pages/3205786621378362771
//	dooray://1387695619080878080/pages/3205786621378362771 "publish"
//
// so the page id is the last run of digits in the string.
func pageIDFromInput(s string) string {
	s = strings.TrimSpace(s)
	if !strings.ContainsAny(s, "/:") {
		return s // already a bare id
	}
	m := regexp.MustCompile(`(\d{6,})`).FindAllString(s, -1)
	if len(m) == 0 {
		return s
	}
	return m[len(m)-1] // last numeric segment is the pageId
}

// WikiTools registers the dooray_wiki tool, which reads Dooray Wiki pages.
// The /project/pages/{id} share URL exposes only a pageId, and the wiki API
// resolves a page by id alone (no wikiId needed), so a page URL converts to its
// markdown content directly. The dooray-sdk (v0.4.1) does not cover the wiki
// endpoints, so the handler calls the REST API via the helpers in project.go.
//
//	GET /wiki/v1/pages/{page-id}                 single page by id (full body) — works without wikiId
//	GET /wiki/v1/wikis                           list wikis (id, name, project, home pageId)
//	GET /wiki/v1/wikis/{wiki-id}/pages           list top-level pages of a wiki
//	GET /wiki/v1/wikis/{wiki-id}/pages/{page-id} single page via wiki path
func WikiTools(s *server.MCPServer, token *string) {
	tool := mcp.NewTool("dooray_wiki",
		mcp.WithDescription("read Dooray Wiki pages. 'get_page': fetch a single wiki page with its full markdown body — pass a /project/pages/{id} share URL (or a bare pageId) as 'page'; wikiId is NOT required (the page id resolves on its own). 'find_wikis': list wikis the token can see (id, name, project, home pageId). 'find_pages': list a wiki's top-level pages (requires wikiId)."),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("The operation to perform. 'get_page' (requires page — a page URL or pageId). 'find_wikis' (no args). 'find_pages' (requires wikiId)."),
			mcp.Enum("get_page", "find_wikis", "find_pages"),
		),
		mcp.WithString("page",
			mcp.Description("for get_page: a Dooray page URL (e.g. https://nhnent.dooray.com/project/pages/3205786621378362771) or a bare pageId. The pageId is extracted as the last numeric segment, so full URLs work as-is."),
		),
		mcp.WithString("wikiId",
			mcp.Description("wiki id (required for find_pages). Obtain it from find_wikis, or from a get_page result's 'wikiId' field."),
		),
		mcp.WithNumber("page_num",
			mcp.Description("page number for find_pages, default 0"),
		),
		mcp.WithNumber("size",
			mcp.Description("number of pages per page for find_pages, default 100, max 100"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		op, _ := args["operation"].(string)

		var result string
		switch op {
		case "get_page":
			raw, _ := args["page"].(string)
			pageId := pageIDFromInput(raw)
			if pageId == "" {
				return mcp.NewToolResultError("page is required for get_page (a page URL or pageId)"), nil
			}
			reqURL := fmt.Sprintf("%s/wiki/v1/pages/%s", doorayAPIEndpoint, url.PathEscape(pageId))
			res, err := getURL(ctx, *token, reqURL)
			if err != nil {
				return nil, err
			}
			result = res
		case "find_wikis":
			reqURL := fmt.Sprintf("%s/wiki/v1/wikis?size=100", doorayAPIEndpoint)
			res, err := getURL(ctx, *token, reqURL)
			if err != nil {
				return nil, err
			}
			result = res
		case "find_pages":
			wikiId, _ := args["wikiId"].(string)
			if wikiId == "" {
				return mcp.NewToolResultError("wikiId is required for find_pages (get it from find_wikis or a get_page result)"), nil
			}
			page := 0
			size := 100
			if v, ok := args["page_num"]; ok {
				page = int(v.(float64))
			}
			if v, ok := args["size"]; ok {
				size = int(v.(float64))
			}
			reqURL := fmt.Sprintf("%s/wiki/v1/wikis/%s/pages?page=%d&size=%d", doorayAPIEndpoint, url.PathEscape(wikiId), page, size)
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
