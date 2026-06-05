package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestPostWeComUsesGatewayAuthorization(t *testing.T) {
	var seen *http.Request
	client := &wecomClient{
		cfg: config{
			GatewayBaseURL: "https://gateway.example.test/wecom",
			GatewayToken:   "gateway-token",
		},
		http: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			seen = req
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"errcode":0}`)),
				Header:     make(http.Header),
			}, nil
		})},
	}

	if _, err := client.postWeComRaw("/cgi-bin/oa/calendar/get", map[string]string{"cal_id": "cal-1"}); err != nil {
		t.Fatalf("postWeComRaw returned error: %v", err)
	}
	if seen == nil {
		t.Fatal("expected HTTP request")
	}
	if got, want := seen.URL.String(), "https://gateway.example.test/wecom/cgi-bin/oa/calendar/get"; got != want {
		t.Fatalf("request URL = %q, want %q", got, want)
	}
	if got, want := seen.Header.Get("Authorization"), "Bearer gateway-token"; got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
	if strings.Contains(seen.URL.RawQuery, "access_token") {
		t.Fatalf("request query unexpectedly includes access_token: %q", seen.URL.RawQuery)
	}
}

func TestRequireCredentialsRequiresGatewayConfig(t *testing.T) {
	client := &wecomClient{cfg: config{}}
	err := client.requireCredentials()
	if err == nil {
		t.Fatal("expected missing gateway configuration error")
	}
	msg := err.Error()
	for _, want := range []string{"--gateway-base-url or WECOM_GATEWAY_BASE_URL", "--identity-file or CLI_IDENTITY_FILE"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q does not mention %q", msg, want)
		}
	}
}
