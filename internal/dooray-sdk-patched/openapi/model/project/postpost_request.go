package project

import "github.com/dooray-go/dooray-sdk/utils"

// PostMember represents a member in the post.
type PostMember struct {
	OrganizationMemberID string `json:"organizationMemberId,omitempty"`
}

// PostEmailUser represents an email user in the post.
type PostEmailUser struct {
	EmailAddress string `json:"emailAddress,omitempty"`
	Name         string `json:"name,omitempty"`
}

// PostRecipient represents a recipient (to or cc) in the post.
type PostRecipient struct {
	Type      string         `json:"type"`
	Member    *PostMember    `json:"member,omitempty"`
	EmailUser *PostEmailUser `json:"emailUser,omitempty"`
}

// PostUsers represents the users involved in the post.
type PostUsers struct {
	To []PostRecipient `json:"to,omitempty"`
	Cc []PostRecipient `json:"cc,omitempty"`
}

// PostBody represents the body of the post.
type PostBody struct {
	MimeType string `json:"mimeType"`
	Content  string `json:"content"`
}

// PostRequest represents the payload for creating a project post.
type PostRequest struct {
	Subject      string          `json:"subject"`
	Body         PostBody        `json:"body"`
	Users        *PostUsers      `json:"users,omitempty"`
	DueDate      *utils.JsonTime `json:"dueDate,omitempty"`
	Priority     string          `json:"priority,omitempty"`     // urgent | high | normal | low
	MilestoneID  string          `json:"milestoneId,omitempty"`
	TagIDs       []string        `json:"tagIds,omitempty"`
	ParentPostID string          `json:"parentPostId,omitempty"`
	WorkflowID   string          `json:"workflowId,omitempty"`
}
