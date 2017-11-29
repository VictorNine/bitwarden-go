package main

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestUnmarshalCipher(t *testing.T) {
	testData := "{\"type\": 1,\"folderId\": null,\"organizationId\": null,\"name\": \"2.d7MttWzJTSSKx1qXjHUxlQ==|01Ath5UqFZHk7csk5DVtkQ==|EMLoLREgCUP5Cu4HqIhcLqhiZHn+NsUDp8dAg1Xu0Io=\",\"notes\": null,\"favorite\": false,\"login\": {\"uri\": \"2.T57BwAuV8ubIn/sZPbQC+A==|EhUSSpJWSzSYOdJ/AQzfXuUXxwzcs/6C4tOXqhWAqcM=|OWV2VIqLfoWPs9DiouXGUOtTEkVeklbtJQHkQFIXkC8=\",\"username\": \"2.JbFkAEZPnuMm70cdP44wtA==|fsN6nbT+udGmOWv8K4otgw==|JbtwmNQa7/48KszT2hAdxpmJ6DRPZst0EDEZx5GzesI=\",\"password\": \"2.e83hIsk6IRevSr/H1lvZhg==|48KNkSCoTacopXRmIZsbWg==|CIcWgNbaIN2ix2Fx1Gar6rWQeVeboehp4bioAwngr0o=\",\"totp\": null}}"

	r := ioutil.NopCloser(bytes.NewBuffer([]byte(testData)))
	Ci, err := unmarshalCipher(r)

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	if Ci.Data.Notes != nil {
		t.Fatal("Should be nil og android app will crash")
	}

	if Ci.Type != 1 {
		t.Fatal("Wrong type")
	}
}
