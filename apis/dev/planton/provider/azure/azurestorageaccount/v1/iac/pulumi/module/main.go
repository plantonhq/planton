package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurestorageaccountv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurestorageaccount/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurestorageaccountv1.AzureStorageAccountStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	// Get the spec from locals
	spec := locals.AzureStorageAccount.Spec

	// Build network rules configuration
	var networkRulesArgs *storage.AccountNetworkRulesTypeArgs
	if spec.NetworkRules != nil {
		bypass := pulumi.StringArray{pulumi.String("None")}
		if spec.NetworkRules.GetBypassAzureServices() {
			bypass = pulumi.StringArray{pulumi.String("AzureServices")}
		}

		networkRulesArgs = &storage.AccountNetworkRulesTypeArgs{
			DefaultAction: pulumi.String(getNetworkDefaultAction(spec.NetworkRules.GetDefaultAction())),
			Bypasses:      bypass,
		}

		// Add IP rules
		if len(spec.NetworkRules.IpRules) > 0 {
			ipRules := make(pulumi.StringArray, 0)
			for _, ipRule := range spec.NetworkRules.IpRules {
				ipRules = append(ipRules, pulumi.String(ipRule))
			}
			networkRulesArgs.IpRules = ipRules
		}

		// Add VNet rules
		if len(spec.NetworkRules.VirtualNetworkSubnetIds) > 0 {
			vnetRules := make(pulumi.StringArray, 0)
			for _, subnetId := range spec.NetworkRules.VirtualNetworkSubnetIds {
				vnetRules = append(vnetRules, pulumi.String(subnetId))
			}
			networkRulesArgs.VirtualNetworkSubnetIds = vnetRules
		}
	} else {
		// Default network rules: Deny all with Azure Services bypass
		networkRulesArgs = &storage.AccountNetworkRulesTypeArgs{
			DefaultAction: pulumi.String("Deny"),
			Bypasses:      pulumi.StringArray{pulumi.String("AzureServices")},
		}
	}

	// Build blob properties configuration
	var blobPropertiesArgs *storage.AccountBlobPropertiesArgs
	if spec.BlobProperties != nil {
		blobPropertiesArgs = &storage.AccountBlobPropertiesArgs{
			VersioningEnabled: pulumi.Bool(spec.BlobProperties.GetEnableVersioning()),
		}

		// Add blob soft delete
		if spec.BlobProperties.GetSoftDeleteRetentionDays() > 0 {
			blobPropertiesArgs.DeleteRetentionPolicy = &storage.AccountBlobPropertiesDeleteRetentionPolicyArgs{
				Days: pulumi.Int(int(spec.BlobProperties.GetSoftDeleteRetentionDays())),
			}
		}

		// Add container soft delete
		if spec.BlobProperties.GetContainerSoftDeleteRetentionDays() > 0 {
			blobPropertiesArgs.ContainerDeleteRetentionPolicy = &storage.AccountBlobPropertiesContainerDeleteRetentionPolicyArgs{
				Days: pulumi.Int(int(spec.BlobProperties.GetContainerSoftDeleteRetentionDays())),
			}
		}
	} else {
		// Default blob properties
		blobPropertiesArgs = &storage.AccountBlobPropertiesArgs{
			VersioningEnabled: pulumi.Bool(false),
			DeleteRetentionPolicy: &storage.AccountBlobPropertiesDeleteRetentionPolicyArgs{
				Days: pulumi.Int(7),
			},
			ContainerDeleteRetentionPolicy: &storage.AccountBlobPropertiesContainerDeleteRetentionPolicyArgs{
				Days: pulumi.Int(7),
			},
		}
	}

	// Create the Storage Account
	storageAccount, err := storage.NewAccount(ctx,
		locals.StorageAccountName,
		&storage.AccountArgs{
			Name:                   pulumi.String(locals.StorageAccountName),
			Location:               pulumi.String(spec.Region),
			ResourceGroupName:      pulumi.String(locals.ResourceGroupName),
			AccountKind:            pulumi.String(getAccountKind(spec.GetAccountKind())),
			AccountTier:            pulumi.String(getAccountTier(spec.GetAccountTier())),
			AccountReplicationType: pulumi.String(getReplicationType(spec.GetReplicationType())),
			AccessTier:             pulumi.String(getAccessTier(spec.GetAccessTier())),

			// Security settings
			HttpsTrafficOnlyEnabled:    pulumi.Bool(spec.GetEnableHttpsTrafficOnly()),
			MinTlsVersion:              pulumi.String(getMinTlsVersion(spec.GetMinTlsVersion())),
			AllowNestedItemsToBePublic: pulumi.Bool(false),

			// Network rules
			NetworkRules: networkRulesArgs,

			// Blob properties
			BlobProperties: blobPropertiesArgs,

			// Tags
			Tags: pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Storage Account %s", locals.StorageAccountName)
	}

	// Create blob containers
	containerUrlMap := make(map[string]pulumi.StringOutput)

	for _, containerSpec := range spec.Containers {
		container, err := storage.NewContainer(ctx,
			fmt.Sprintf("container-%s", containerSpec.Name),
			&storage.ContainerArgs{
				Name:                pulumi.String(containerSpec.Name),
				StorageAccountName:  storageAccount.Name,
				ContainerAccessType: pulumi.String(getContainerAccessType(containerSpec.GetAccessType())),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(storageAccount))
		if err != nil {
			return errors.Wrapf(err, "failed to create container %s", containerSpec.Name)
		}

		// Build container URL
		containerUrl := pulumi.Sprintf("https://%s.blob.core.windows.net/%s", storageAccount.Name, container.Name)
		containerUrlMap[containerSpec.Name] = containerUrl
	}

	// Export stack outputs
	ctx.Export(OpStorageAccountId, storageAccount.ID())
	ctx.Export(OpStorageAccountName, storageAccount.Name)
	ctx.Export(OpPrimaryBlobEndpoint, storageAccount.PrimaryBlobEndpoint)
	ctx.Export(OpPrimaryQueueEndpoint, storageAccount.PrimaryQueueEndpoint)
	ctx.Export(OpPrimaryTableEndpoint, storageAccount.PrimaryTableEndpoint)
	ctx.Export(OpPrimaryFileEndpoint, storageAccount.PrimaryFileEndpoint)
	ctx.Export(OpPrimaryDfsEndpoint, storageAccount.PrimaryBlobHost)
	ctx.Export(OpPrimaryWebEndpoint, storageAccount.PrimaryWebEndpoint)
	ctx.Export(OpRegion, pulumi.String(spec.Region))
	ctx.Export(OpResourceGroup, pulumi.String(locals.ResourceGroupName))

	// Export container URL map
	if len(containerUrlMap) > 0 {
		urlMap := pulumi.StringMap{}
		for name, url := range containerUrlMap {
			urlMap[name] = url.ToStringOutput()
		}
		ctx.Export(OpContainerUrlMap, urlMap)
	}

	return nil
}
