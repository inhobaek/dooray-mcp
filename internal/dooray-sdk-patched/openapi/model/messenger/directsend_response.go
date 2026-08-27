package messenger

import "github.com/dooray-go/dooray-sdk/openapi/model"

type DirectSendResult struct {
	ID int64 `json:"id"`
}
type DirectSendResponse struct {
	Header  model.ResponseHeader `json:"header"`
	Result  DirectSendResult     `json:"result"`
	RawJSON string               `json:"-"` // Raw JSON response for debugging or logging
}
