package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCalendarCreateStoresAGWResource(t *testing.T) {
	var requests []struct {
		method string
		path   string
		auth   string
		body   map[string]any
	}
	client := &wecomClient{
		cfg: config{
			GatewayBaseURL: "https://gateway.example.test/wecom",
			AGWBaseURL:     "https://gateway.example.test",
			GatewayToken:   "gateway-token",
		},
		http: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			entry := struct {
				method string
				path   string
				auth   string
				body   map[string]any
			}{method: req.Method, path: req.URL.RequestURI(), auth: req.Header.Get("Authorization")}
			if req.Body != nil {
				_ = json.NewDecoder(req.Body).Decode(&entry.body)
			}
			requests = append(requests, entry)
			body := `{"errcode":0,"cal_id":"cal-1"}`
			if req.URL.Path == "/adm/api/user-agent-resource/add" {
				body = `{"code":0,"msg":"success","data":true}`
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		})},
	}

	req := calendarAddRequest{Calendar: calendarPayload{Summary: "Project Launch"}}
	if err := client.addCalendar(req); err != nil {
		t.Fatalf("addCalendar returned error: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("got %d requests, want 2", len(requests))
	}
	if got, want := requests[0].path, "/wecom/cgi-bin/oa/calendar/add"; got != want {
		t.Fatalf("WeCom request path = %q, want %q", got, want)
	}
	if got, want := requests[1].path, "/adm/api/user-agent-resource/add"; got != want {
		t.Fatalf("AGW request path = %q, want %q", got, want)
	}
	for i, request := range requests {
		if request.auth != "Bearer gateway-token" {
			t.Fatalf("request %d Authorization = %q, want bearer token", i, request.auth)
		}
	}
	want := map[string]string{
		"resourceType":  "calendar",
		"platformField": "cal_id",
		"externalId":    "cal-1",
		"name":          "Project Launch",
		"status":        "ACTIVE",
	}
	for key, value := range want {
		if got := requests[1].body[key]; got != value {
			t.Fatalf("stored resource %s = %#v, want %q", key, got, value)
		}
	}
	if metadata, ok := requests[1].body["metadataJson"].(string); !ok || !strings.Contains(metadata, "calendar create") {
		t.Fatalf("metadataJson = %#v, want command metadata", requests[1].body["metadataJson"])
	}
}

func TestUsersLookupUsesAGWMappingEndpoint(t *testing.T) {
	var seen *http.Request
	client := &wecomClient{
		cfg: config{
			GatewayBaseURL: "https://gateway.example.test/wecom",
			AGWBaseURL:     "https://gateway.example.test",
			GatewayToken:   "gateway-token",
		},
		http: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			seen = req
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"code":0,"data":{"qwUserid":"qw-1"}}`)),
				Header:     make(http.Header),
			}, nil
		})},
	}

	if err := runUsers(client, []string{"get-by-name", "--user-name", "Zhang San"}); err != nil {
		t.Fatalf("runUsers returned error: %v", err)
	}
	if seen == nil {
		t.Fatal("expected HTTP request")
	}
	if got, want := seen.URL.RequestURI(), "/adm/api/service/employee-wecom-mapping/by-user-name?userName=Zhang+San"; got != want {
		t.Fatalf("request URI = %q, want %q", got, want)
	}
	if got, want := seen.Header.Get("Authorization"), "Bearer gateway-token"; got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}
