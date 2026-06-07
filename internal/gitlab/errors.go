package gitlab

import (
	"errors"
	"log/slog"
	"net/http"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

// skipInaccessible reports whether err is a GitLab API error that means the
// resource is inaccessible to us (403 Forbidden) or no longer exists (404 Not
// Found). Such an error should skip the affected resource rather than abort the
// whole scan — a single restricted, archived, or just-deleted project or
// subgroup must not fail an entire instance scan. When it returns true it logs a
// path-named warning describing what was skipped.
func skipInaccessible(err error, resource, path string) bool {
	var errResp *gl.ErrorResponse
	if errors.As(err, &errResp) &&
		(errResp.HasStatusCode(http.StatusForbidden) || errResp.HasStatusCode(http.StatusNotFound)) {
		slog.Warn("skipping inaccessible resource",
			"resource", resource, "path", path, "status", errResp.StatusCode)
		return true
	}
	return false
}
