package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpmemorystoreinstancev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmemorystoreinstance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig      *gcpprovider.GcpProviderConfig
	GcpMemorystoreInstance *gcpmemorystoreinstancev1alpha1.GcpMemorystoreInstance
	GcpLabels              map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpmemorystoreinstancev1alpha1.GcpMemorystoreInstanceIacInput) *Locals {
	locals := &Locals{}
	locals.GcpMemorystoreInstance = iacInput.Target

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpMemorystoreInstance.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpMemorystoreInstance.Spec.InstanceName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpMemorystoreInstance.String())

	if locals.GcpMemorystoreInstance.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpMemorystoreInstance.Metadata.Org
	}
	if locals.GcpMemorystoreInstance.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpMemorystoreInstance.Metadata.Env
	}
	if locals.GcpMemorystoreInstance.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpMemorystoreInstance.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
