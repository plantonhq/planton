package module

import (
	"github.com/pkg/errors"
	azurecognitiveaccountprojectv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecognitiveaccountproject/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cognitive"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecognitiveaccountprojectv1alpha1.AzureCognitiveAccountProjectIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureCognitiveAccountProject.Spec

	// Required by the provider: every project carries an identity -- it
	// is what the project's agents and evaluations act as.
	identityArgs := &cognitive.AccountProjectIdentityArgs{
		Type: pulumi.String(identityTypeWire[spec.Identity.Type]),
	}
	if len(spec.Identity.IdentityIds) > 0 {
		identityIds := pulumi.StringArray{}
		for _, identityId := range spec.Identity.IdentityIds {
			identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
		}
		identityArgs.IdentityIds = identityIds
	}

	// Create the AI Foundry project on its Azure AI services account --
	// the workspace a team's agents, evaluations and files live in. The
	// parent account must be kind "AIServices" with
	// project_management_enabled true; the FIRST project created on an
	// account becomes the account's default (the is_default output).
	//
	// ARM cannot UPDATE description or display_name to an EMPTY value --
	// the provider replaces the project when either is cleared (setting
	// or changing them updates in place).
	projectArgs := &cognitive.AccountProjectArgs{
		Name: pulumi.String(spec.Name),
		// The parent account. ForceNew.
		CognitiveAccountId: pulumi.String(locals.CognitiveAccountId),
		Location:           pulumi.String(spec.Region),
		Identity:           identityArgs,
		Tags:               pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.Description != "" {
		projectArgs.Description = pulumi.String(spec.Description)
	}
	if spec.DisplayName != "" {
		projectArgs.DisplayName = pulumi.String(spec.DisplayName)
	}

	createdProject, err := cognitive.NewAccountProject(ctx,
		spec.Name,
		projectArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create cognitive account project %s", spec.Name)
	}

	ctx.Export(OpProjectId, createdProject.ID())
	ctx.Export(OpProjectName, createdProject.Name)
	ctx.Export(OpEndpoints, createdProject.Endpoints)
	ctx.Export(OpIsDefault, createdProject.Default)
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdProject.Identity.PrincipalId())

	return nil
}
