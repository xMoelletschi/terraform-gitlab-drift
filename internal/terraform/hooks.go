package terraform

import (
	"io"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	gl "gitlab.com/gitlab-org/api/client-go"
)

func normalizeHookURL(rawURL string) string {
	u := strings.TrimPrefix(rawURL, "https://")
	u = strings.TrimPrefix(u, "http://")
	u = strings.ReplaceAll(u, ":", "_")
	return normalizeName(u)
}

func projectHookResourceName(p *gl.Project, h *gl.ProjectHook) string {
	return projectResourceName(p) + "_" + normalizeHookURL(h.URL)
}

func groupHookResourceName(g *gl.Group, h *gl.GroupHook) string {
	return normalizeToTerraformName(g.Path) + "_" + normalizeHookURL(h.URL)
}

func WriteProjectHooks(p *gl.Project, hooks []*gl.ProjectHook, w io.Writer) error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	projName := projectResourceName(p)

	for i, h := range hooks {
		if i > 0 {
			rootBody.AppendNewline()
		}
		name := projectHookResourceName(p, h)
		block := rootBody.AppendNewBlock("resource", []string{"gitlab_project_hook", name})
		body := block.Body()

		body.SetAttributeTraversal("project", hcl.Traversal{
			hcl.TraverseRoot{Name: "gitlab_project"},
			hcl.TraverseAttr{Name: projName},
			hcl.TraverseAttr{Name: "id"},
		})
		body.SetAttributeValue("url", cty.StringVal(h.URL))
		if h.Name != "" {
			body.SetAttributeValue("name", cty.StringVal(h.Name))
		}
		if h.Description != "" {
			body.SetAttributeValue("description", cty.StringVal(h.Description))
		}
		body.SetAttributeValue("enable_ssl_verification", cty.BoolVal(h.EnableSSLVerification))

		// push_events: always write (provider default is true, so we need explicit false)
		body.SetAttributeValue("push_events", cty.BoolVal(h.PushEvents))
		if h.PushEventsBranchFilter != "" {
			body.SetAttributeValue("push_events_branch_filter", cty.StringVal(h.PushEventsBranchFilter))
		}

		writeEvents(body, []hookEvent{
			{"tag_push_events", h.TagPushEvents},
			{"issues_events", h.IssuesEvents},
			{"confidential_issues_events", h.ConfidentialIssuesEvents},
			{"merge_requests_events", h.MergeRequestsEvents},
			{"note_events", h.NoteEvents},
			{"confidential_note_events", h.ConfidentialNoteEvents},
			{"job_events", h.JobEvents},
			{"pipeline_events", h.PipelineEvents},
			{"wiki_page_events", h.WikiPageEvents},
			{"deployment_events", h.DeploymentEvents},
			{"releases_events", h.ReleasesEvents},
		})
	}

	_, err := w.Write(f.Bytes())
	return err
}

func WriteGroupHooks(g *gl.Group, hooks []*gl.GroupHook, w io.Writer) error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	groupName := normalizeToTerraformName(g.Path)

	for i, h := range hooks {
		if i > 0 {
			rootBody.AppendNewline()
		}
		name := groupHookResourceName(g, h)
		block := rootBody.AppendNewBlock("resource", []string{"gitlab_group_hook", name})
		body := block.Body()

		body.SetAttributeTraversal("group", hcl.Traversal{
			hcl.TraverseRoot{Name: "gitlab_group"},
			hcl.TraverseAttr{Name: groupName},
			hcl.TraverseAttr{Name: "id"},
		})
		body.SetAttributeValue("url", cty.StringVal(h.URL))
		if h.Name != "" {
			body.SetAttributeValue("name", cty.StringVal(h.Name))
		}
		if h.Description != "" {
			body.SetAttributeValue("description", cty.StringVal(h.Description))
		}
		body.SetAttributeValue("enable_ssl_verification", cty.BoolVal(h.EnableSSLVerification))

		// push_events: always write
		body.SetAttributeValue("push_events", cty.BoolVal(h.PushEvents))
		if h.PushEventsBranchFilter != "" {
			body.SetAttributeValue("push_events_branch_filter", cty.StringVal(h.PushEventsBranchFilter))
		}
		if h.BranchFilterStrategy != "" {
			body.SetAttributeValue("branch_filter_strategy", cty.StringVal(h.BranchFilterStrategy))
		}

		writeEvents(body, []hookEvent{
			{"tag_push_events", h.TagPushEvents},
			{"issues_events", h.IssuesEvents},
			{"confidential_issues_events", h.ConfidentialIssuesEvents},
			{"merge_requests_events", h.MergeRequestsEvents},
			{"note_events", h.NoteEvents},
			{"confidential_note_events", h.ConfidentialNoteEvents},
			{"job_events", h.JobEvents},
			{"pipeline_events", h.PipelineEvents},
			{"wiki_page_events", h.WikiPageEvents},
			{"deployment_events", h.DeploymentEvents},
			{"releases_events", h.ReleasesEvents},
			{"subgroup_events", h.SubGroupEvents},
			{"emoji_events", h.EmojiEvents},
			{"feature_flag_events", h.FeatureFlagEvents},
		})
	}

	_, err := w.Write(f.Bytes())
	return err
}

type hookEvent struct {
	attr string
	val  bool
}

func writeEvents(body *hclwrite.Body, events []hookEvent) {
	for _, e := range events {
		if e.val {
			body.SetAttributeValue(e.attr, cty.True)
		}
	}
}
