package hook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestPostWebhook_OK(t *testing.T) {

	var receivedPayload WebhookMessage

	mux := http.NewServeMux()
	mux.HandleFunc("/services", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&receivedPayload)
		if err != nil {
			t.Errorf("Request contained invalid JSON, %s", err)
		}

		response := []byte(`{}`)
		rw.Write(response)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	url := server.URL + "/services"

	payload := &WebhookMessage{
		Text: "Test Text",
		Attachments: []Attachment{
			{
				Text: "Foo",
			},
		},
	}

	err := PostWebhook(url, payload)

	if err != nil {
		t.Errorf("Expected not to receive error: %s", err)
	}

	if !reflect.DeepEqual(payload, &receivedPayload) {
		t.Errorf("Payload did not match\nwant: %#v\n got: %#v", payload, receivedPayload)
	}
}
