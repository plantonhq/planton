package module

import (
	"strings"

	azureexpressrouteportv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureexpressrouteport/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureExpressRoutePort *azureexpressrouteportv1alpha1.AzureExpressRoutePort

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureexpressrouteportv1alpha1.AzureExpressRoutePortIacInput) *Locals {
	locals := &Locals{}

	locals.AzureExpressRoutePort = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureExpressRoutePort.String()),
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

// encapsulationWireValue maps the spec's encapsulation enum onto ARM's
// vocabulary ("Dot1Q", "QinQ").
func encapsulationWireValue(encapsulation azureexpressrouteportv1alpha1.AzureExpressRoutePortEncapsulation) string {
	switch encapsulation {
	case azureexpressrouteportv1alpha1.AzureExpressRoutePortEncapsulation_DOT1Q:
		return "Dot1Q"
	case azureexpressrouteportv1alpha1.AzureExpressRoutePortEncapsulation_QINQ:
		return "QinQ"
	}
	// Unreachable: the spec's encapsulation_required contract rejects
	// unspecified before the module runs.
	return ""
}

// billingTypeWireValue maps the spec's optional billing-type enum onto
// ARM's vocabulary, applying ARM's default (MeteredData) when the field
// is unset -- mirroring the Terraform variable default.
func billingTypeWireValue(billingType *azureexpressrouteportv1alpha1.AzureExpressRoutePortBillingType) string {
	if billingType == nil {
		return "MeteredData"
	}
	switch *billingType {
	case azureexpressrouteportv1alpha1.AzureExpressRoutePortBillingType_UNLIMITED_DATA:
		return "UnlimitedData"
	default:
		return "MeteredData"
	}
}

// macsecCipherWireValue maps the link's optional cipher enum onto ARM's
// vocabulary, applying ARM's default (GcmAes128) when the field is
// unset -- mirroring the Terraform variable default.
func macsecCipherWireValue(cipher *azureexpressrouteportv1alpha1.AzureExpressRoutePortMacsecCipher) string {
	if cipher == nil {
		return "GcmAes128"
	}
	switch *cipher {
	case azureexpressrouteportv1alpha1.AzureExpressRoutePortMacsecCipher_GCM_AES_256:
		return "GcmAes256"
	case azureexpressrouteportv1alpha1.AzureExpressRoutePortMacsecCipher_GCM_AES_XPN_128:
		return "GcmAesXpn128"
	case azureexpressrouteportv1alpha1.AzureExpressRoutePortMacsecCipher_GCM_AES_XPN_256:
		return "GcmAesXpn256"
	default:
		return "GcmAes128"
	}
}

// identityTypeWireValue maps the identity type enum's name onto ARM's
// comma-separated vocabulary.
func identityTypeWireValue(identityType azureexpressrouteportv1alpha1.AzureExpressRoutePortIdentityType) string {
	switch identityType {
	case azureexpressrouteportv1alpha1.AzureExpressRoutePortIdentityType_SYSTEM_ASSIGNED:
		return "SystemAssigned"
	case azureexpressrouteportv1alpha1.AzureExpressRoutePortIdentityType_USER_ASSIGNED:
		return "UserAssigned"
	case azureexpressrouteportv1alpha1.AzureExpressRoutePortIdentityType_SYSTEM_AND_USER_ASSIGNED:
		return "SystemAssigned, UserAssigned"
	}
	// Unreachable: the identity message requires a defined type.
	return ""
}
