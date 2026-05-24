package terraform

import (
	"bytes"
	"testing"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func TestWriteJobTokenScopes_ManagedRefs(t *testing.T) {
	project := &gl.Project{
		ID:                1,
		Path:              "my-project",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/my-project",
	}

	allowedProjects := []*gl.Project{
		{ID: 10, Path: "foo", Namespace: &gl.ProjectNamespace{FullPath: "my-group"}, PathWithNamespace: "my-group/foo"},
		{ID: 11, Path: "bar", Namespace: &gl.ProjectNamespace{FullPath: "my-group"}, PathWithNamespace: "my-group/bar"},
	}
	allowedGroups := []*gl.Group{
		{ID: 20, Path: "other-group", FullPath: "other-group"},
	}
	scope := &gitlab.JobTokenScope{
		InboundEnabled:  true,
		AllowedProjects: allowedProjects,
		AllowedGroups:   allowedGroups,
	}

	projectRefs := buildProjectRefMap(append([]*gl.Project{project}, allowedProjects...))
	groupRefs := buildGroupRefMap(allowedGroups)

	var buf bytes.Buffer
	if err := WriteJobTokenScopes(project, scope, projectRefs, groupRefs, &buf); err != nil {
		t.Fatalf("WriteJobTokenScopes error: %v", err)
	}

	compareGolden(t, "job_token_scopes_managed_refs.tf", buf.String())
}

func TestWriteJobTokenScopes_Mixed(t *testing.T) {
	project := &gl.Project{
		ID:                1,
		Path:              "my-project",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/my-project",
	}

	inScopeProject := &gl.Project{
		ID:                10,
		Path:              "foo",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/foo",
	}
	externalProject := &gl.Project{
		ID:                999,
		Path:              "external",
		PathWithNamespace: "other-org/external",
	}
	externalGroup := &gl.Group{
		ID:       888,
		Path:     "external-group",
		FullPath: "other-org/external-group",
	}
	scope := &gitlab.JobTokenScope{
		InboundEnabled:  true,
		AllowedProjects: []*gl.Project{inScopeProject, externalProject},
		AllowedGroups:   []*gl.Group{externalGroup},
	}

	projectRefs := buildProjectRefMap([]*gl.Project{project, inScopeProject})
	groupRefs := buildGroupRefMap(nil)

	var buf bytes.Buffer
	if err := WriteJobTokenScopes(project, scope, projectRefs, groupRefs, &buf); err != nil {
		t.Fatalf("WriteJobTokenScopes error: %v", err)
	}

	compareGolden(t, "job_token_scopes_mixed.tf", buf.String())
}

func TestWriteJobTokenScopes_EnabledOnly(t *testing.T) {
	project := &gl.Project{
		ID:                1,
		Path:              "my-project",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/my-project",
	}

	scope := &gitlab.JobTokenScope{
		InboundEnabled: true,
	}

	projectRefs := buildProjectRefMap([]*gl.Project{project})
	groupRefs := buildGroupRefMap(nil)

	var buf bytes.Buffer
	if err := WriteJobTokenScopes(project, scope, projectRefs, groupRefs, &buf); err != nil {
		t.Fatalf("WriteJobTokenScopes error: %v", err)
	}

	compareGolden(t, "job_token_scopes_enabled_only.tf", buf.String())
}

func TestWriteJobTokenScopes_WithSelf(t *testing.T) {
	project := &gl.Project{
		ID:                1,
		Path:              "my-project",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/my-project",
	}

	other := &gl.Project{
		ID:                10,
		Path:              "foo",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/foo",
	}
	scope := &gitlab.JobTokenScope{
		InboundEnabled:  true,
		AllowedProjects: []*gl.Project{project, other},
	}

	projectRefs := buildProjectRefMap([]*gl.Project{project, other})
	groupRefs := buildGroupRefMap(nil)

	var buf bytes.Buffer
	if err := WriteJobTokenScopes(project, scope, projectRefs, groupRefs, &buf); err != nil {
		t.Fatalf("WriteJobTokenScopes error: %v", err)
	}

	compareGolden(t, "job_token_scopes_with_self.tf", buf.String())
}

func TestWriteJobTokenScopes_Disabled(t *testing.T) {
	project := &gl.Project{
		ID:                1,
		Path:              "my-project",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/my-project",
	}

	scope := &gitlab.JobTokenScope{
		InboundEnabled: false,
	}

	projectRefs := buildProjectRefMap([]*gl.Project{project})
	groupRefs := buildGroupRefMap(nil)

	var buf bytes.Buffer
	if err := WriteJobTokenScopes(project, scope, projectRefs, groupRefs, &buf); err != nil {
		t.Fatalf("WriteJobTokenScopes error: %v", err)
	}

	compareGolden(t, "job_token_scopes_disabled.tf", buf.String())
}
