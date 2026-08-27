package account

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestAccount_GetMembers_OK(t *testing.T) {
	name := "Manty"
	userCode := "test"
	externalEmailAddress := "manty@manty.dooray.com" 

	expectResponse := `{
    "header": {
        "isSuccessful": true,
        "resultCode": 0,
        "resultMessage": ""
    },
    "result": [{
        "id": "1",                                           
        "name": "Manty",                                       
        "userCode": "test",                                         
        "externalEmailAddress": "`+externalEmailAddress+`"        

    }],
    "totalCount": 1                                             
}`

	var actualName string
	var actualUserCode string
	mux := http.NewServeMux()
	mux.HandleFunc("/common/v1/members", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()
		actualName = query.Get("name")
		actualUserCode = query.Get("userCode")

		response := []byte(expectResponse)
		rw.Write(response)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	actualResponse, err := NewAccount(server.URL).GetMembers("dooray-api-key", name, userCode)
	if err != nil {
		t.Errorf("Expected not to receive error: %s", err)
	}

	if !reflect.DeepEqual(name, actualName) {
		t.Errorf("Response did not match\nwant: %#v\n got: %#v", name, actualName)
	}

	if !reflect.DeepEqual(userCode, actualUserCode) {
		t.Errorf("Response did not match\nwant: %#v\n got: %#v", userCode, actualUserCode)
	}

	if !reflect.DeepEqual(expectResponse, actualResponse.RawJSON) {
		t.Errorf("Response did not match\nwant: %#v\n got: %#v", expectResponse, actualResponse.RawJSON)
	}

	if !reflect.DeepEqual(externalEmailAddress, actualResponse.Result[0].ExternalEmailAddress) {
		t.Errorf("Response did not match\nwant: %#v\n got: %#v", externalEmailAddress, actualResponse.Result[0].ExternalEmailAddress)
	}
}
