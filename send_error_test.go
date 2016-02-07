package httpctx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestShouldSendJSON(t *testing.T) {
	for i, tc := range []struct {
		Accept, Path string
		Want         bool
	}{
		{"", "", false},
		{"", "/api", false},
		{"", "/api/", true},
		{"application/json", "", true},
		{"application/json", "/api/", true},
	} {
		got := shouldSendJSON(&http.Request{
			URL:    &url.URL{Path: tc.Path},
			Header: http.Header(map[string][]string{"Accept": {tc.Accept}}),
		})
		if got != tc.Want {
			t.Errorf("%d. got %t, want %t (Path=%q Accept=%q).",
				i, got, tc.Want, tc.Path, tc.Accept)
		}
	}
}

func TestSendError(t *testing.T) {
	for i, tc := range []struct {
		Code int
		Err  error
		Body string
	}{
		{http.StatusInternalServerError, fmt.Errorf("ERR"),
			"ERR\n"},
		{3, HTTPError{Err: nil, Code: 3}, "\n"},
		{3, HTTPError{Err: fmt.Errorf("bad status code"), Code: 3}, "bad status code\n"},
		{500, HTTPErrorCode{Err: nil, StatusCode: "418 I'm not a Teapot"},
			`{"error":{"code":"418 I'm not a Teapot","message":"","status":500}}`},
		{500, HTTPErrorCode{Err: fmt.Errorf("something happened"), StatusCode: "418 I'm not a Teapot"},
			`{"error":{"code":"418 I'm not a Teapot","message":"something happened","status":500}}`},
	} {
		rr := httptest.NewRecorder()
		var hdr http.Header
		if _, ok := tc.Err.(HTTPErrorCode); ok {
			hdr = http.Header(map[string][]string{"Accept": {"application/json"}})
		}
		sendError(rr, &http.Request{URL: &url.URL{}, Header: hdr}, tc.Err)
		if rr.Code != tc.Code {
			t.Errorf("%d. got %d, want %d for %#v.", i, rr.Code, tc.Code, tc.Err)
		}
		got := rr.Body.String()
		if got != tc.Body {
			t.Errorf("%d. got %q, want %q for %v.", i, got, tc.Body, tc.Err)
		}
	}
}

type HTTPErrorCode struct {
	Err        error
	StatusCode string
}

// Code returns the HTTP status code accompanies the error.
func (he HTTPErrorCode) Code() string {
	return he.StatusCode
}
func (he HTTPErrorCode) Error() string {
	if he.Err == nil {
		return ""
	}
	return he.Err.Error()
}

// ErrorWithStatusCode allows errors to provide the HTTP status code to return.
type ErrorWithStatusCode interface {
	StatusCode() int
}

var _ = ErrorWithStatusCode(HTTPError{})
var _ = error(HTTPError{})

type HTTPError struct {
	Err  error
	Code int
}

// StatusCode returns the HTTP status code accompanies the error.
func (he HTTPError) StatusCode() int {
	return he.Code
}
func (he HTTPError) Error() string {
	if he.Err == nil {
		return ""
	}
	return he.Err.Error()
}
