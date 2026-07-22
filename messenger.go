package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dooray-go/dooray-sdk/openapi/messenger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const doorayMessengerBaseURL = "https://api.dooray.com"

func MessengerTools(s *server.MCPServer, token *string) {
	doorayMessengerTool := mcp.NewTool("dooray_messenger",
		mcp.WithDescription("send dooray messenger DM or channel message, list channels, or fetch channel logs"),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("The operation to perform. 'send': direct message to a member (requires to, message). 'send_channel': send message to a channel (requires channelId, message). 'reply_thread': reply in a thread under an existing channel message (requires channelId, parentMessageId, message). 'find_channels': list channels the current user belongs to. 'find_channel_logs': fetch recent messages of a channel (requires channelId). 'send_webhook': POST a message to a dooray messenger webhook URL (requires webhookUrl, message)"),
			mcp.Enum("send", "send_channel", "reply_thread", "find_channels", "find_channel_logs", "send_webhook"),
		),
		mcp.WithString("to",
			mcp.Description("recipient organizationMemberId (required for 'send')"),
		),
		mcp.WithString("message",
			mcp.Description("message to send (required for 'send', 'send_channel', and 'send_webhook')"),
		),
		mcp.WithString("channelId",
			mcp.Description("channel id (required for 'send_channel', 'reply_thread', and 'find_channel_logs')"),
		),
		mcp.WithString("parentMessageId",
			mcp.Description("id of the channel message to reply under (required for 'reply_thread'; this is the message log id returned by a prior 'send_channel')"),
		),
		mcp.WithNumber("limit",
			mcp.Description("max number of logs to fetch for 'find_channel_logs' (sent as 'size' query parameter, default 50)"),
		),
		mcp.WithString("webhookUrl",
			mcp.Description("dooray messenger webhook URL, e.g. https://example.dooray.com/services/... (required for 'send_webhook')"),
		),
	)

	s.AddTool(doorayMessengerTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		op := request.GetArguments()["operation"].(string)

		var result string
		switch op {
		case "send":
			to, _ := request.GetArguments()["to"].(string)
			message, _ := request.GetArguments()["message"].(string)
			if to == "" || message == "" {
				return mcp.NewToolResultError("'to' and 'message' are required for send"), nil
			}
			res, err := messenger.NewDefaultMessenger().DirectSend(*token,
				&messenger.DirectSendRequest{
					Text:                 message,
					OrganizationMemberId: to,
				})

			if err != nil {
				return nil, err
			}
			result = res.RawJSON
		case "send_channel":
			channelId, _ := request.GetArguments()["channelId"].(string)
			message, _ := request.GetArguments()["message"].(string)
			if channelId == "" || message == "" {
				return mcp.NewToolResultError("'channelId' and 'message' are required for send_channel"), nil
			}
			res, err := messenger.NewDefaultMessenger().SendMessageContext(ctx, *token, channelId,
				&messenger.SendMessageRequest{
					Text: message,
				})
			if err != nil {
				return nil, err
			}
			result = res.RawJSON
		case "reply_thread":
			channelId, _ := request.GetArguments()["channelId"].(string)
			parentMessageId, _ := request.GetArguments()["parentMessageId"].(string)
			message, _ := request.GetArguments()["message"].(string)
			if channelId == "" || parentMessageId == "" || message == "" {
				return mcp.NewToolResultError("'channelId', 'parentMessageId' and 'message' are required for reply_thread"), nil
			}
			body, err := replyThread(ctx, *token, channelId, parentMessageId, message)
			if err != nil {
				return nil, err
			}
			result = body
		case "find_channels":
			body, err := getChannels(ctx, *token)
			if err != nil {
				return nil, err
			}
			result = body
		case "find_channel_logs":
			channelId, _ := request.GetArguments()["channelId"].(string)
			if channelId == "" {
				return mcp.NewToolResultError("'channelId' is required for find_channel_logs"), nil
			}
			limit := 50
			if v, ok := request.GetArguments()["limit"]; ok {
				if f, ok := v.(float64); ok && f > 0 {
					limit = int(f)
				}
			}
			body, err := getChannelLogs(ctx, *token, channelId, limit)
			if err != nil {
				return nil, err
			}
			result = body
		case "send_webhook":
			webhookUrl, _ := request.GetArguments()["webhookUrl"].(string)
			message, _ := request.GetArguments()["message"].(string)
			if webhookUrl == "" || message == "" {
				return mcp.NewToolResultError("'webhookUrl' and 'message' are required for send_webhook"), nil
			}
			if verr := validateWebhookURL(webhookUrl); verr != nil {
				return mcp.NewToolResultError(verr.Error()), nil
			}
			body, err := sendWebhookMessage(ctx, webhookUrl, message)
			if err != nil {
				return nil, err
			}
			result = body
		}
		return mcp.NewToolResultText(fmt.Sprintf("%s", result)), nil
	})
}

func getChannels(ctx context.Context, token string) (string, error) {
	u := fmt.Sprintf("%s/messenger/v1/channels", doorayMessengerBaseURL)
	return doorayGet(ctx, token, u)
}

func getChannelLogs(ctx context.Context, token, channelId string, limit int) (string, error) {
	u, err := url.Parse(fmt.Sprintf("%s/messenger/v1/channels/%s/logs", doorayMessengerBaseURL, url.PathEscape(channelId)))
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("size", strconv.Itoa(limit))
	u.RawQuery = q.Encode()
	return doorayGet(ctx, token, u.String())
}

func replyThread(ctx context.Context, token, channelId, parentMessageId, message string) (string, error) {
	u := fmt.Sprintf("%s/messenger/v1/channels/%s/logs/%s/threads/create-and-send",
		doorayMessengerBaseURL, url.PathEscape(channelId), url.PathEscape(parentMessageId))
	payload, err := json.Marshal(map[string]string{"text": message})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("dooray-api %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("dooray POST %s failed: status=%d body=%s", u, resp.StatusCode, string(body))
	}
	return string(body), nil
}

func doorayGet(ctx context.Context, token, urlStr string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("dooray-api %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("dooray GET %s failed: status=%d body=%s", urlStr, resp.StatusCode, string(body))
	}
	return string(body), nil
}

// validateWebhookURL rejects any URL that is not an https URL on a *.dooray.com
// host, to prevent send_webhook being used as an SSRF vector to arbitrary hosts.
func validateWebhookURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}
	if u.Scheme != "https" || !strings.HasSuffix(u.Hostname(), ".dooray.com") {
		return fmt.Errorf("webhookUrl must be an https URL on *.dooray.com, got: %s", rawURL)
	}
	return nil
}

// sendWebhookMessage POSTs a JSON {"text": message} body to a Dooray webhook URL.
// Webhook URLs are self-authenticating (the URL path itself is the secret), so no
// Authorization header is sent.
func sendWebhookMessage(ctx context.Context, webhookUrl, message string) (string, error) {
	payload, err := json.Marshal(map[string]string{"text": message})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookUrl, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("webhook POST failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	return string(body), nil
}
