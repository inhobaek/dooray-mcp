package calendar

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCalendar_GetCalendars_OK(t *testing.T) {
	expectResponse := `{"id":"1234567890"}`

	mux := http.NewServeMux()
	mux.HandleFunc("/calendar/v1/calendars", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		response := []byte(expectResponse)
		rw.Write(response)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	response, err := NewCalendar(server.URL).GetCalendars("dooray-api-key")
	if err != nil {
		t.Errorf("Expected not to receive error: %s", err)
	}

	if !reflect.DeepEqual(expectResponse, response.RawJSON) {
		t.Errorf("Response did not match\nwant: %#v\n got: %#v", expectResponse, response.RawJSON)
	}
}
