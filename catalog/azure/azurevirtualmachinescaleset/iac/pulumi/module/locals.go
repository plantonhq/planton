package module

import (
	"strings"

	azurevirtualmachinescalesetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualmachinescaleset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureVirtualMachineScaleSet *azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSet
	ResourceGroupName           string
	AzureTags                   map[string]string
	IsUniform                   bool
	IsLinux                     bool
	ComputerNamePrefix          string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetIacInput) *Locals {
	locals := &Locals{}

	locals.AzureVirtualMachineScaleSet = iacInput.Target
	target := iacInput.Target
	spec := target.Spec

	locals.ResourceGroupName = spec.ResourceGroup.GetValue()

	// The dispatch axes: ONE proto surface realizes onto azurerm's three
	// scale-set resources. Unset orchestration_mode applies FLEXIBLE
	// (Azure's recommendation for new workloads).
	locals.IsUniform = spec.OrchestrationMode == azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetOrchestrationMode_UNIFORM
	locals.IsLinux = spec.OsProfile.GetLinux() != nil

	// Instance computer names derive from this prefix (Azure appends a
	// unique suffix); unset defaults to the scale-set name.
	locals.ComputerNamePrefix = spec.OsProfile.ComputerNamePrefix

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureVirtualMachineScaleSet.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	for key, value := range spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
