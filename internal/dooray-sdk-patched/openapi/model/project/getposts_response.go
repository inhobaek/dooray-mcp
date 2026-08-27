package project

import (
	"github.com/dooray-go/dooray-sdk/openapi/model"
)

type ParentInfo struct {
	ID      string `json:"id"`      // 상위 업무 ID
	Number  int    `json:"number"`  // 상위 업무 번호
	Subject string `json:"subject"` // 상위 업무 제목
}

type MilestoneInfo struct {
	ID   string `json:"id"`   // 마일스톤 ID
	Name string `json:"name"` // 마일스톤 이름
}

type TagInfo struct {
	ID string `json:"id"` // 태그 ID
}

type UserFrom struct {
	Type   string `json:"type"` // 생성자 타입 (member | emailuser)
	Member struct {
		OrganizationMemberID string `json:"organizationmemberid"` // 생성자 멤버 ID
	} `json:"member"`
}

type UserTo struct {
	Type   string `json:"type"` // 담당자 타입 (member | emailUser | group)
	Member struct {
		OrganizationMemberID string `json:"organizationMemberId"` // 담당자 멤버 ID
	} `json:"member,omitempty"`
	EmailUser struct {
		EmailAddress string `json:"emailAddress"` // 담당자 이메일 주소
		Name         string `json:"name"`         // 담당자 이름
	} `json:"emailUser,omitempty"`
	Workflow struct {
		ID   string `json:"id"`   // 워크플로우 ID
		Name string `json:"name"` // 워크플로우 이름
	} `json:"workflow"`
}

type UserCC struct {
	Type  string `json:"type"` // 참조자 타입 (group)
	Group struct {
		ProjectMemberGroupID string `json:"projectMemberGroupId"` // 그룹 ID
		Members              []struct {
			OrganizationMemberID string `json:"organizationMemberId"` // 그룹 멤버 ID
		} `json:"members"` // 그룹 멤버 목록
	} `json:"group"`
}

type UsersInfo struct {
	From UserFrom `json:"from"` // 생성자
	To   []UserTo `json:"to"`   // 담당자 목록
	CC   []UserCC `json:"cc"`   // 참조자 목록
}

type WorkflowInfo struct {
	ID   string `json:"id"`   // 워크플로우 ID
	Name string `json:"name"` // 워크플로우 이름
}

type PostInfo struct {
	ID            string        `json:"id"`            // 업무 ID
	Subject       string        `json:"subject"`       // 업무 제목
	Project       ProjectInfo   `json:"project"`       // 업무가 속한 프로젝트
	TaskNumber    string        `json:"taskNumber"`    // projectCode/number
	Closed        bool          `json:"closed"`        // 업무 완료 상태
	CreatedAt     string        `json:"createdAt"`     // 업무 생성 날짜시간 (ISO8601)
	DueDate       string        `json:"dueDate"`       // 업무 만기 날짜시간 (ISO8601)
	DueDateFlag   bool          `json:"dueDateFlag"`   // 만기 플래그
	UpdatedAt     string        `json:"updatedAt"`     // 업무 업데이트 날짜시간
	Number        int           `json:"number"`        // 업무 번호
	Priority      string        `json:"priority"`      // 우선순위
	Parent        ParentInfo    `json:"parent"`        // 상위 업무
	WorkflowClass string        `json:"workflowClass"` // 업무 상태 (registered | working | closed)
	Milestone     MilestoneInfo `json:"milestone"`     // 마일스톤
	Tags          []TagInfo     `json:"tags"`          // 태그 목록
	Users         UsersInfo     `json:"users"`         // 사용자 정보
	Workflow      WorkflowInfo  `json:"workflow"`      // 워크플로우 정보
}

type GetPostsResponse struct {
	Header     model.ResponseHeader `json:"header"`
	Result     []PostInfo           `json:"result"`
	TotalCount int                  `json:"totalCount"`
	RawJSON    string               `json:"-"` // 원본 JSON 응답 (디버깅 또는 로깅용)
}
