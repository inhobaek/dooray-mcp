package account

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestAccount_GetMember_OK(t *testing.T) {
	id := "1234567890"
	expectResponse := `{"id":"1234567890"}`

	var actual string
	mux := http.NewServeMux()
	mux.HandleFunc("/common/v1/members/{id}", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		actual = r.PathValue("id")

		response := []byte(expectResponse)
		rw.Write(response)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	response, err := NewAccount(server.URL).GetMember("dooray-api-key", id)
	if err != nil {
		t.Errorf("Expected not to receive error: %s", err)
	}

	if !reflect.DeepEqual(id, actual) {
		t.Errorf("Response did not match\nwant: %#v\n got: %#v", id, actual)
	}

	if !reflect.DeepEqual(expectResponse, response.RawJSON) {
		t.Errorf("Response did not match\nwant: %#v\n got: %#v", expectResponse, response.RawJSON)
	}
}
