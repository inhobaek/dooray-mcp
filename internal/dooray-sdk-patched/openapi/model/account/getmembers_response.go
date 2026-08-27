package account

import "github.com/dooray-go/dooray-sdk/openapi/model"

// Member represents a single member in the API response.
type Member struct {
	ID                   string `json:"id"`                   // Dooray Member ID
	Name                 string `json:"name"`                 // 사용자 이름
	UserCode             string `json:"userCode"`             // 사용자 ID
	ExternalEmailAddress string `json:"externalEmailAddress"` // 외부 이메일 주소
}

// GetMembersResponse represents the full API response for member retrieval.
type GetMembersResponse struct {
	Header     model.ResponseHeader `json:"header"`
	Result     []Member             `json:"result"`
	TotalCount int                  `json:"totalCount"` // 필터 조건에 맞는 전체 아이템 수
	RawJSON    string               `json:"-"`          // 원본 JSON 응답 (디버깅 또는 로깅용)
}
