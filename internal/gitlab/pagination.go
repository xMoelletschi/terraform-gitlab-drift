package gitlab

import gl "gitlab.com/gitlab-org/api/client-go/v2"

// paginate walks a GitLab list endpoint to completion, concatenating every page.
// fetch performs one request with the current options; paginate advances the
// page cursor via listOpts.Page between calls.
//
// A 403/404 is treated as "inaccessible": paginate stops and returns whatever it
// has collected so far with no error (see skipInaccessible), so one restricted or
// just-deleted parent never aborts the whole scan. Any other error is returned
// for the caller to wrap with resource-specific context.
func paginate[T any](listOpts *gl.ListOptions, resource, path string, fetch func() ([]T, *gl.Response, error)) ([]T, error) {
	var all []T
	for {
		page, resp, err := fetch()
		if err != nil {
			if skipInaccessible(err, resource, path) {
				return all, nil
			}
			return nil, err
		}
		all = append(all, page...)
		if resp.NextPage == 0 {
			return all, nil
		}
		listOpts.Page = resp.NextPage
	}
}
