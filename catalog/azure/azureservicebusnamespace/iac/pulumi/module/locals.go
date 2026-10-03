package module

import (
	"strings"

	azureservicebusnamespacev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureservicebusnamespace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureServiceBusNamespace *azureservicebusnamespacev1alpha1.AzureServiceBusNamespace
	ResourceGroupName        string
	AzureTags                map[string]string
}

// skuStrings maps the spec's SKU enum to ARM's wire values. The unspecified
// row deploys STANDARD -- the full-featured multi-tenant tier -- because an
// unmapped enum would send the empty string, which the provider rejects.
var skuStrings = map[azureservicebusnamespacev1alpha1.AzureServiceBusNamespaceSku]string{
	azureservicebusnamespacev1alpha1.AzureServiceBusNamespaceSku_azure_service_bus_namespace_sku_unspecified: "Standard",
	azureservicebusnamespacev1alpha1.AzureServiceBusNamespaceSku_BASIC:                                       "Basic",
	azureservicebusnamespacev1alpha1.AzureServiceBusNamespaceSku_STANDARD:                                    "Standard",
	azureservicebusnamespacev1alpha1.AzureServiceBusNamespaceSku_PREMIUM:                                     "Premium",
}

// identityTypeStrings maps the identity-model enum to ARM's values --
// Service Bus namespaces support all three managed-identity models.
var identityTypeStrings = map[azureservicebusnamespacev1alpha1.AzureServiceBusNamespaceIdentityType]string{
	azureservicebusnamespacev1alpha1.AzureServiceBusNamespaceIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azureservicebusnamespacev1alpha1.AzureServiceBusNamespaceIdentityType_USER_ASSIGNED:            "UserAssigned",
	azureservicebusnamespacev1alpha1.AzureServiceBusNamespaceIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

// networkDefaultActionStrings maps the firewall's default-action enum to
// ARM's values. The unspecified row keeps Azure's open default (Allow) --
// the block may be declared just for trusted-services or public-access
// dials.
var networkDefaultActionStrings = map[azureservicebusnamespacev1alpha1.AzureServiceBusNetworkDefaultAction]string{
	azureservicebusnamespacev1alpha1.AzureServiceBusNetworkDefaultAction_azure_service_bus_network_default_action_unspecified: "Allow",
	azureservicebusnamespacev1alpha1.AzureServiceBusNetworkDefaultAction_ALLOW:                                                "Allow",
	azureservicebusnamespacev1alpha1.AzureServiceBusNetworkDefaultAction_DENY:                                                 "Deny",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureservicebusnamespacev1alpha1.AzureServiceBusNamespaceIacInput) *Locals {
	locals := &Locals{}

	locals.AzureServiceBusNamespace = iacInput.Target
	target := iacInput.Target

	// The resource_group field is a StringValueOrRef. The platform middleware
	// resolves valueFrom references before IaC modules run, so .GetValue()
	// always returns the resolved literal string.
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Identity tags derived from metadata; user tags merge OVER these (the
	// governance surface belongs to the user).
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureServiceBusNamespace.String()),
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

	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
