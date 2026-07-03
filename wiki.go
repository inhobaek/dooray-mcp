package main

import (
	"context"
	"encoding/json"
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

// wikiBody is the {mimeType, content} body Dooray wiki pages/comments accept.
type wikiBody struct {
	MimeType string `json:"mimeType,omitempty"`
	Content  string `json:"content"`
}

// WikiTools registers the dooray_wiki tool. It reads Dooray Wiki pages and also
// creates/edits pages, uploads page files, and manages page comments. The
// dooray-sdk (v0.4.1) does not cover wiki endpoints, so handlers call the REST API
// via the helpers in project.go (getURL/postJSON/putJSON/deleteURL/uploadFileMultipart).
//
// Read:
//	GET  /wiki/v1/pages/{page-id}                              single page by id (no wikiId needed)
//	GET  /wiki/v1/wikis                                        list wikis
//	GET  /wiki/v1/wikis/{wiki-id}/pages                        list top-level pages
// Write (all need wikiId; get it from a get_page result's wikiId field):
//	POST /wiki/v1/wikis/{wiki-id}/pages                        create page
//	PUT  /wiki/v1/wikis/{wiki-id}/pages/{page-id}              update page (subject+body)
//	PUT  /wiki/v1/wikis/{wiki-id}/pages/{page-id}/content      update body only (append-friendly)
//	PUT  /wiki/v1/wikis/{wiki-id}/pages/{page-id}/title        update title only
//	POST /wiki/v1/wikis/{wiki-id}/pages/{page-id}/files        upload a file (multipart)
//	POST /wiki/v1/wikis/{wiki-id}/pages/{page-id}/comments     add comment
//	GET  /wiki/v1/wikis/{wiki-id}/pages/{page-id}/comments     list comments
//	PUT  /wiki/v1/wikis/{wiki-id}/pages/{page-id}/comments/{id} update comment
//	DELETE /wiki/v1/wikis/{wiki-id}/pages/{page-id}/comments/{id} delete comment
func WikiTools(s *server.MCPServer, token *string) {
	tool := mcp.NewTool("dooray_wiki",
		mcp.WithDescription("read and write Dooray Wiki pages. READ: 'get_page' (fetch full markdown body — pass a /project/pages/{id} URL or pageId; also returns wikiId), 'find_wikis' (list wikis), 'find_pages' (list a wiki's pages, needs wikiId). WRITE (need wikiId — get it from a get_page result): 'create_page' (needs wikiId, subject, content, parentPageId), 'update_page' (full page edit — needs wikiId, pageId, subject, content), 'update_content' (body only, best for appending — needs wikiId, pageId, content), 'update_title' (needs wikiId, pageId, subject), 'upload_file' (attach a local file, multipart — needs wikiId, pageId, filePath; returns insert markdown), 'create_comment' (needs wikiId, pageId, content), 'list_comments' (needs wikiId, pageId), 'update_comment' (needs wikiId, pageId, commentId, content), 'delete_comment' (needs wikiId, pageId, commentId). To append to a page: get_page for wikiId+current body, then update_content with old+new text."),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("The operation to perform."),
			mcp.Enum("get_page", "find_wikis", "find_pages",
				"create_page", "update_page", "update_content", "update_title",
				"upload_file", "create_comment", "list_comments", "update_comment", "delete_comment"),
		),
		mcp.WithString("page",
			mcp.Description("for get_page: a Dooray page URL or a bare pageId (last numeric segment is used)."),
		),
		mcp.WithString("wikiId",
			mcp.Description("wiki id. Required for find_pages and all write operations. Get it from find_wikis or a get_page result's 'wikiId' field."),
		),
		mcp.WithString("pageId",
			mcp.Description("wiki page id (a URL also works). Required for update_page/update_content/update_title/upload_file and all comment operations."),
		),
		mcp.WithString("subject",
			mcp.Description("page title. Required for create_page and update_title; optional for update_page."),
		),
		mcp.WithString("content",
			mcp.Description("markdown body. Required for create_page/update_content/create_comment/update_comment; optional for update_page."),
		),
		mcp.WithString("parentPageId",
			mcp.Description("parent page id for create_page (required by the API; a URL also works). Use a top-level page id to create directly under a wiki."),
		),
		mcp.WithString("commentId",
			mcp.Description("comment id. Required for update_comment and delete_comment."),
		),
		mcp.WithString("filePath",
			mcp.Description("absolute path to a local file for upload_file."),
		),
		mcp.WithString("fileType",
			mcp.Description("upload_file only: 'general' (default, attachment) or 'inline_image' (for embedding in the body)."),
		),
		mcp.WithNumber("page_num",
			mcp.Description("page number for find_pages/list_comments, default 0"),
		),
		mcp.WithNumber("size",
			mcp.Description("page size for find_pages (default 100) / list_comments (default 20, max 100)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		op, _ := args["operation"].(string)

		str := func(k string) string { v, _ := args[k].(string); return strings.TrimSpace(v) }
		numOr := func(k string, def int) int {
			if v, ok := args[k]; ok {
				return int(v.(float64))
			}
			return def
		}
		// wiki markdown body payload {"body":{"mimeType":..,"content":..}}
		bodyPayload := func(content string) []byte {
			p, _ := json.Marshal(map[string]any{"body": wikiBody{MimeType: "text/x-markdown", Content: content}})
			return p
		}

		var result string
		var err error

		switch op {
		case "get_page":
			pageId := pageIDFromInput(str("page"))
			if pageId == "" {
				return mcp.NewToolResultError("page is required for get_page (a page URL or pageId)"), nil
			}
			result, err = getURL(ctx, *token, fmt.Sprintf("%s/wiki/v1/pages/%s", doorayAPIEndpoint, url.PathEscape(pageId)))

		case "find_wikis":
			result, err = getURL(ctx, *token, fmt.Sprintf("%s/wiki/v1/wikis?size=100", doorayAPIEndpoint))

		case "find_pages":
			wikiId := str("wikiId")
			if wikiId == "" {
				return mcp.NewToolResultError("wikiId is required for find_pages"), nil
			}
			result, err = getURL(ctx, *token, fmt.Sprintf("%s/wiki/v1/wikis/%s/pages?page=%d&size=%d",
				doorayAPIEndpoint, url.PathEscape(wikiId), numOr("page_num", 0), numOr("size", 100)))

		case "create_page":
			wikiId, subject, content := str("wikiId"), str("subject"), str("content")
			parentPageId := pageIDFromInput(str("parentPageId"))
			if wikiId == "" || subject == "" || content == "" || parentPageId == "" {
				return mcp.NewToolResultError("wikiId, subject, content, parentPageId are required for create_page"), nil
			}
			payload, _ := json.Marshal(map[string]any{
				"subject":      subject,
				"body":         wikiBody{MimeType: "text/x-markdown", Content: content},
				"parentPageId": parentPageId,
			})
			result, err = postJSON(ctx, *token, fmt.Sprintf("%s/wiki/v1/wikis/%s/pages", doorayAPIEndpoint, url.PathEscape(wikiId)), payload)

		case "update_page":
			wikiId := str("wikiId")
			pageId := pageIDFromInput(str("pageId"))
			subject, content := str("subject"), str("content")
			if wikiId == "" || pageId == "" || (subject == "" && content == "") {
				return mcp.NewToolResultError("wikiId, pageId and at least one of subject/content are required for update_page"), nil
			}
			body := map[string]any{}
			if subject != "" {
				body["subject"] = subject
			}
			if content != "" {
				body["body"] = wikiBody{MimeType: "text/x-markdown", Content: content}
			}
			payload, _ := json.Marshal(body)
			result, err = putJSON(ctx, *token, fmt.Sprintf("%s/wiki/v1/wikis/%s/pages/%s", doorayAPIEndpoint, url.PathEscape(wikiId), url.PathEscape(pageId)), payload)

		case "update_content":
			wikiId := str("wikiId")
			pageId := pageIDFromInput(str("pageId"))
			content := str("content")
			if wikiId == "" || pageId == "" || content == "" {
				return mcp.NewToolResultError("wikiId, pageId, content are required for update_content"), nil
			}
			result, err = putJSON(ctx, *token, fmt.Sprintf("%s/wiki/v1/wikis/%s/pages/%s/content", doorayAPIEndpoint, url.PathEscape(wikiId), url.PathEscape(pageId)), bodyPayload(content))

		case "update_title":
			wikiId := str("wikiId")
			pageId := pageIDFromInput(str("pageId"))
			subject := str("subject")
			if wikiId == "" || pageId == "" || subject == "" {
				return mcp.NewToolResultError("wikiId, pageId, subject are required for update_title"), nil
			}
			payload, _ := json.Marshal(map[string]any{"subject": subject})
			result, err = putJSON(ctx, *token, fmt.Sprintf("%s/wiki/v1/wikis/%s/pages/%s/title", doorayAPIEndpoint, url.PathEscape(wikiId), url.PathEscape(pageId)), payload)

		case "upload_file":
			wikiId := str("wikiId")
			pageId := pageIDFromInput(str("pageId"))
			filePath := str("filePath")
			if wikiId == "" || pageId == "" || filePath == "" {
				return mcp.NewToolResultError("wikiId, pageId, filePath are required for upload_file"), nil
			}
			fileType := str("fileType")
			if fileType == "" {
				fileType = "general"
			}
			uploadURL := fmt.Sprintf("%s/wiki/v1/wikis/%s/pages/%s/files", doorayAPIEndpoint, url.PathEscape(wikiId), url.PathEscape(pageId))
			raw, uerr := uploadFileMultipart(ctx, *token, uploadURL, filePath, fileType)
			if uerr != nil {
				return nil, uerr
			}
			// Append the body-insert markdown: ![{name}](/wikis/{wikiId}/files/{attachFileId})
			result = raw + "\n" + wikiInsertMarkdown(raw, wikiId)

		case "create_comment":
			wikiId := str("wikiId")
			pageId := pageIDFromInput(str("pageId"))
			content := str("content")
			if wikiId == "" || pageId == "" || content == "" {
				return mcp.NewToolResultError("wikiId, pageId, content are required for create_comment"), nil
			}
			result, err = postJSON(ctx, *token, fmt.Sprintf("%s/wiki/v1/wikis/%s/pages/%s/comments", doorayAPIEndpoint, url.PathEscape(wikiId), url.PathEscape(pageId)), bodyPayload(content))

		case "list_comments":
			wikiId := str("wikiId")
			pageId := pageIDFromInput(str("pageId"))
			if wikiId == "" || pageId == "" {
				return mcp.NewToolResultError("wikiId, pageId are required for list_comments"), nil
			}
			result, err = getURL(ctx, *token, fmt.Sprintf("%s/wiki/v1/wikis/%s/pages/%s/comments?page=%d&size=%d",
				doorayAPIEndpoint, url.PathEscape(wikiId), url.PathEscape(pageId), numOr("page_num", 0), numOr("size", 20)))

		case "update_comment":
			wikiId := str("wikiId")
			pageId := pageIDFromInput(str("pageId"))
			commentId, content := str("commentId"), str("content")
			if wikiId == "" || pageId == "" || commentId == "" || content == "" {
				return mcp.NewToolResultError("wikiId, pageId, commentId, content are required for update_comment"), nil
			}
			result, err = putJSON(ctx, *token, fmt.Sprintf("%s/wiki/v1/wikis/%s/pages/%s/comments/%s", doorayAPIEndpoint, url.PathEscape(wikiId), url.PathEscape(pageId), url.PathEscape(commentId)), bodyPayload(content))

		case "delete_comment":
			wikiId := str("wikiId")
			pageId := pageIDFromInput(str("pageId"))
			commentId := str("commentId")
			if wikiId == "" || pageId == "" || commentId == "" {
				return mcp.NewToolResultError("wikiId, pageId, commentId are required for delete_comment"), nil
			}
			result, err = deleteURL(ctx, *token, fmt.Sprintf("%s/wiki/v1/wikis/%s/pages/%s/comments/%s", doorayAPIEndpoint, url.PathEscape(wikiId), url.PathEscape(pageId), url.PathEscape(commentId)))

		default:
			return mcp.NewToolResultError("unknown operation: " + op), nil
		}

		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(result), nil
	})
}

// wikiInsertMarkdown builds the body-insert markdown for an uploaded wiki file
// from the upload response JSON: ![{name}](/wikis/{wikiId}/files/{attachFileId}).
// Returns a hint line if the response can't be parsed.
func wikiInsertMarkdown(uploadJSON, wikiId string) string {
	var r struct {
		Result struct {
			Name         string `json:"name"`
			AttachFileID string `json:"attachFileId"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(uploadJSON), &r); err != nil || r.Result.AttachFileID == "" {
		return "(본문 삽입 markdown 생성 실패: 응답의 attachFileId를 확인하세요)"
	}
	return fmt.Sprintf("본문 삽입: ![%s](/wikis/%s/files/%s)", r.Result.Name, wikiId, r.Result.AttachFileID)
}
