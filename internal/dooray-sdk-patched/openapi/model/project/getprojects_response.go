package project

import "github.com/dooray-go/dooray-sdk/openapi/model"

type GetProjectsResponse struct {
	Header     model.ResponseHeader `json:"header"`
	Result     []ProjectInfo        `json:"result"`
	TotalCount int                  `json:"totalCount"`
	RawJSON    string               `json:"-"`
}

type ProjectInfo struct {
	ID           string       `json:"id"`
	Code         string       `json:"code"`
	Description  string       `json:"description"`
	State        string       `json:"state"`
	Scope        string       `json:"scope"`
	Type         string       `json:"type"`
	Organization Organization `json:"organization"`
	Drive        DriveInfo    `json:"drive"`
	Wiki         WikiInfo     `json:"wiki"`
}

type Organization struct {
	ID string `json:"id"`
}

type DriveInfo struct {
	ID string `json:"id"`
}

type WikiInfo struct {
	ID string `json:"id"`
}
