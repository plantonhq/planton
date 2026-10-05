package module

import (
	"github.com/pkg/errors"
	azurefrontdoorsecuritypolicyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorsecuritypolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefrontdoorsecuritypolicyv1alpha1.AzureFrontDoorSecurityPolicyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFrontDoorSecurityPolicy.Spec

	domains := make(cdn.FrontdoorSecurityPolicySecurityPoliciesFirewallAssociationDomainArray, 0, len(locals.DomainIds))
	for _, domainId := range locals.DomainIds {
		domains = append(domains, &cdn.FrontdoorSecurityPolicySecurityPoliciesFirewallAssociationDomainArgs{
			CdnFrontdoorDomainId: pulumi.String(domainId),
		})
	}

	// The provider's securityPolicies -> firewall -> association nesting
	// is a one-choice ARM union (WebApplicationFirewall is the only
	// security-policy type); the spec models it flat and this block is
	// where the wrapper shape gets rebuilt.
	createdSecurityPolicy, err := cdn.NewFrontdoorSecurityPolicy(ctx,
		spec.SecurityPolicyName,
		&cdn.FrontdoorSecurityPolicyArgs{
			Name:                  pulumi.String(spec.SecurityPolicyName),
			CdnFrontdoorProfileId: pulumi.String(locals.ProfileId),
			SecurityPolicies: &cdn.FrontdoorSecurityPolicySecurityPoliciesArgs{
				Firewall: &cdn.FrontdoorSecurityPolicySecurityPoliciesFirewallArgs{
					CdnFrontdoorFirewallPolicyId: pulumi.String(locals.FirewallPolicyId),
					Association: &cdn.FrontdoorSecurityPolicySecurityPoliciesFirewallAssociationArgs{
						Domains: domains,
						// The service accepts exactly one pattern ("/*") --
						// a constant, not configuration; scope enforcement
						// by choosing WHICH domains to associate. NOTE the
						// bridge dialect: pulumi flattens azurerm's
						// one-item patterns list to a single string (the
						// Terraform module sends ["/*"]) -- same ARM
						// payload.
						PatternsToMatch: pulumi.String("/*"),
					},
				},
			},
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create front door security policy %s", spec.SecurityPolicyName)
	}

	// Export outputs. Nothing composes on a security policy (it is
	// itself the association); the id serves operational addressing.
	ctx.Export(OpSecurityPolicyId, createdSecurityPolicy.ID())
	ctx.Export(OpSecurityPolicyName, createdSecurityPolicy.Name)

	return nil
}
