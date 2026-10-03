package module

import (
	"strconv"
	"strings"

	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcomputeinstancev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcomputeinstance/v1alpha1"
)

// Locals holds handy references and derived values used across this module.
type Locals struct {
	GcpProviderConfig  *gcpprovider.GcpProviderConfig
	GcpComputeInstance *gcpcomputeinstancev1alpha1.GcpComputeInstance
	GcpLabels          map[string]string
	// InstanceName is the cloud-side name: spec.instance_name when set,
	// metadata.name otherwise — the same explicit conditional as the
	// Terraform module, so both engines derive the identical name.
	InstanceName string
}

// initializeLocals fills the Locals struct from the incoming IaC input.
func initializeLocals(iacInput *gcpcomputeinstancev1alpha1.GcpComputeInstanceIacInput) *Locals {
	locals := &Locals{}

	locals.GcpComputeInstance = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig

	target := iacInput.Target

	locals.InstanceName = target.Spec.InstanceName
	if locals.InstanceName == "" {
		locals.InstanceName = target.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range target.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = strconv.FormatBool(true)
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.InstanceName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpComputeInstance.String())

	if target.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = target.Metadata.Env
	}

	return locals
}
