package config

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/hashicorp/hil/ast"
	"github.com/hashicorp/terraform/config/configschema"
	"github.com/zclconf/go-zcl/zcl"
	"github.com/zclconf/go-zcl/zcldec"
)

// InterpolateZcl is the experimental alternative to RawConfig.Interpolate
// when zcl parsing is active.
func InterpolateZcl(body zcl.Body, schema *configschema.Block, vars map[string]ast.Variable) (map[string]interface{}, error) {
	panic("InterpolateZcl is not yet implemented")
}

// DetectVariablesZcl is the experimental alternative to RawConfig.Variables
// when zcl parsing is active.
func DetectVariablesZcl(body zcl.Body, schema *configschema.Block) map[string]InterpolatedVariable {
	spec := schema.DecoderSpec()
	vars := zcldec.Variables(body, spec)
	ret := map[string]InterpolatedVariable{}

	// Convert our traversals into strings that match what our existing
	// code expects, and then create InterpolatedVariable values for them.
Variable:
	for _, traversal := range vars {
		var buf bytes.Buffer
		dot := false

	Step:
		for _, stepI := range traversal {

			if dot {
				buf.WriteByte('.')
			}
			dot = true

			switch step := stepI.(type) {
			case zcl.TraverseRoot:
				buf.WriteString(step.Name)
			case zcl.TraverseAttr:
				buf.WriteString(step.Name)
			case *zcl.TraverseIndex:
				switch step.Key.Type() {
				case cty.Number:
					var index int
					err := gocty.FromCtyValue(step.Key, &index)
					if err != nil {
						// ignore invalid indices; we'll catch them during
						// interpolation and report an error there.
						continue Variable
					}

					buf.WriteString(strconv.Itoa(index))
				case cty.String:
					// In this lowering we lose information about whether
					// we had an attr or an index, but the underlying variable
					// resolution code doesn't care anyway aside from some
					// weirdness with keys containing dots, which will just
					// be a hazard for now until we rewrite the variable
					// processing to operate directly on a zcl.Traversal.
					buf.WriteString(step.Key.AsString())
				default:
					// ignore invalid indices; we'll catch them during
					// interpolation and report an error there.
					continue Variable
				}
			default:
				// if we find a step we don't understand, stop processing
				// and try to make the best of what we've found so far.
				break Step
			}
		}

		name := buf.String()
		if name[len(name)-1] == '.' {
			// if we aborted early above, we will have an errant trailing period
			name = name[:len(name)-1]
		}

		v, err := NewInterpolatedVariable(name)
		if err != nil {
			fmt.Printf("error for %s: %s\n", name, err)
			continue Variable
		}
		ret[name] = v
	}

	return ret
}
