package project

import (
	"github.com/dooray-go/dooray-sdk/utils"
	"net/http"
)

type Project struct {
	endPoint   string
	httpClient *http.Client
}

func NewDefaultProject() *Project {
	return &Project{
		endPoint:   "https://api.dooray.com",
		httpClient: utils.NewDefaultHTTPClient(),
	}
}

func NewProject(endPoint string) *Project {
	return &Project{
		endPoint:   endPoint,
		httpClient: utils.NewDefaultHTTPClient(),
	}
}

func NewProjectWithClient(endPoint string, httpClient *http.Client) *Project {
	return &Project{
		endPoint:   endPoint,
		httpClient: httpClient,
	}
}
