package terraform

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// hclString must produce a literal that parses back to the exact original
// string for any input — quotes, newlines, tabs, backslashes, and ${...}
// interpolation sequences included.
func TestHCLString_RoundTrips(t *testing.T) {
	inputs := []string{
		"plain",
		`he said "hi"`,
		"line1\nline2",
		"${var.x}",
		"tab\tend",
		`back\slash`,
		"",
	}
	for _, in := range inputs {
		lit := hclString(in)
		expr, diags := hclsyntax.ParseExpression([]byte(lit), "test.hcl", hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			t.Errorf("hclString(%q) = %q is not parseable: %v", in, lit, diags)
			continue
		}
		v, diags := expr.Value(nil)
		if diags.HasErrors() {
			t.Errorf("hclString(%q) = %q did not evaluate: %v", in, lit, diags)
			continue
		}
		if v.AsString() != in {
			t.Errorf("round-trip of %q via %q gave %q", in, lit, v.AsString())
		}
	}
}
