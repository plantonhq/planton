package module

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	azurecosmosdbsqlroleassignmentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecosmosdbsqlroleassignment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cosmosdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecosmosdbsqlroleassignmentv1alpha1.AzureCosmosdbSqlRoleAssignmentIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureCosmosdbSqlRoleAssignment.Spec

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

	// The Terraform provider validates the role definition ID's shape
	// at plan time; mirror that contract here so a malformed literal
	// fails before any ARM call on both engines.
	if err := validateRoleDefinitionId(locals.RoleDefinitionId); err != nil {
		return err
	}

	// No Azure tags: ARM does not support tags on Cosmos child
	// resources, so the platform's identity tags live on the account.
	roleAssignmentArgs := &cosmosdb.SqlRoleAssignmentArgs{
		ResourceGroupName: pulumi.String(resourceGroupName),
		AccountName:       pulumi.String(accountName),
		RoleDefinitionId:  pulumi.String(locals.RoleDefinitionId),
		PrincipalId:       pulumi.String(locals.PrincipalId),
		Scope:             pulumi.String(locals.Scope),
	}

	// The assignment's ARM resource name is a GUID. When the spec does
	// not pin one, generate it here -- mirroring azurerm's create-time
	// UUID generation. Pulumi-azure autogenerates from the logical
	// resource name when Name is unset, which produces invalid non-UUID
	// values and must not be relied on.
	assignmentGuid := spec.Name
	if assignmentGuid == "" {
		assignmentGuid = uuid.New().String()
	}
	roleAssignmentArgs.Name = pulumi.String(assignmentGuid)

	// pulumi-azure enforces a 24-character logical resource name on
	// SqlRoleAssignment. Each Planton stack deploys exactly one
	// assignment, so "main" mirrors the Terraform module's single-
	// resource name; the Azure GUID lives in Name above.
	createdRoleAssignment, err := cosmosdb.NewSqlRoleAssignment(ctx,
		"main",
		roleAssignmentArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create cosmosdb sql role assignment for principal %s", locals.PrincipalId)
	}

	// Export outputs -- the grant record's ARM identity.
	ctx.Export(OpRoleAssignmentId, createdRoleAssignment.ID())
	ctx.Export(OpRoleAssignmentGuid, createdRoleAssignment.Name)
	ctx.Export(OpCosmosdbAccountName, pulumi.String(accountName))

	return nil
}
