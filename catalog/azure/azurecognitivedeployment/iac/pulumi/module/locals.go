package module

import (
	azurecognitivedeploymentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecognitivedeployment/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The deployment is an ARM child of its account: the provider's schema
// carries no location, resource group, or tags for it (ARM derives all
// three through the account), so these locals derive no tag map.
type Locals struct {
	AzureCognitiveDeployment *azurecognitivedeploymentv1alpha1.AzureCognitiveDeployment

	// CognitiveAccountId is a StringValueOrRef field; the platform
	// middleware resolves valueFrom references before IaC modules run, so
	// GetValue() always returns the resolved literal ARM ID.
	CognitiveAccountId string
}

// skuTierWire maps the spec's SKU tiers to the provider's wire values.
// Unspecified is absent -- the property is omitted so ARM derives the
// tier from the SKU name.
var skuTierWire = map[azurecognitivedeploymentv1alpha1.AzureCognitiveDeploymentSkuTier]string{
	azurecognitivedeploymentv1alpha1.AzureCognitiveDeploymentSkuTier_FREE:       "Free",
	azurecognitivedeploymentv1alpha1.AzureCognitiveDeploymentSkuTier_BASIC:      "Basic",
	azurecognitivedeploymentv1alpha1.AzureCognitiveDeploymentSkuTier_STANDARD:   "Standard",
	azurecognitivedeploymentv1alpha1.AzureCognitiveDeploymentSkuTier_PREMIUM:    "Premium",
	azurecognitivedeploymentv1alpha1.AzureCognitiveDeploymentSkuTier_ENTERPRISE: "Enterprise",
}

// versionUpgradeOptionWire maps the spec's version-upgrade options to the
// provider's wire values. Unspecified is absent and the provider applies
// its default, "OnceNewDefaultVersionAvailable".
var versionUpgradeOptionWire = map[azurecognitivedeploymentv1alpha1.AzureCognitiveDeploymentVersionUpgradeOption]string{
	azurecognitivedeploymentv1alpha1.AzureCognitiveDeploymentVersionUpgradeOption_ONCE_CURRENT_VERSION_EXPIRED:       "OnceCurrentVersionExpired",
	azurecognitivedeploymentv1alpha1.AzureCognitiveDeploymentVersionUpgradeOption_ONCE_NEW_DEFAULT_VERSION_AVAILABLE: "OnceNewDefaultVersionAvailable",
	azurecognitivedeploymentv1alpha1.AzureCognitiveDeploymentVersionUpgradeOption_NO_AUTO_UPGRADE:                    "NoAutoUpgrade",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurecognitivedeploymentv1alpha1.AzureCognitiveDeploymentIacInput) *Locals {
	locals := &Locals{}

	locals.AzureCognitiveDeployment = iacInput.Target
	locals.CognitiveAccountId = iacInput.Target.Spec.CognitiveAccountId.GetValue()

	return locals
}
