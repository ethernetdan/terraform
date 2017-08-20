package configschema

import (
	"fmt"

	"github.com/zclconf/go-zcl/zcldec"
)

// DecoderSpec returns a zcldec.Spec that can be used to decode a zcl Body
// using the facilities in the zcldec package.
//
// The returned specification is guaranteed to return a value of the type
// returned by method ImpliedType, but it may contain null values if any
// of the block attributes are defined as optional and/or computed.
func (b *Block) DecoderSpec() zcldec.Spec {
	ret := zcldec.ObjectSpec{}

	for name, attr := range b.Attributes {
		ret[name] = &zcldec.AttrSpec{
			Name:     name,
			Required: attr.Required,
			Type:     attr.Type,
		}
	}

	for name, nested := range b.BlockTypes {
		switch nested.Nesting {
		case NestingSingle:
			ret[name] = &zcldec.BlockSpec{
				TypeName: name,
				Nested:   nested.Block.DecoderSpec(),
			}
		case NestingList:
			ret[name] = &zcldec.BlockListSpec{
				TypeName: name,
				Nested:   nested.Block.DecoderSpec(),
			}
		case NestingSet:
			ret[name] = &zcldec.BlockSetSpec{
				TypeName: name,
				Nested:   nested.Block.DecoderSpec(),
			}
		case NestingMap:
			ret[name] = &zcldec.BlockMapSpec{
				TypeName:   name,
				Nested:     nested.Block.DecoderSpec(),
				LabelNames: []string{"key"}, // forced since configschema can't specify
			}

		default:
			// indicates caller error
			panic(fmt.Errorf("unsupported child block nesting mode %s for %q", nested.Nesting, name))
		}
	}

	return ret
}
