package module

import (
	"strings"

	azurenetworkinterfacev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurenetworkinterface/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureNetworkInterface *azurenetworkinterfacev1alpha1.AzureNetworkInterface

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// NetworkSecurityGroupId is the resolved ARM ID of the NIC-level NSG,
	// or empty when the spec attaches none.
	NetworkSecurityGroupId string

	// AuxiliaryMode and AuxiliarySku are the ARM strings for the spec's
	// enums, or empty when unspecified so both engines send nothing (the
	// non-appliance default).
	AuxiliaryMode string
	AuxiliarySku  string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurenetworkinterfacev1alpha1.AzureNetworkInterfaceIacInput) *Locals {
	locals := &Locals{}

	locals.AzureNetworkInterface = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()
	locals.NetworkSecurityGroupId = target.Spec.NetworkSecurityGroupId.GetValue()

	switch target.Spec.AuxiliaryMode {
	case azurenetworkinterfacev1alpha1.AzureNetworkInterfaceAuxiliaryMode_ACCELERATED_CONNECTIONS:
		locals.AuxiliaryMode = "AcceleratedConnections"
	case azurenetworkinterfacev1alpha1.AzureNetworkInterfaceAuxiliaryMode_FLOATING:
		locals.AuxiliaryMode = "Floating"
	case azurenetworkinterfacev1alpha1.AzureNetworkInterfaceAuxiliaryMode_MAX_CONNECTIONS:
		locals.AuxiliaryMode = "MaxConnections"
	}

	switch target.Spec.AuxiliarySku {
	case azurenetworkinterfacev1alpha1.AzureNetworkInterfaceAuxiliarySku_A1:
		locals.AuxiliarySku = "A1"
	case azurenetworkinterfacev1alpha1.AzureNetworkInterfaceAuxiliarySku_A2:
		locals.AuxiliarySku = "A2"
	case azurenetworkinterfacev1alpha1.AzureNetworkInterfaceAuxiliarySku_A4:
		locals.AuxiliarySku = "A4"
	case azurenetworkinterfacev1alpha1.AzureNetworkInterfaceAuxiliarySku_A8:
		locals.AuxiliarySku = "A8"
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureNetworkInterface.String()),
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

	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
