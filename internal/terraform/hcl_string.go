package terraform

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// hclString renders s as a safely-quoted, escaped HCL string literal (including
// the surrounding quotes), using hclwrite's encoder.
//
// The template-based writers (labels, memberships, pipeline schedules) build HCL
// with fmt.Fprintf. Interpolating a raw value into "%s" breaks the output when
// the value contains a quote, newline, or ${...} interpolation sequence (e.g. a
// label or schedule description, or a CI variable value). Wrapping such values
// in hclString produces a correct literal for any input.
func hclString(s string) string {
	return string(hclwrite.TokensForValue(cty.StringVal(s)).Bytes())
}
