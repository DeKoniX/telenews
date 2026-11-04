package parse

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAPIErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		vk     bool
	}{
		{"X unauthorized", 401, `{}`, false},
		{"X rate limited", 429, `{}`, false},
		{"X malformed JSON", 200, `{`, false},
		{"X API error", 200, `{"errors":[{"title":"Invalid Request"}]}`, false},
		{"VK unavailable", 503, `{}`, true},
		{"VK malformed JSON", 200, `{`, true},
		{"VK rate limited", 200, `{"error":{"error_code":6,"error_msg":"Too many requests"}}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			client := &http.Client{Transport: testTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})}
			parser := ParseNewsStruct{}
			parseRequest := parser.parseTwitter
			if tc.vk {
				parseRequest = parser.parseVKWall
			}

			// When
			news, err := parseRequest(client, "golang")

			// Then
			if err == nil || len(news) != 0 {
				t.Fatalf("expected error without news, got %v, %v", news, err)
			}
		})
	}
}
