package calendar

import (
	"github.com/dooray-go/dooray-sdk/utils"
	"net/http"
)

type Calendar struct {
	endPoint   string
	httpClient *http.Client
}

func NewDefaultCalendar() *Calendar {
	return &Calendar{
		endPoint:   "https://api.dooray.com",
		httpClient: utils.NewDefaultHTTPClient(),
	}
}

func NewCalendar(endPoint string) *Calendar {
	return &Calendar{
		endPoint:   endPoint,
		httpClient: utils.NewDefaultHTTPClient(),
	}
}

func NewCalendarWithClient(endPoint string, httpClient *http.Client) *Calendar {
	return &Calendar{
		endPoint:   endPoint,
		httpClient: httpClient,
	}
}
