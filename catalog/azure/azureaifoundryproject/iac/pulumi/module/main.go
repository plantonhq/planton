package module

import (
	"github.com/pkg/errors"
	azureaifoundryprojectv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureaifoundryproject/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/aifoundry"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureaifoundryprojectv1alpha1.AzureAiFoundryProjectIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureAiFoundryProject.Spec

	// Create the AI Foundry project -- ARM-wise an ML workspace of
	// kind "Project" linked to its hub. The project INHERITS the
	// hub's posture (vault, storage, insights, registry, network,
	// encryption) and deploys into the HUB's resource group -- the
	// provider derives the group from the hub reference, which is
	// why no resource-group argument exists here.
	projectArgs := &aifoundry.ProjectArgs{
		Name:            pulumi.String(spec.Name),
		Location:        pulumi.String(spec.Region),
		AiServicesHubId: pulumi.String(spec.AiServicesHubId.GetValue()),
		Tags:            pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.Identity != nil {
		identityArgs := &aifoundry.ProjectIdentityArgs{
			Type: pulumi.String(identityTypeWire[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		projectArgs.Identity = identityArgs
	}

	// Only legal alongside the identity block (spec CEL mirrors the
	// provider's RequiredWith).
	if spec.PrimaryUserAssignedIdentity.GetValue() != "" {
		projectArgs.PrimaryUserAssignedIdentity = pulumi.String(spec.PrimaryUserAssignedIdentity.GetValue())
	}

	// Sent only when true (both engines): the property is
	// Optional+Computed and the SERVICE flips it true when hub
	// encryption applies -- a pinned false would fight that
	// read-back. ForceNew.
	if spec.HighBusinessImpactEnabled {
		projectArgs.HighBusinessImpactEnabled = pulumi.Bool(true)
	}

	if spec.Description != "" {
		projectArgs.Description = pulumi.String(spec.Description)
	}
	if spec.FriendlyName != "" {
		projectArgs.FriendlyName = pulumi.String(spec.FriendlyName)
	}

	createdProject, err := aifoundry.NewProject(ctx,
		spec.Name,
		projectArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create ai foundry project %s", spec.Name)
	}

	ctx.Export(OpAiFoundryProjectId, createdProject.ID())
	ctx.Export(OpAiFoundryProjectName, createdProject.Name)
	ctx.Export(OpProjectGuid, createdProject.ProjectId)
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdProject.Identity.PrincipalId())

	return nil
}
