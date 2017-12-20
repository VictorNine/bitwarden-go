package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	bw "github.com/VictorNine/bitwarden-go/internal/common"
)

func TestHandleLogin(t *testing.T) {
	cases := []struct {
		data     url.Values
		expected int
	}{{url.Values{"client_id": {"android"}, "grant_type": {"password"}, "username": {"nobody@example.com"}, "password": {"sjlcxv1TSe1wTHoYF50WJL3X07oCFxqhXYFeGfrbtII="}}, 200},
		{url.Values{"client_id": {"android"}, "grant_type": {"refresh_token"}, "refresh_token": {"abcdef"}}, 200},
		{url.Values{"client_id": {"android"}, "grant_type": {"password"}, "username": {"nobody@example.com"}, "password": {""}}, 401},
		{url.Values{"client_id": {"android"}, "grant_type": {"refresh_token"}, "refresh_token": {"nasdfasdf"}}, 401},
	}

	keyHash, _ := reHashPassword("sjlcxv1TSe1wTHoYF50WJL3X07oCFxqhXYFeGfrbtII=", "nobody@example.com")
	db := &bw.MockDB{Username: "nobody@example.com", Password: keyHash, RefreshToken: "abcdef"}
	authHandler := New(db, nil, 3600)

	for _, c := range cases {
		req, err := http.NewRequest("POST", "/identity/connect/token", strings.NewReader(c.data.Encode()))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(c.data.Encode())))

		res := httptest.NewRecorder()

		authHandler.HandleLogin(res, req)
		if res.Code != c.expected {
			t.Errorf("Expected %v got %v", c.expected, res.Code)
		}
	}

}
