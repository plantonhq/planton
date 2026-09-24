package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpprivatecapoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecapool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpPrivateCaPool  *gcpprivatecapoolv1alpha1.GcpPrivateCaPool
	GcpLabels         map[string]string

	// CaPoolId is spec.ca_pool_id when set, otherwise metadata.name --
	// identical to the Terraform module's locals.ca_pool_id.
	CaPoolId string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpprivatecapoolv1alpha1.GcpPrivateCaPoolStackInput) *Locals {
	locals := &Locals{}
	locals.GcpPrivateCaPool = stackInput.Target
	metadata := locals.GcpPrivateCaPool.Metadata
	spec := locals.GcpPrivateCaPool.Spec

	locals.CaPoolId = spec.CaPoolId
	if locals.CaPoolId == "" {
		locals.CaPoolId = metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpPrivateCaPool.String())

	if metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = metadata.Org
	}
	if metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = metadata.Env
	}
	if metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}

// lastSegment returns what follows the last "/" -- the bare ID of a full
// resource path, or the value itself when it is already bare.
func lastSegment(value string) string {
	return value[strings.LastIndex(value, "/")+1:]
}
