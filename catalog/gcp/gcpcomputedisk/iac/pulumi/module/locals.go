package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcomputediskv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcomputedisk/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpComputeDisk    *gcpcomputediskv1alpha1.GcpComputeDisk
	GcpLabels         map[string]string
	// DiskName is the cloud-side name: spec.disk_name when set,
	// metadata.name otherwise — the same explicit conditional as the
	// Terraform module, so both engines derive the identical name.
	DiskName string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcomputediskv1alpha1.GcpComputeDiskIacInput) *Locals {
	locals := &Locals{}
	locals.GcpComputeDisk = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig

	locals.DiskName = iacInput.Target.Spec.DiskName
	if locals.DiskName == "" {
		locals.DiskName = iacInput.Target.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range iacInput.Target.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.DiskName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpComputeDisk.String())

	if locals.GcpComputeDisk.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpComputeDisk.Metadata.Org
	}
	if locals.GcpComputeDisk.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpComputeDisk.Metadata.Env
	}
	if locals.GcpComputeDisk.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpComputeDisk.Metadata.Id
	}

	return locals
}
