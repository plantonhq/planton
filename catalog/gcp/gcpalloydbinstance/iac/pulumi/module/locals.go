package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpalloydbinstancev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpalloydbinstance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig  *gcpprovider.GcpProviderConfig
	GcpAlloydbInstance *gcpalloydbinstancev1alpha1.GcpAlloydbInstance
	GcpLabels          map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpalloydbinstancev1alpha1.GcpAlloydbInstanceIacInput) *Locals {
	locals := &Locals{}
	locals.GcpAlloydbInstance = iacInput.Target
	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpAlloydbInstance.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpAlloydbInstance.Spec.InstanceId
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpAlloydbInstance.String())

	if locals.GcpAlloydbInstance.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpAlloydbInstance.Metadata.Org
	}
	if locals.GcpAlloydbInstance.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpAlloydbInstance.Metadata.Env
	}
	if locals.GcpAlloydbInstance.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpAlloydbInstance.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
