package terraform

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/config/configschema"
)

// EvalGetResourceSchema is an EvalNode that tries to load the schema for
// a given resource.
//
// At present not all providers support resource schema, so the result may
// be a pointer to a nil pointer if a resource from such a provider is
// selected, and it is the caller's responsibility to handle its absense.
type EvalGetResourceSchema struct {
	Mode     config.ResourceMode
	Resource **Resource
	Provider *ResourceProvider
	Output   **configschema.Block
}

// TODO: test
func (n *EvalGetResourceSchema) Eval(ctx EvalContext) (interface{}, error) {
	provider := *n.Provider
	resourceType := (**n.Resource).Type

	switch n.Mode {
	case config.DataResourceMode:
		log.Printf("[TRACE] Getting schema for data source %s", resourceType)
		all := provider.DataSources()
		available := false
		for _, dsMeta := range all {
			if dsMeta.Name == resourceType {
				available = dsMeta.SchemaAvailable
				break
			}
		}
		if !available {
			// No schema available, presumably due to being from an old
			// provider that doesn't yet support the new schema API.
			// (We also get here if the provider doesn't know this data source,
			// which is actually a configuration error but one taken care of
			// elsewhere.)
			log.Printf("[TRACE] Schema not available for data source %s", resourceType)
			*n.Output = nil
			return nil, nil
		}

		schema, err := provider.DataSourceSchema(resourceType)
		if err != nil {
			return nil, err
		}

		*n.Output = schema
		log.Printf("[TRACE] Data source %s has schema %#v", resourceType, schema)
		return nil, nil

	case config.ManagedResourceMode:
		log.Printf("[TRACE] Getting schema for resource type %s", resourceType)
		all := provider.Resources()
		available := false
		for _, rMeta := range all {
			if rMeta.Name == resourceType {
				available = rMeta.SchemaAvailable
				break
			}
		}
		if !available {
			// No schema available, presumably due to being from an old
			// provider that doesn't yet support the new schema API.
			// (We also get here if the provider doesn't know this resource type,
			// which is actually a configuration error but one taken care of
			// elsewhere.)
			log.Printf("[TRACE] Schema not available for resource type %s", resourceType)
			*n.Output = nil
			return nil, nil
		}

		schema, err := provider.ResourceTypeSchema(resourceType)
		if err != nil {
			return nil, err
		}

		*n.Output = schema
		log.Printf("[TRACE] Resource type %s has schema %#v", resourceType, schema)
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported resource mode %s", n.Mode)
	}
}

// EvalInstanceInfo is an EvalNode implementation that fills in the
// InstanceInfo as much as it can.
type EvalInstanceInfo struct {
	Info *InstanceInfo
}

// TODO: test
func (n *EvalInstanceInfo) Eval(ctx EvalContext) (interface{}, error) {
	n.Info.ModulePath = ctx.Path()
	return nil, nil
}
