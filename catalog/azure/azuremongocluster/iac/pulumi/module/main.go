package module

import (
	"fmt"

	"github.com/pkg/errors"
	azuremongoclusterv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremongocluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mongocluster"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremongoclusterv1alpha1.AzureMongoClusterIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMongoCluster.Spec

	// Platform default "Default" -- always sent, so the rendered plan
	// states the mode. (Proto defaults applied here, matching the TF
	// module's coalesce.)
	createMode := "Default"
	if spec.CreateMode != nil {
		createMode = *spec.CreateMode
	}

	// Platform default true, mapped to the provider's Enabled/Disabled
	// tokens -- always sent (mirrors Azure's own default).
	publicNetworkAccessEnabled := true
	if spec.PublicNetworkAccessEnabled != nil {
		publicNetworkAccessEnabled = *spec.PublicNetworkAccessEnabled
	}

	// Create the Mongo vCore cluster. The provider owns the mode
	// machinery this module deliberately does not reimplement: Default
	// mode stages the Data API in a separate post-create update,
	// upgrades away from Free/M25 stage a tier-first update, and a
	// create_mode change is forced to a replacement (Azure never returns
	// the mode on reads).
	clusterArgs := &mongocluster.MongoClusterArgs{
		Name:                pulumi.String(spec.Name),
		ResourceGroupName:   pulumi.String(spec.ResourceGroup.GetValue()),
		Location:            pulumi.String(spec.Region),
		CreateMode:          pulumi.String(createMode),
		PublicNetworkAccess: pulumi.String(publicNetworkAccessToken(publicNetworkAccessEnabled)),
		Tags:                pulumi.ToStringMap(locals.AzureTags),
	}

	// The native administrator pair travels together (spec CEL mirrors
	// the provider's RequiredWith); replicas and restores inherit the
	// source's administrator, so both stay unsent when unset.
	if spec.AdministratorUsername != "" {
		clusterArgs.AdministratorUsername = pulumi.String(spec.AdministratorUsername)
		clusterArgs.AdministratorPassword = pulumi.String(spec.AdministratorPassword.GetValue())
	}

	// Sizing fields are sent only when set: replica and restore modes
	// inherit them from the source, and the provider sends each to ARM
	// only when it carries a value.
	if spec.Version != nil {
		clusterArgs.Version = pulumi.String(spec.GetVersion())
	}
	if spec.ComputeTier != nil {
		clusterArgs.ComputeTier = pulumi.String(spec.GetComputeTier())
	}
	if spec.StorageSizeInGb != nil {
		clusterArgs.StorageSizeInGb = pulumi.Int(int(spec.GetStorageSizeInGb()))
		// The storage type rides the size (the provider's RequiredWith):
		// platform default "PremiumSSD" when the block is sent at all.
		storageType := "PremiumSSD"
		if spec.StorageType != nil {
			storageType = *spec.StorageType
		}
		clusterArgs.StorageType = pulumi.String(storageType)
	}
	if spec.ShardCount != nil {
		clusterArgs.ShardCount = pulumi.Int(int(spec.GetShardCount()))
	}
	if spec.HighAvailabilityMode != nil {
		clusterArgs.HighAvailabilityMode = pulumi.String(spec.GetHighAvailabilityMode())
	}

	// Sent only when set: Azure defaults an unset list to ["NativeAuth"]
	// server-side (the provider models the argument Optional+Computed for
	// exactly that reason).
	if len(spec.AuthenticationMethods) > 0 {
		clusterArgs.AuthenticationMethods = pulumi.ToStringArray(spec.AuthenticationMethods)
	}

	// User-assigned is the only identity flavor the service supports on
	// this resource. Adding the first identity or removing the last one
	// is a REPLACEMENT (the provider forces it -- Azure rejects the
	// in-place transition); documented on the spec field.
	if len(spec.UserAssignedIdentityIds) > 0 {
		identityIds := pulumi.StringArray{}
		for _, identityId := range spec.UserAssignedIdentityIds {
			identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
		}
		clusterArgs.Identity = &mongocluster.MongoClusterIdentityArgs{
			Type:        pulumi.String("UserAssigned"),
			IdentityIds: identityIds,
		}
	}

	if spec.CustomerManagedKey != nil {
		clusterArgs.CustomerManagedKey = &mongocluster.MongoClusterCustomerManagedKeyArgs{
			KeyVaultKeyId:          pulumi.String(spec.CustomerManagedKey.KeyVaultKeyId.GetValue()),
			UserAssignedIdentityId: pulumi.String(spec.CustomerManagedKey.UserAssignedIdentityId.GetValue()),
		}
	}

	if len(spec.PreviewFeatures) > 0 {
		clusterArgs.PreviewFeatures = pulumi.ToStringArray(spec.PreviewFeatures)
	}

	// GeoReplica coordinates -- the spec CELs mirror the provider's
	// create-time contract (both required in GeoReplica mode; location
	// requires the server id).
	if spec.SourceServerId.GetValue() != "" {
		clusterArgs.SourceServerId = pulumi.String(spec.SourceServerId.GetValue())
	}
	if spec.SourceLocation != "" {
		clusterArgs.SourceLocation = pulumi.String(spec.SourceLocation)
	}

	if spec.Restore != nil {
		clusterArgs.Restore = &mongocluster.MongoClusterRestoreArgs{
			PointInTimeUtc: pulumi.String(spec.Restore.PointInTimeUtc),
			SourceId:       pulumi.String(spec.Restore.SourceId.GetValue()),
		}
	}

	// Sent only when the manifest carries it: the provider ERRORS when
	// the raw config sets this on a non-Default-mode cluster (even
	// false), and the spec CEL front-loads that contract.
	if spec.DataApiModeEnabled != nil {
		clusterArgs.DataApiModeEnabled = pulumi.BoolPtr(spec.GetDataApiModeEnabled())
	}

	createdCluster, err := mongocluster.NewMongoCluster(ctx,
		locals.AzureMongoCluster.Metadata.Name,
		clusterArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create mongo cluster %s",
			locals.AzureMongoCluster.Metadata.Name)
	}

	// Composed firewall rules -- one provider resource per named rule,
	// keyed by the rule's name (renames replace only that rule, sibling
	// rules stay untouched), in lockstep with the TF module's for_each.
	for _, rule := range spec.FirewallRules {
		_, err := mongocluster.NewFirewallRule(ctx,
			fmt.Sprintf("%s-%s", locals.AzureMongoCluster.Metadata.Name, rule.Name),
			&mongocluster.FirewallRuleArgs{
				Name:           pulumi.String(rule.Name),
				MongoClusterId: createdCluster.ID(),
				StartIpAddress: pulumi.String(rule.StartIpAddress),
				EndIpAddress:   pulumi.String(rule.EndIpAddress),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create mongo cluster firewall rule %s", rule.Name)
		}
	}

	ctx.Export(OpMongoClusterId, createdCluster.ID())
	ctx.Export(OpMongoClusterName, createdCluster.Name)
	// Azure substitutes the administrator credentials into the
	// <user>:<password> placeholder (the provider does the substitution;
	// empty without a native administrator). The first entry is Azure's
	// primary/global connection string -- the singular chart edge; the
	// map carries every published variant keyed by Azure's name for it.
	ctx.Export(OpConnectionString, pulumi.ToSecret(createdCluster.ConnectionStrings.ApplyT(func(connectionStrings []mongocluster.MongoClusterConnectionString) string {
		if len(connectionStrings) == 0 || connectionStrings[0].Value == nil {
			return ""
		}
		return *connectionStrings[0].Value
	}).(pulumi.StringOutput)))
	ctx.Export(OpConnectionStrings, pulumi.ToSecret(createdCluster.ConnectionStrings.ApplyT(func(connectionStrings []mongocluster.MongoClusterConnectionString) map[string]string {
		result := map[string]string{}
		for _, connectionString := range connectionStrings {
			if connectionString.Name == nil || connectionString.Value == nil {
				continue
			}
			result[*connectionString.Name] = *connectionString.Value
		}
		return result
	}).(pulumi.StringMapOutput)))

	return nil
}

// publicNetworkAccessToken maps the spec's bool onto the provider's
// Enabled/Disabled vocabulary.
func publicNetworkAccessToken(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}
