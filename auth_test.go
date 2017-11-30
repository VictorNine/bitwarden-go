package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestHandleLogin(t *testing.T) {
	cases := []struct {
		data     url.Values
		expected int
	}{{url.Values{"grant_type": {"password"}, "username": {"nobody@example.com"}, "password": {"sjlcxv1TSe1wTHoYF50WJL3X07oCFxqhXYFeGfrbtII="}}, 200},
		{url.Values{"grant_type": {"refresh_token"}, "refresh_token": {"abcdef"}}, 200},
		{url.Values{"grant_type": {"password"}, "username": {"nobody@example.com"}, "password": {""}}, 401},
		{url.Values{"grant_type": {"refresh_token"}, "refresh_token": {"nasdfasdf"}}, 401}}

	keyHash, _ := reHashPassword("sjlcxv1TSe1wTHoYF50WJL3X07oCFxqhXYFeGfrbtII=", "nobody@example.com")
	db = &mockDB{username: "nobody@example.com", password: keyHash, refreshToken: "abcdef"}

	for _, c := range cases {
		req, err := http.NewRequest("POST", "/identity/connect/token", strings.NewReader(c.data.Encode()))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(c.data.Encode())))

		res := httptest.NewRecorder()

		handleLogin(res, req)
		if res.Code != c.expected {
			t.Errorf("Expected %v got %v", c.expected, res.Code)
		}
	}

}
