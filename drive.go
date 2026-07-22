package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const doorayDriveBaseURL = "https://api.dooray.com/drive/v1/drives"

// DriveTools registers the dooray_drive tool. dooray-sdk (v0.4.1) does not cover
// drive endpoints, so handlers call the REST API directly via the helpers in
// project.go (getURL/postJSON not used here).
//
// Read:
//	GET /drive/v1/drives?{queryString}                       list accessible drives
//	GET /drive/v1/drives/{driveId}/files?{queryString}        list folders/files in a drive
//
// Write:
//	POST /drive/v1/drives/{driveId}/files?parentId={parentId}  upload file/content
func DriveTools(s *server.MCPServer, token *string) {
	tool := mcp.NewTool("dooray_drive",
		mcp.WithDescription("list, find, and upload files in Dooray drives. Operations: 'find_drives' (list accessible drives; needs queryString), 'find_folders' (list drive contents; needs driveId, queryString), 'upload_file' (upload file from disk; needs driveId, parentId, filePath), 'upload_content' (upload string content; needs driveId, parentId, fileName, content)."),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("The operation to perform."),
			mcp.Enum("find_drives", "find_folders", "upload_file", "upload_content"),
		),
		mcp.WithString("queryString",
			mcp.Description("query string for find_drives/find_folders, e.g. 'type=project&scope=public' or 'type=folder&subTypes=root'."),
		),
		mcp.WithString("driveId",
			mcp.Description("drive id. Required for find_folders, upload_file, upload_content."),
		),
		mcp.WithString("parentId",
			mcp.Description("parent folder id. Required for upload_file, upload_content."),
		),
		mcp.WithString("filePath",
			mcp.Description("path to file on disk. Required for upload_file."),
		),
		mcp.WithString("fileName",
			mcp.Description("name of the file. Required for upload_content."),
		),
		mcp.WithString("content",
			mcp.Description("file content as string. Required for upload_content."),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		op, _ := args["operation"].(string)
		str := func(k string) string { v, _ := args[k].(string); return v }

		var result string
		var err error

		switch op {
		case "find_drives":
			queryString := str("queryString")
			if queryString == "" {
				return mcp.NewToolResultError("queryString is required for find_drives"), nil
			}
			result, err = getURL(ctx, *token, fmt.Sprintf("%s?%s", doorayDriveBaseURL, queryString))

		case "find_folders":
			driveId, queryString := str("driveId"), str("queryString")
			if driveId == "" || queryString == "" {
				return mcp.NewToolResultError("driveId and queryString are required for find_folders"), nil
			}
			result, err = getURL(ctx, *token, fmt.Sprintf("%s/%s/files?%s", doorayDriveBaseURL, url.PathEscape(driveId), queryString))

		case "upload_file":
			driveId, parentId, filePath := str("driveId"), str("parentId"), str("filePath")
			if driveId == "" || parentId == "" || filePath == "" {
				return mcp.NewToolResultError("driveId, parentId, filePath are required for upload_file"), nil
			}
			f, ferr := os.Open(filePath)
			if ferr != nil {
				return nil, ferr
			}
			defer f.Close()
			uploadURL := fmt.Sprintf("%s/%s/files?parentId=%s", doorayDriveBaseURL, url.PathEscape(driveId), url.QueryEscape(parentId))
			result, err = driveUploadMultipart(ctx, *token, uploadURL, filepath.Base(filePath), f)

		case "upload_content":
			driveId, parentId, fileName, content := str("driveId"), str("parentId"), str("fileName"), str("content")
			if driveId == "" || parentId == "" || fileName == "" || content == "" {
				return mcp.NewToolResultError("driveId, parentId, fileName, content are required for upload_content"), nil
			}
			uploadURL := fmt.Sprintf("%s/%s/files?parentId=%s", doorayDriveBaseURL, url.PathEscape(driveId), url.QueryEscape(parentId))
			result, err = driveUploadMultipart(ctx, *token, uploadURL, fileName, strings.NewReader(content))

		default:
			return mcp.NewToolResultError("unknown operation: " + op), nil
		}

		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(result), nil
	})
}

// driveUploadMultipart POSTs file content (from any io.Reader) to a Dooray drive
// upload endpoint via multipart/form-data and returns the raw JSON response.
// Unlike project.go's uploadFileMultipart (used for post/wiki attachments), the
// drive upload API takes no "type" form field. The Dooray file API answers the
// first request with 307 + a file-api.dooray.com location; Go's default client
// strips the Authorization header on a cross-host redirect, so we disable
// auto-redirect and re-issue the request to the location ourselves.
func driveUploadMultipart(ctx context.Context, token, uploadURL, fileName string, r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	buildReq := func(u string) (*http.Request, error) {
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		fw, err := w.CreateFormFile("file", fileName)
		if err != nil {
			return nil, err
		}
		if _, err := fw.Write(data); err != nil {
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, &buf)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "dooray-api "+token)
		req.Header.Set("Content-Type", w.FormDataContentType())
		return req, nil
	}

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	reqURL := uploadURL
	for i := 0; i < 2; i++ { // at most one redirect hop
		req, err := buildReq(reqURL)
		if err != nil {
			return "", err
		}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", err
		}
		if resp.StatusCode == http.StatusTemporaryRedirect {
			loc := resp.Header.Get("Location")
			if loc == "" {
				return "", fmt.Errorf("drive upload redirect without Location header: status %d", resp.StatusCode)
			}
			reqURL = loc
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", fmt.Errorf("drive upload failed: status %d, body: %s", resp.StatusCode, string(body))
		}
		return string(body), nil
	}
	return "", fmt.Errorf("drive upload exceeded redirect limit")
}
