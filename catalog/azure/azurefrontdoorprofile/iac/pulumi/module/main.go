package module

import (
	"github.com/pkg/errors"
	azurefrontdoorprofilev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorprofile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFrontDoorProfile.Spec

	// The profile is deliberately just the container: endpoints, origin
	// groups, origins, and routes are their own kinds referencing this
	// profile's outputs, mirroring Azure's own child-resource model.
	// Front Door is a global service -- the provider forces location to
	// "global"; there is no region to configure.
	profileArgs := &cdn.FrontdoorProfileArgs{
		Name:              pulumi.String(spec.ProfileName),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		SkuName:           pulumi.String(locals.SkuName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Sent only when set: the platform materializes the documented
	// default (120 s) centrally, and Azure applies the same default when
	// the property is omitted.
	if spec.ResponseTimeoutSeconds != nil {
		profileArgs.ResponseTimeoutSeconds = pulumi.Int(int(spec.GetResponseTimeoutSeconds()))
	}

	// The managed identity is how Front Door reads customer-managed TLS
	// certificates from Key Vault without an access-policy secret. The
	// spec's CEL already guarantees identity ids are present exactly when
	// the type includes UserAssigned.
	if spec.Identity != nil {
		identityArgs := &cdn.FrontdoorProfileIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.UserAssignedIdentityIds) > 0 {
			identityIds := make(pulumi.StringArray, 0, len(spec.Identity.UserAssignedIdentityIds))
			for _, identityId := range spec.Identity.UserAssignedIdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		profileArgs.Identity = identityArgs
	}

	// Log scrubbing is enabled by the PRESENCE of rules (Azure semantics:
	// no rules == scrubbing disabled). The service supports only the
	// match-everything operator on profile scrubbing rules, so each entry
	// is just the request part to mask.
	if len(spec.LogScrubbingVariables) > 0 {
		scrubbingRules := make(cdn.FrontdoorProfileLogScrubbingRuleArray, 0, len(spec.LogScrubbingVariables))
		for _, matchVariable := range spec.LogScrubbingVariables {
			scrubbingRules = append(scrubbingRules, &cdn.FrontdoorProfileLogScrubbingRuleArgs{
				MatchVariable: pulumi.String(logScrubbingVariableStrings[matchVariable]),
			})
		}
		profileArgs.LogScrubbingRules = scrubbingRules
	}

	createdProfile, err := cdn.NewFrontdoorProfile(ctx,
		spec.ProfileName,
		profileArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create front door profile %s", spec.ProfileName)
	}

	// Export outputs. Client-facing hostnames deliberately live on
	// the AzureFrontDoorEndpoint kind's outputs, not here.
	ctx.Export(OpProfileId, createdProfile.ID())
	ctx.Export(OpProfileName, createdProfile.Name)
	ctx.Export(OpResourceGuid, createdProfile.ResourceGuid)
	// The principal id exists only when the identity block carries a
	// system-assigned identity; exported empty otherwise so the output
	// shape stays constant across configurations.
	ctx.Export(OpIdentityPrincipalId, createdProfile.Identity.ApplyT(func(identity *cdn.FrontdoorProfileIdentity) string {
		if identity == nil || identity.PrincipalId == nil {
			return ""
		}
		return *identity.PrincipalId
	}).(pulumi.StringOutput))

	return nil
}
