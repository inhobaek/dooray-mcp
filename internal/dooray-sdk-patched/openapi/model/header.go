package model

// EventResponseHeader represents the header part of the API response.
type ResponseHeader struct {
	IsSuccessful  bool   `json:"isSuccessful"`
	ResultCode    int    `json:"resultCode"`
	ResultMessage string `json:"resultMessage"`
}
