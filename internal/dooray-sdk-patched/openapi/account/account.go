package account

import (
	"github.com/dooray-go/dooray-sdk/utils"
	"net/http"
)

type Account struct {
	endPoint   string
	httpClient *http.Client
}

func NewDefaultAccount() *Account {
	return &Account{
		endPoint:   "https://api.dooray.com",
		httpClient: utils.NewDefaultHTTPClient(),
	}
}

func NewAccount(endPoint string) *Account {
	return &Account{
		endPoint:   endPoint,
		httpClient: utils.NewDefaultHTTPClient(),
	}
}

func NewAccountWithClient(endPoint string, httpClient *http.Client) *Account {
	return &Account{
		endPoint:   endPoint,
		httpClient: httpClient,
	}
}
