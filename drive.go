package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const doorayDriveBaseURL = "https://api.dooray.com/drive/v1/drives"

// DriveTools registers the dooray_drive tool. dooray-sdk (v0.4.1) does not cover
// drive endpoints, so handlers call the REST API directly via the helpers in
// project.go (getURL/postJSON not used here; upload uses a drive-specific
// multipart helper defined in this file).
//
// Read:
//	GET /drive/v1/drives?{queryString}                       list accessible drives
//	GET /drive/v1/drives/{driveId}/files?{queryString}        list folders/files in a drive
// Write:
//	POST /drive/v1/drives/{driveId}/files?parentId={parentId} upload a file or string content (multipart)
func DriveTools(s *server.MCPServer, token *string) {
	tool := mcp.NewTool("dooray_drive",
		mcp.WithDescription("list Dooray drives/folders and upload files. 'find_drives' (needs queryString, e.g. 'type=project&scope=public'), 'find_folders' (needs driveId, queryString, e.g. 'type=folder&subTypes=root'), 'upload_file' (needs driveId, parentId, filePath — uploads a local file), 'upload_content' (needs driveId, parentId, fileName, content — uploads a string as a file)."),
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
			mcp.Description("target folder id within the drive. Required for upload_file, upload_content."),
		),
		mcp.WithString("filePath",
			mcp.Description("absolute path to a local file. Required for upload_file."),
		),
		mcp.WithString("fileName",
			mcp.Description("file name to save as. Required for upload_content."),
		),
		mcp.WithString("content",
			mcp.Description("string content to upload as a file. Required for upload_content."),
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

		default:
			return mcp.NewToolResultError("unknown operation: " + op), nil
		}

		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(result), nil
	})
}
