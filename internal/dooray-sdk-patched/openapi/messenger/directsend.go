package messenger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	model "github.com/dooray-go/dooray-sdk/openapi/model/messenger"
	"github.com/dooray-go/dooray-sdk/utils"
	"io"
	"net/http"
)

type DirectSendRequest struct {
	Text                 string `json:"text"`
	OrganizationMemberId string `json:"organizationMemberId"`
}

func (m Messenger) DirectSend(apikey string, msg *DirectSendRequest) (*model.DirectSendResponse, error) {
	return m.DirectSendCustomHTTPContext(context.Background(), apikey, m.httpClient, msg)
}

func (m Messenger) DirectSendContext(ctx context.Context, apikey string, msg *DirectSendRequest) (*model.DirectSendResponse, error) {
	return m.DirectSendCustomHTTPContext(ctx, apikey, m.httpClient, msg)
}

func (m Messenger) DirectSendCustomHTTP(apikey string, httpClient *http.Client, msg *DirectSendRequest) (*model.DirectSendResponse, error) {
	return m.DirectSendCustomHTTPContext(context.Background(), apikey, httpClient, msg)
}

func (m Messenger) DirectSendCustomHTTPContext(ctx context.Context, apikey string, httpClient *http.Client, msg *DirectSendRequest) (*model.DirectSendResponse, error) {
	raw, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}

	url := m.endPoint + "/messenger/v1/channels/direct-send"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("failed new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("dooray-api %s", apikey))

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to post webhook: %w", err)
	}
	defer resp.Body.Close()

	if err := checkStatusCode(resp); err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response model.DirectSendResponse
	if err := json.Unmarshal(resBody, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	response.RawJSON = string(resBody)

	return &response, nil
}

func checkStatusCode(resp *http.Response) error {

	if resp.StatusCode != http.StatusOK {
		return utils.StatusCodeError{Code: resp.StatusCode, Status: resp.Status}
	}

	return nil
}
