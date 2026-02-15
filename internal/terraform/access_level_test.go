package terraform

import (
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func TestAccessLevelToString(t *testing.T) {
	tests := []struct {
		name  string
		level gl.AccessLevelValue
		want  string
	}{
		{"no permissions", gl.NoPermissions, "no one"},
		{"minimal", gl.MinimalAccessPermissions, "minimal"},
		{"guest", gl.GuestPermissions, "guest"},
		{"planner", gl.PlannerPermissions, "planner"},
		{"reporter", gl.ReporterPermissions, "reporter"},
		{"developer", gl.DeveloperPermissions, "developer"},
		{"maintainer", gl.MaintainerPermissions, "maintainer"},
		{"owner", gl.OwnerPermissions, "owner"},
		{"unknown defaults to guest", gl.AccessLevelValue(99), "guest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := accessLevelToString(tt.level)
			if got != tt.want {
				t.Errorf("accessLevelToString(%d) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

func TestAccessLevelIntToString(t *testing.T) {
	tests := []struct {
		name  string
		level int64
		want  string
	}{
		{"developer", 30, "developer"},
		{"maintainer", 40, "maintainer"},
		{"owner", 50, "owner"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := accessLevelIntToString(tt.level)
			if got != tt.want {
				t.Errorf("accessLevelIntToString(%d) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}
