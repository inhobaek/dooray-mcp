package messenger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	model "github.com/dooray-go/dooray-sdk/openapi/model/messenger"
	"io"
	"net/http"
)

type SendMessageRequest struct {
	Text string `json:"text"`
}

func (m Messenger) SendMessage(apikey string, channelID string, msg *SendMessageRequest) (*model.SendMessageResponse, error) {
	return m.SendMessageCustomHTTPContext(context.Background(), apikey, channelID, m.httpClient, msg)
}

func (m Messenger) SendMessageContext(ctx context.Context, apikey string, channelID string, msg *SendMessageRequest) (*model.SendMessageResponse, error) {
	return m.SendMessageCustomHTTPContext(ctx, apikey, channelID, m.httpClient, msg)
}

func (m Messenger) SendMessageCustomHTTP(apikey string, channelID string, httpClient *http.Client, msg *SendMessageRequest) (*model.SendMessageResponse, error) {
	return m.SendMessageCustomHTTPContext(context.Background(), apikey, channelID, httpClient, msg)
}

func (m Messenger) SendMessageCustomHTTPContext(ctx context.Context, apikey string, channelID string, httpClient *http.Client, msg *SendMessageRequest) (*model.SendMessageResponse, error) {
	raw, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}

	url := fmt.Sprintf("%s/messenger/v1/channels/%s/logs", m.endPoint, channelID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("failed new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("dooray-api %s", apikey))

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if err := checkStatusCode(resp); err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response model.SendMessageResponse
	if err := json.Unmarshal(resBody, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	response.RawJSON = string(resBody)

	return &response, nil
}

