package account

import "github.com/dooray-go/dooray-sdk/openapi/model"

// DefaultOrganization represents the default organization information in the response.
type DefaultOrganization struct {
	ID string `json:"id"` // Default Organization ID
}

// MemberResult represents the detailed member information in the response.
type MemberResult struct {
	ID                   string              `json:"id"`                   // Dooray Member ID
	IDProviderType       string              `json:"idProviderType"`       // sso, service
	IDProviderUserID     string              `json:"idProviderUserId"`     // ID Provider User ID
	Name                 string              `json:"name"`                 // Member Name
	UserCode             string              `json:"userCode"`             // User Code
	ExternalEmailAddress string              `json:"externalEmailAddress"` // External Email Address
	DefaultOrganization  DefaultOrganization `json:"defaultOrganization"`  // Default Organization
	Locale               string              `json:"locale"`               // Locale
	TimezoneName         string              `json:"timezoneName"`         // Timezone Name
	EnglishName          string              `json:"englishName"`          // English Name
	NativeName           string              `json:"nativeName"`           // Native Name
	Nickname             string              `json:"nickname"`             // Nickname
	DisplayMemberID      string              `json:"displayMemberId"`      // Display Member ID
}

// GetMemberResponse represents the full API response for retrieving a single member.
type GetMemberResponse struct {
	Header  model.ResponseHeader `json:"header"` // Response Header
	Result  MemberResult         `json:"result"` // Member Details
	RawJSON string               `json:"-"`      // Raw JSON Response (for debugging or logging)
}
