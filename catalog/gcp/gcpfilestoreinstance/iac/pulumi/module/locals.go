package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpfilestoreinstancev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfilestoreinstance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig    *gcpprovider.GcpProviderConfig
	GcpFilestoreInstance *gcpfilestoreinstancev1alpha1.GcpFilestoreInstance
	GcpLabels            map[string]string
	// InstanceName is the cloud-side name: spec.instance_name when set,
	// metadata.name otherwise — the same explicit conditional as the
	// Terraform module, so both engines derive the identical name.
	InstanceName string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpfilestoreinstancev1alpha1.GcpFilestoreInstanceIacInput) *Locals {
	locals := &Locals{}
	locals.GcpFilestoreInstance = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig

	locals.InstanceName = iacInput.Target.Spec.InstanceName
	if locals.InstanceName == "" {
		locals.InstanceName = iacInput.Target.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range iacInput.Target.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.InstanceName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpFilestoreInstance.String())

	if locals.GcpFilestoreInstance.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpFilestoreInstance.Metadata.Org
	}
	if locals.GcpFilestoreInstance.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpFilestoreInstance.Metadata.Env
	}
	if locals.GcpFilestoreInstance.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpFilestoreInstance.Metadata.Id
	}

	return locals
}
