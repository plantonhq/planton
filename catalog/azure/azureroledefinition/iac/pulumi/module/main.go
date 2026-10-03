package module

import (
	"github.com/pkg/errors"
	azureroledefinitionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureroledefinition/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/authorization"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureroledefinitionv1alpha1.AzureRoleDefinitionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureRoleDefinition.Spec

	// Permission blocks map 1:1 from the spec. Azure treats multiple blocks
	// as a union of grants; the carve-out lists (not_*) trim only THIS role's
	// grant -- they are not deny rules.
	permissions := authorization.RoleDefinitionPermissionArray{}
	for _, p := range spec.Permissions {
		permissions = append(permissions, authorization.RoleDefinitionPermissionArgs{
			Actions:        pulumi.ToStringArray(p.Actions),
			NotActions:     pulumi.ToStringArray(p.NotActions),
			DataActions:    pulumi.ToStringArray(p.DataActions),
			NotDataActions: pulumi.ToStringArray(p.NotDataActions),
		})
	}

	// Lifecycle notes worth knowing before operating this resource:
	// - name, description, permissions, and assignable_scopes update IN
	//   PLACE; scope and the pinned GUID are the definition's ARM identity,
	//   so changing either replaces it (delete + create).
	// - Updates and deletes are eventually consistent: the bridged azurerm
	//   logic polls until Azure's records settle, so those operations take a
	//   few minutes. Azure refuses to delete a definition that still has role
	//   assignments -- destroy the assignments first (a composed
	//   environment's DAG reverse order does this naturally).
	definitionArgs := &authorization.RoleDefinitionArgs{
		Name:        pulumi.String(spec.Name),
		Scope:       pulumi.String(locals.Scope),
		Permissions: permissions,
	}

	// Optional strings are set only when non-empty: passing an empty string
	// where the provider expects "unset" would make Azure validate the empty
	// value instead of applying its own defaulting.
	if spec.Description != "" {
		definitionArgs.Description = pulumi.String(spec.Description)
	}
	// When omitted, azurerm defaults the assignable scopes to [scope] -- the
	// same server-side defaulting the Terraform module gets by passing null,
	// so both engines deploy identical definitions for an identical spec.
	if len(locals.AssignableScopes) > 0 {
		definitionArgs.AssignableScopes = pulumi.ToStringArray(locals.AssignableScopes)
	}
	// The ARM resource name is a GUID; omitted means Azure generates one at
	// create time. Pinning it keeps the definition's full ARM ID stable
	// across replacements.
	if spec.RoleDefinitionId != "" {
		definitionArgs.RoleDefinitionId = pulumi.String(spec.RoleDefinitionId)
	}

	definition, err := authorization.NewRoleDefinition(ctx,
		locals.AzureRoleDefinition.Metadata.Name,
		definitionArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create role definition %s", locals.AzureRoleDefinition.Metadata.Name)
	}

	// Export outputs from the created resource (not the spec) so they
	// carry the values Azure resolved -- the generated GUID and the defaulted
	// assignable scopes in particular. The fully-scoped ARM ID
	// (role_definition_resource_id in azurerm terms) is exported as
	// role_definition_id because that is the form Planton's Azure surface
	// consistently uses for role definition IDs -- it is exactly what an
	// AzureRoleAssignment's role_definition_id field consumes.
	ctx.Export(OpRoleDefinitionId, definition.RoleDefinitionResourceId)
	ctx.Export(OpRoleDefinitionGuid, definition.RoleDefinitionId)
	ctx.Export(OpRoleName, definition.Name)
	ctx.Export(OpScope, definition.Scope)
	ctx.Export(OpAssignableScopes, definition.AssignableScopes)

	return nil
}
