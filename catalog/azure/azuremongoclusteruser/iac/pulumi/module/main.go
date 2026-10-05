package module

import (
	"github.com/pkg/errors"
	azuremongoclusteruserv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremongoclusteruser/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mongocluster"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremongoclusteruserv1alpha1.AzureMongoClusterUserIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMongoClusterUser.Spec

	roles := mongocluster.UserRoleArray{}
	for _, role := range spec.Roles {
		roles = append(roles, &mongocluster.UserRoleArgs{
			Database: pulumi.String(role.Database),
			Name:     pulumi.String(role.Role),
		})
	}

	// Create the Entra access grant. "MicrosoftEntraID" is the identity
	// provider's only legal value today -- deliberately not part of the
	// spec; both engines send it explicitly. Every argument is
	// create-only (the resource has no update path); the target cluster
	// must allow MicrosoftEntraID authentication (deploy-time contract,
	// documented on the spec).
	createdUser, err := mongocluster.NewUser(ctx,
		locals.AzureMongoClusterUser.Metadata.Name,
		&mongocluster.UserArgs{
			ObjectId:             pulumi.String(spec.ObjectId.GetValue()),
			MongoClusterId:       pulumi.String(spec.MongoClusterId.GetValue()),
			IdentityProviderType: pulumi.String("MicrosoftEntraID"),
			PrincipalType:        pulumi.String(spec.PrincipalType),
			Roles:                roles,
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create mongo cluster user %s",
			locals.AzureMongoClusterUser.Metadata.Name)
	}

	ctx.Export(OpMongoClusterUserId, createdUser.ID())
	// The grant's ARM name IS the principal's object id.
	ctx.Export(OpMongoClusterUserName, createdUser.ObjectId)

	return nil
}
