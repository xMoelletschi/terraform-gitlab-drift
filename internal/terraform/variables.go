package terraform

import (
	"io"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	gl "gitlab.com/gitlab-org/api/client-go"
)

func variableResourceName(owner, key, envScope string) string {
	name := owner + "_" + key
	if envScope != "*" {
		name += "_" + envScope
	}
	return normalizeName(name)
}

func groupVariableResourceName(g *gl.Group, v *gl.GroupVariable) string {
	return variableResourceName(g.Path, v.Key, v.EnvironmentScope)
}

func projectVariableResourceName(p *gl.Project, v *gl.ProjectVariable) string {
	return variableResourceName(projectResourceName(p), v.Key, v.EnvironmentScope)
}

func writeVariableAttrs(body *hclwrite.Body, key, value, varType, envScope, description string, protected, raw bool) {
	body.SetAttributeValue("key", cty.StringVal(key))
	body.SetAttributeValue("value", cty.StringVal(value))
	body.SetAttributeValue("variable_type", cty.StringVal(varType))
	body.SetAttributeValue("protected", cty.BoolVal(protected))
	body.SetAttributeValue("raw", cty.BoolVal(raw))
	if envScope != "*" {
		body.SetAttributeValue("environment_scope", cty.StringVal(envScope))
	}
	if description != "" {
		body.SetAttributeValue("description", cty.StringVal(description))
	}
}

func WriteGroupVariables(g *gl.Group, vars []*gl.GroupVariable, w io.Writer) error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	groupName := normalizeToTerraformName(g.Path)

	for i, v := range vars {
		if i > 0 {
			rootBody.AppendNewline()
		}
		name := groupVariableResourceName(g, v)
		block := rootBody.AppendNewBlock("resource", []string{"gitlab_group_variable", name})
		body := block.Body()

		body.SetAttributeTraversal("group", hcl.Traversal{
			hcl.TraverseRoot{Name: "gitlab_group"},
			hcl.TraverseAttr{Name: groupName},
			hcl.TraverseAttr{Name: "id"},
		})
		writeVariableAttrs(body, v.Key, v.Value, string(v.VariableType), v.EnvironmentScope, v.Description, v.Protected, v.Raw)
	}

	_, err := w.Write(f.Bytes())
	return err
}

func WriteProjectVariables(p *gl.Project, vars []*gl.ProjectVariable, w io.Writer) error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	projName := projectResourceName(p)

	for i, v := range vars {
		if i > 0 {
			rootBody.AppendNewline()
		}
		name := projectVariableResourceName(p, v)
		block := rootBody.AppendNewBlock("resource", []string{"gitlab_project_variable", name})
		body := block.Body()

		body.SetAttributeTraversal("project", hcl.Traversal{
			hcl.TraverseRoot{Name: "gitlab_project"},
			hcl.TraverseAttr{Name: projName},
			hcl.TraverseAttr{Name: "id"},
		})
		writeVariableAttrs(body, v.Key, v.Value, string(v.VariableType), v.EnvironmentScope, v.Description, v.Protected, v.Raw)
	}

	_, err := w.Write(f.Bytes())
	return err
}
