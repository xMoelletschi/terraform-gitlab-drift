package terraform

import gl "gitlab.com/gitlab-org/api/client-go"

func accessLevelToString(level gl.AccessLevelValue) string {
	switch level {
	case gl.NoPermissions:
		return "no one"
	case gl.MinimalAccessPermissions:
		return "minimal"
	case gl.GuestPermissions:
		return "guest"
	case gl.PlannerPermissions:
		return "planner"
	case gl.ReporterPermissions:
		return "reporter"
	case gl.DeveloperPermissions:
		return "developer"
	case gl.MaintainerPermissions:
		return "maintainer"
	case gl.OwnerPermissions:
		return "owner"
	default:
		return "guest"
	}
}

func accessLevelIntToString(level int64) string {
	return accessLevelToString(gl.AccessLevelValue(level))
}
