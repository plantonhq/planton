package module

import (
	"github.com/pkg/errors"
	azureroleassignmentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureroleassignment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/authorization"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureroleassignmentv1alpha1.AzureRoleAssignmentIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureRoleAssignment.Spec

	// Every field on a role assignment is immutable in Azure (the ARM object is
	// an atomic grant record), so any spec change replaces the assignment.
	// That is the correct, expected behavior -- no lifecycle games needed here.
	assignmentArgs := &authorization.AssignmentArgs{
		Scope:       pulumi.String(locals.Scope),
		PrincipalId: pulumi.String(locals.PrincipalId),
		// azurerm rejects a request carrying both role coordinates; the spec's
		// exactly-one-of validation guarantees only one is non-empty here.
		SkipServicePrincipalAadCheck: pulumi.Bool(spec.SkipServicePrincipalAadCheck),
	}

	// Optional strings are set only when non-empty: passing an empty string
	// where the provider expects "unset" would make Azure validate the empty
	// value (e.g. an empty role_definition_name fails the role lookup) instead
	// of applying its own defaulting.
	if spec.RoleDefinitionName != "" {
		assignmentArgs.RoleDefinitionName = pulumi.String(spec.RoleDefinitionName)
	}
	// role_definition_id is a StringValueOrRef (an AzureRoleDefinition's
	// role_definition_id output, or a literal ID); by module time the
	// platform has resolved any reference, so the flat value is read here.
	if spec.GetRoleDefinitionId().GetValue() != "" {
		assignmentArgs.RoleDefinitionId = pulumi.String(spec.GetRoleDefinitionId().GetValue())
	}
	if locals.PrincipalType != "" {
		assignmentArgs.PrincipalType = pulumi.String(locals.PrincipalType)
	}
	if spec.Description != "" {
		assignmentArgs.Description = pulumi.String(spec.Description)
	}
	// When condition is set without a version, the provider applies Azure's
	// default version ("2.0") -- both engines inherit that same defaulting from
	// azurerm, so the deployed result is identical either way.
	if spec.Condition != "" {
		assignmentArgs.Condition = pulumi.String(spec.Condition)
	}
	if spec.ConditionVersion != "" {
		assignmentArgs.ConditionVersion = pulumi.String(spec.ConditionVersion)
	}
	if spec.DelegatedManagedIdentityResourceId != "" {
		assignmentArgs.DelegatedManagedIdentityResourceId = pulumi.String(spec.DelegatedManagedIdentityResourceId)
	}
	// The ARM resource name is a GUID; omitted means Azure generates one at
	// create time. Pinning it keeps the assignment's full ARM ID stable across
	// replacements.
	if spec.Name != "" {
		assignmentArgs.Name = pulumi.String(spec.Name)
	}

	assignment, err := authorization.NewAssignment(ctx,
		locals.AzureRoleAssignment.Metadata.Name,
		assignmentArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create role assignment %s", locals.AzureRoleAssignment.Metadata.Name)
	}

	// Export outputs. RoleDefinitionId and PrincipalType are exported
	// from the created resource (not the spec) so they carry the values Azure
	// resolved -- the definition ID behind a role name, the inferred principal
	// type -- which downstream automation can rely on.
	ctx.Export(OpRoleAssignmentId, assignment.ID())
	ctx.Export(OpName, assignment.Name)
	ctx.Export(OpScope, assignment.Scope)
	ctx.Export(OpRoleDefinitionId, assignment.RoleDefinitionId)
	ctx.Export(OpPrincipalId, assignment.PrincipalId)
	ctx.Export(OpPrincipalType, assignment.PrincipalType)

	return nil
}
