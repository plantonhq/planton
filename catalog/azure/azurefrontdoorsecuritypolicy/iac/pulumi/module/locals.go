package module

import (
	azurefrontdoorsecuritypolicyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorsecuritypolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFrontDoorSecurityPolicy *azurefrontdoorsecuritypolicyv1alpha1.AzureFrontDoorSecurityPolicy
	ProfileId                    string
	FirewallPolicyId             string
	// DomainIds are the resolved endpoint / custom-domain ARM ids the
	// WAF protects (references are resolved to literals by the platform
	// before the module runs).
	DomainIds []string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurefrontdoorsecuritypolicyv1alpha1.AzureFrontDoorSecurityPolicyIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFrontDoorSecurityPolicy = iacInput.Target
	target := iacInput.Target

	locals.ProfileId = target.Spec.ProfileId.GetValue()
	locals.FirewallPolicyId = target.Spec.FirewallPolicyId.GetValue()

	for _, domainId := range target.Spec.DomainIds {
		locals.DomainIds = append(locals.DomainIds, domainId.GetValue())
	}

	// No Azure tags: ARM does not support tags on Front Door security
	// policies, so the platform's identity tags live on the profile.

	return locals
}
