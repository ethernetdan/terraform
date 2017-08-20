package config

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-zcl/zclparse"

	"github.com/hashicorp/terraform/config/configschema"
)

func TestDetectVariablesZcl(t *testing.T) {
	src := `
literal = "hello"
interp  = "${var.interp}"
var     = var.var
list    = [var.list1, var.list2]

complex = test_resource.foo[1].bar["baz"]

splat = test_resource.splatted.*.id

for_expr = [for x in var.for_expr: [x, var.for_expr_elem]]

ignored_because_not_in_schema = var.ignored

block {
  in_block = var.in_block
}
`
	parser := zclparse.NewParser()
	f, diags := parser.ParseZCL([]byte(src), "")
	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics")
		for _, diag := range diags {
			t.Logf("- %s", diag.Error())
		}
	}

	body := f.Body
	schema := &configschema.Block{
		Attributes: map[string]*configschema.Attribute{
			"literal": {
				Type:     cty.String,
				Required: true,
			},
			"interp": {
				Type:     cty.String,
				Required: true,
			},
			"var": {
				Type:     cty.String,
				Required: true,
			},
			"list": {
				Type:     cty.List(cty.String),
				Required: true,
			},
			"complex": {
				Type:     cty.String,
				Required: true,
			},
			"splat": {
				Type:     cty.List(cty.String),
				Required: true,
			},
			"for_expr": {
				Type:     cty.List(cty.String),
				Required: true,
			},
		},
		BlockTypes: map[string]*configschema.NestedBlock{
			"block": {
				Nesting: configschema.NestingSingle,
				Block: configschema.Block{
					Attributes: map[string]*configschema.Attribute{
						"in_block": {
							Type:     cty.String,
							Required: true,
						},
					},
				},
			},
		},
	}

	got := DetectVariablesZcl(body, schema)
	want := map[string]InterpolatedVariable{
		"var.interp": &UserVariable{
			Name: "interp",
			Elem: "",
			key:  "var.interp",
		},
		"var.var": &UserVariable{
			Name: "var",
			Elem: "",
			key:  "var.var",
		},
		"var.list1": &UserVariable{
			Name: "list1",
			Elem: "",
			key:  "var.list1",
		},
		"var.list2": &UserVariable{
			Name: "list2",
			Elem: "",
			key:  "var.list2",
		},
		"test_resource.foo.1.bar": &ResourceVariable{
			Mode:  ManagedResourceMode,
			Type:  "test_resource",
			Name:  "foo",
			Index: 1,
			Field: "bar",
			Multi: true,
			key:   "test_resource.foo.1.bar",
		},
		"test_resource.splatted": &ResourceVariable{
			Mode:  ManagedResourceMode,
			Type:  "test_resource",
			Name:  "splatted",
			Index: -1,
			Field: "",
			Multi: true,
			key:   "test_resource.splatted",
		},
		"var.for_expr": &UserVariable{
			Name: "for_expr",
			Elem: "",
			key:  "var.for_expr",
		},
		"var.for_expr_elem": &UserVariable{
			Name: "for_expr_elem",
			Elem: "",
			key:  "var.for_expr_elem",
		},
		"var.in_block": &UserVariable{
			Name: "in_block",
			Elem: "",
			key:  "var.in_block",
		},
	}

	for k := range want {
		wantVar := want[k]
		gotVar := got[k]
		if !reflect.DeepEqual(gotVar, wantVar) {
			t.Errorf("wrong result for %s\ngot: %swant: %s", k, spew.Sdump(gotVar), spew.Sdump(wantVar))
		}
	}

	for k := range got {
		if _, exists := want[k]; !exists {
			t.Errorf("unwanted extra variable %s", k)
		}
	}
}
