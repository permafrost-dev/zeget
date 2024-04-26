package app_test

import (
	"net/http"
	"testing"

	"github.com/permafrost-dev/eget/app"
)

func TestGithubError_Error(t *testing.T) {
	tests := []struct {
		name        string
		githubError app.GithubError
		want        string
	}{
		{
			name: "Forbidden error with message",
			githubError: app.GithubError{
				Code:   http.StatusForbidden,
				Status: "403 Forbidden",
				Body:   []byte(`{"message":"rate limit exceeded","documentation_url":"https://developer.github.com/v3/#rate-limiting"}`),
				Url:    "https://api.github.com/users/octocat",
			},
			want: "403 Forbidden: rate limit exceeded: https://developer.github.com/v3/#rate-limiting",
		},
		{
			name: "Other error without specific message",
			githubError: app.GithubError{
				Code:   http.StatusNotFound,
				Status: "404 Not Found",
				Body:   []byte(`{"message":"Not Found","documentation_url":"https://developer.github.com/v3"}`),
				Url:    "https://api.github.com/users/octocat",
			},
			want: "404 Not Found (URL: https://api.github.com/users/octocat)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.githubError.Error(); got != tt.want {
				t.Errorf("GithubError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
