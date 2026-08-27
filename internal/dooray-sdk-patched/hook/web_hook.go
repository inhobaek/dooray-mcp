package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dooray-go/dooray-sdk/utils"
	"io"
	"net/http"
)

type WebhookMessage struct {
	BotName      string       `json:"botName,omitempty"`
	BotIconImage string       `json:"botIconImage,omitempty"`
	Text         string       `json:"text,omitempty"`
	Attachments  []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Title     string `json:"title"`
	TitleLink string `json:"titleLink"`
	Text      string `json:"text"`
	Color     string `json:"color"`
}

var defaultHTTPClient = utils.NewDefaultHTTPClient()

func PostWebhook(url string, msg *WebhookMessage) error {
	return PostWebhookCustomHTTPContext(context.Background(), url, defaultHTTPClient, msg)
}

func PostWebhookContext(ctx context.Context, url string, msg *WebhookMessage) error {
	return PostWebhookCustomHTTPContext(ctx, url, defaultHTTPClient, msg)
}

func PostWebhookCustomHTTP(url string, httpClient *http.Client, msg *WebhookMessage) error {
	return PostWebhookCustomHTTPContext(context.Background(), url, httpClient, msg)
}

func PostWebhookCustomHTTPContext(ctx context.Context, url string, httpClient *http.Client, msg *WebhookMessage) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("failed new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post webhook: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	return checkStatusCode(resp)
}

func checkStatusCode(resp *http.Response) error {

	if resp.StatusCode != http.StatusOK {
		return utils.StatusCodeError{Code: resp.StatusCode, Status: resp.Status}
	}

	return nil
}
