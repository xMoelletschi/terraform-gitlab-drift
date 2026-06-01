package gitlab

import (
	"errors"
	"net/http"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func TestSkipInaccessible(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"forbidden", &gl.ErrorResponse{StatusCode: http.StatusForbidden}, true},
		{"not found", &gl.ErrorResponse{StatusCode: http.StatusNotFound}, true},
		{"not found sentinel", gl.ErrNotFound, true},
		{"server error", &gl.ErrorResponse{StatusCode: http.StatusInternalServerError}, false},
		{"generic error", errors.New("boom"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := skipInaccessible(tt.err, "test resource", "group/project")
			if got != tt.want {
				t.Errorf("skipInaccessible(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
