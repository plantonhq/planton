package module

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	azurecosmosdbsqlroledefinitionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecosmosdbsqlroledefinition/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cosmosdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// roleTypeMap maps the spec's closed type enum to ARM's exact wire
// vocabulary (the provider validates these two strings case-sensitively).
// Unspecified sends nothing so the provider's own CustomRole default
// applies -- identical behavior to the Terraform module.
var roleTypeMap = map[azurecosmosdbsqlroledefinitionv1alpha1.AzureCosmosdbSqlRoleDefinitionType]string{
	azurecosmosdbsqlroledefinitionv1alpha1.AzureCosmosdbSqlRoleDefinitionType_CUSTOM_ROLE:   "CustomRole",
	azurecosmosdbsqlroledefinitionv1alpha1.AzureCosmosdbSqlRoleDefinitionType_BUILT_IN_ROLE: "BuiltInRole",
}

func Resources(ctx *pulumi.Context, iacInput *azurecosmosdbsqlroledefinitionv1alpha1.AzureCosmosdbSqlRoleDefinitionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureCosmosdbSqlRoleDefinition.Spec

	// The provider addresses Cosmos RBAC resources by the (resource
	// group, account, GUID) trio rather than an ARM ID, so both names
	// are parsed from the resolved account ID -- the spec models a
	// single parent reference and the module derives the rest (no
	// redundant, contradictable state). Parsing matches the Terraform
	// module's anchored regexes, so a malformed ID fails loudly here
	// instead of computing wrong names.
	accountName, resourceGroupName, err := parseCosmosdbAccountId(locals.CosmosdbAccountId)
	if err != nil {
		return err
	}

	// WHERE assignments of this role may be created: the account itself
	// or database/container paths under it. References resolve to
	// literal paths before the module runs.
	assignableScopes := pulumi.StringArray{}
	for _, scope := range spec.AssignableScopes {
		assignableScopes = append(assignableScopes, pulumi.String(scope.GetValue()))
	}

	// WHAT the role allows. Blocks are additive (a union). Cosmos
	// supports ALLOW rules only -- no not_data_actions carve-out exists
	// in this RBAC system.
	permissions := cosmosdb.SqlRoleDefinitionPermissionArray{}
	for _, permission := range spec.Permissions {
		dataActions := pulumi.StringArray{}
		for _, action := range permission.DataActions {
			dataActions = append(dataActions, pulumi.String(action))
		}
		permissions = append(permissions, cosmosdb.SqlRoleDefinitionPermissionArgs{
			DataActions: dataActions,
		})
	}

	// No Azure tags: ARM does not support tags on Cosmos child
	// resources, so the platform's identity tags live on the account.
	roleDefinitionArgs := &cosmosdb.SqlRoleDefinitionArgs{
		Name:              pulumi.String(spec.RoleName),
		ResourceGroupName: pulumi.String(resourceGroupName),
		AccountName:       pulumi.String(accountName),
		AssignableScopes:  assignableScopes,
		Permissions:       permissions,
	}

	// The definition's ARM resource name is a GUID. When the spec does
	// not pin one, generate it here -- mirroring azurerm's create-time
	// UUID generation (Pulumi-azure must not autogenerate from the
	// logical resource name).
	roleDefinitionGuid := spec.RoleDefinitionId
	if roleDefinitionGuid == "" {
		roleDefinitionGuid = uuid.New().String()
	}
	roleDefinitionArgs.RoleDefinitionId = pulumi.String(roleDefinitionGuid)

	// Unset sends nothing so the provider's own CustomRole default
	// applies -- the only type organizations author (built-in
	// definitions already exist in every account).
	if roleType, ok := roleTypeMap[spec.Type]; ok {
		roleDefinitionArgs.Type = pulumi.String(roleType)
	}

	createdRoleDefinition, err := cosmosdb.NewSqlRoleDefinition(ctx,
		locals.AzureCosmosdbSqlRoleDefinition.Metadata.Name,
		roleDefinitionArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create cosmosdb sql role definition %s", spec.RoleName)
	}

	// Export outputs. The fully-scoped ARM ID is the composition
	// seam: it is exactly what an AzureCosmosdbSqlRoleAssignment's
	// role_definition_id field consumes.
	ctx.Export(OpRoleDefinitionId, createdRoleDefinition.ID())
	ctx.Export(OpRoleDefinitionGuid, createdRoleDefinition.RoleDefinitionId)
	ctx.Export(OpRoleName, createdRoleDefinition.Name)
	ctx.Export(OpCosmosdbAccountName, pulumi.String(accountName))

	return nil
}
