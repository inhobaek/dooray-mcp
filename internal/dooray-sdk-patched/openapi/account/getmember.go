package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	model "github.com/dooray-go/dooray-sdk/openapi/model/account" // 모델 패키지 임포트 추가
)

func (a *Account) GetMember(apikey string, id string) (*model.GetMemberResponse, error) { // 반환 타입 수정
	return a.GetMemberCustomHTTPContext(context.Background(), apikey, a.httpClient, id)
}

func (a *Account) GetMemberContext(ctx context.Context, apikey string, id string) (*model.GetMemberResponse, error) { // 반환 타입 수정
	return a.GetMemberCustomHTTPContext(ctx, apikey, a.httpClient, id)
}

func (a *Account) GetMemberCustomHTTP(apikey string, httpClient *http.Client, id string) (*model.GetMemberResponse, error) { // 반환 타입 수정
	return a.GetMemberCustomHTTPContext(context.Background(), apikey, httpClient, id)
}

func (a *Account) GetMemberCustomHTTPContext(ctx context.Context, apikey string, httpClient *http.Client, id string) (*model.GetMemberResponse, error) { // 반환 타입 유지

	url := a.endPoint + "/common/v1/members/" + id
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed new request: %w", err)
	}

	query := req.URL.Query()
	req.URL.RawQuery = query.Encode()

	req.Header.Set("Authorization", fmt.Sprintf("dooray-api %s", apikey))

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil { // nil 체크 추가
			resp.Body.Close()
		}
	}()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var memberResponse model.GetMemberResponse
	if err := json.Unmarshal(resBody, &memberResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	memberResponse.RawJSON = string(resBody)

	return &memberResponse, nil
}
