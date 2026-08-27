package messenger

import (
	"github.com/dooray-go/dooray-sdk/utils"
	"net/http"
)

type Messenger struct {
	endPoint   string
	httpClient *http.Client
}

func NewDefaultMessenger() *Messenger {
	return &Messenger{
		endPoint:   "https://api.dooray.com",
		httpClient: utils.NewDefaultHTTPClient(),
	}
}

func NewMessenger(endPoint string) *Messenger {
	return &Messenger{
		endPoint:   endPoint,
		httpClient: utils.NewDefaultHTTPClient(),
	}
}

func NewMessengerWithClient(endPoint string, httpClient *http.Client) *Messenger {
	return &Messenger{
		endPoint:   endPoint,
		httpClient: httpClient,
	}
}
