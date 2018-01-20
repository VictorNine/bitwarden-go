package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/VictorNine/bitwarden-go/internal/database/mock"
)

func TestHandleLogin(t *testing.T) {
	keyHash, _ := reHashPassword("sjlcxv1TSe1wTHoYF50WJL3X07oCFxqhXYFeGfrbtII=", "nobody@example.com")
	db := &mock.MockDB{Username: "nobody@example.com", Password: keyHash, RefreshToken: "abcdef"}

	cases := []struct {
		data     url.Values
		db       database
		expected int
	}{{url.Values{"client_id": {"android"}, "grant_type": {"password"}, "username": {"nobody@example.com"}, "password": {"sjlcxv1TSe1wTHoYF50WJL3X07oCFxqhXYFeGfrbtII="}}, db, 200},
		{url.Values{"client_id": {"android"}, "grant_type": {"refresh_token"}, "refresh_token": {"abcdef"}}, db, 200},
		{url.Values{"client_id": {"android"}, "grant_type": {"password"}, "username": {"nobody@example.com"}, "password": {""}}, db, 401},
		{url.Values{"client_id": {"android"}, "grant_type": {"refresh_token"}, "refresh_token": {"nasdfasdf"}}, db, 401},
		{url.Values{"client_id": {"web"}, "grant_type": {"password"}, "username": {"nobody@example.com"}, "password": {"sjlcxv1TSe1wTHoYF50WJL3X07oCFxqhXYFeGfrbtII="}}, &mock.MockDB{Username: "nobody@example.com", Password: keyHash, RefreshToken: "abcdef", TwoFactorSecret: "ABC"}, 400}, // Test two factor login
	}

	for _, c := range cases {
		authHandler := New(c.db, "", 3600)

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
