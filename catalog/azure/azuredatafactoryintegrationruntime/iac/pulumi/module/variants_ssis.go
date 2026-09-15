package module

import (
	"github.com/pkg/errors"
	azuredatafactoryintegrationruntimev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactoryintegrationruntime/v1alpha1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/datafactory"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The managed SSIS package runtime builder
// (azurerm_data_factory_integration_runtime_azure_ssis).
func createAzureSsis(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactoryintegrationruntimev1alpha1.AzureDataFactoryIntegrationRuntimeSpec,
	azureProvider pulumi.ProviderResource,
) (*runtimeOutputs, error) {
	ssis := spec.GetAzureSsis()

	args := &datafactory.IntegrationRuntimeSsisArgs{
		Name:                         pulumi.String(spec.Name),
		DataFactoryId:                pulumi.String(spec.DataFactoryId.GetValue()),
		Location:                     pulumi.String(ssis.Region),
		NodeSize:                     pulumi.String(ssis.NodeSize),
		Description:                  stringPtrWhenSet(spec.Description),
		NumberOfNodes:                intPtrWhenSet(ssis.NumberOfNodes),
		MaxParallelExecutionsPerNode: intPtrWhenSet(ssis.MaxParallelExecutionsPerNode),
		Edition:                      stringPtrWhenSet(ssis.Edition),
		LicenseType:                  stringPtrWhenSet(ssis.LicenseType),
		CredentialName:               stringPtrWhenSet(ssis.CredentialName),
	}

	if catalog := ssis.CatalogInfo; catalog != nil {
		args.CatalogInfo = datafactory.IntegrationRuntimeSsisCatalogInfoArgs{
			ServerEndpoint:        pulumi.String(catalog.ServerEndpoint),
			AdministratorLogin:    stringPtrWhenSet(catalog.AdministratorLogin),
			AdministratorPassword: stringPtrWhenSet(catalog.AdministratorPassword),
			PricingTier:           stringPtrWhenSet(catalog.PricingTier),
			ElasticPoolName:       stringPtrWhenSet(catalog.ElasticPoolName),
			DualStandbyPairName:   stringPtrWhenSet(catalog.DualStandbyPairName),
		}
	}

	if script := ssis.CustomSetupScript; script != nil {
		args.CustomSetupScript = datafactory.IntegrationRuntimeSsisCustomSetupScriptArgs{
			BlobContainerUri: pulumi.String(script.BlobContainerUri),
			SasToken:         pulumi.String(script.SasToken),
		}
	}

	if setup := ssis.ExpressCustomSetup; setup != nil {
		setupArgs := datafactory.IntegrationRuntimeSsisExpressCustomSetupArgs{
			PowershellVersion: stringPtrWhenSet(setup.PowershellVersion),
		}
		if len(setup.Environment) > 0 {
			setupArgs.Environment = pulumi.ToStringMap(setup.Environment)
		}
		if len(setup.CommandKey) > 0 {
			commandKeys := make(datafactory.IntegrationRuntimeSsisExpressCustomSetupCommandKeyArray, 0, len(setup.CommandKey))
			for _, commandKey := range setup.CommandKey {
				commandKeyArgs := datafactory.IntegrationRuntimeSsisExpressCustomSetupCommandKeyArgs{
					TargetName: pulumi.String(commandKey.TargetName),
					UserName:   pulumi.String(commandKey.UserName),
					// When both password forms are set, Azure receives the
					// inline password (the provider's own precedence).
					Password: stringPtrWhenSet(commandKey.Password),
				}
				if commandKey.KeyVaultPassword != nil {
					commandKeyArgs.KeyVaultPassword = datafactory.IntegrationRuntimeSsisExpressCustomSetupCommandKeyKeyVaultPasswordArgs{
						LinkedServiceName: pulumi.String(commandKey.KeyVaultPassword.LinkedServiceName.GetValue()),
						SecretName:        pulumi.String(commandKey.KeyVaultPassword.SecretName),
						SecretVersion:     stringPtrWhenSet(commandKey.KeyVaultPassword.SecretVersion),
						Parameters:        parametersMapWhenSet(commandKey.KeyVaultPassword.Parameters),
					}
				}
				commandKeys = append(commandKeys, commandKeyArgs)
			}
			setupArgs.CommandKeys = commandKeys
		}
		if len(setup.Component) > 0 {
			components := make(datafactory.IntegrationRuntimeSsisExpressCustomSetupComponentArray, 0, len(setup.Component))
			for _, component := range setup.Component {
				componentArgs := datafactory.IntegrationRuntimeSsisExpressCustomSetupComponentArgs{
					Name: pulumi.String(component.Name),
					// When both license forms are set, Azure receives the
					// inline license (the provider's own precedence).
					License: stringPtrWhenSet(component.License),
				}
				if component.KeyVaultLicense != nil {
					componentArgs.KeyVaultLicense = datafactory.IntegrationRuntimeSsisExpressCustomSetupComponentKeyVaultLicenseArgs{
						LinkedServiceName: pulumi.String(component.KeyVaultLicense.LinkedServiceName.GetValue()),
						SecretName:        pulumi.String(component.KeyVaultLicense.SecretName),
						SecretVersion:     stringPtrWhenSet(component.KeyVaultLicense.SecretVersion),
						Parameters:        parametersMapWhenSet(component.KeyVaultLicense.Parameters),
					}
				}
				components = append(components, componentArgs)
			}
			setupArgs.Components = components
		}
		args.ExpressCustomSetup = setupArgs
	}

	if express := ssis.ExpressVnetIntegration; express != nil {
		args.ExpressVnetIntegration = datafactory.IntegrationRuntimeSsisExpressVnetIntegrationArgs{
			SubnetId: pulumi.String(express.SubnetId.GetValue()),
		}
	}

	if vnet := ssis.VnetIntegration; vnet != nil {
		vnetArgs := datafactory.IntegrationRuntimeSsisVnetIntegrationArgs{
			SubnetName: stringPtrWhenSet(vnet.SubnetName),
		}
		if vnet.VnetId != nil {
			vnetArgs.VnetId = pulumi.StringPtr(vnet.VnetId.GetValue())
		}
		if vnet.SubnetId != nil {
			vnetArgs.SubnetId = pulumi.StringPtr(vnet.SubnetId.GetValue())
		}
		if len(vnet.PublicIps) > 0 {
			publicIps := make(pulumi.StringArray, 0, len(vnet.PublicIps))
			for _, publicIp := range vnet.PublicIps {
				publicIps = append(publicIps, pulumi.String(publicIp.GetValue()))
			}
			vnetArgs.PublicIps = publicIps
		}
		args.VnetIntegration = vnetArgs
	}

	if len(ssis.PackageStore) > 0 {
		stores := make(datafactory.IntegrationRuntimeSsisPackageStoreArray, 0, len(ssis.PackageStore))
		for _, store := range ssis.PackageStore {
			stores = append(stores, datafactory.IntegrationRuntimeSsisPackageStoreArgs{
				Name:              pulumi.String(store.Name),
				LinkedServiceName: pulumi.String(store.LinkedServiceName.GetValue()),
			})
		}
		args.PackageStores = stores
	}

	if scale := ssis.CopyComputeScale; scale != nil {
		args.CopyComputeScale = datafactory.IntegrationRuntimeSsisCopyComputeScaleArgs{
			DataIntegrationUnit: intPtrWhenSet(scale.DataIntegrationUnit),
			TimeToLive:          intPtrWhenSet(scale.TimeToLive),
		}
	}

	if scale := ssis.PipelineExternalComputeScale; scale != nil {
		// number_of_external_nodes is sent correctly but Azure's read
		// API mirrors number_of_pipeline_nodes back for it -- a provider
		// read seam, documented on the spec field.
		args.PipelineExternalComputeScale = datafactory.IntegrationRuntimeSsisPipelineExternalComputeScaleArgs{
			NumberOfExternalNodes: intPtrWhenSet(scale.NumberOfExternalNodes),
			NumberOfPipelineNodes: intPtrWhenSet(scale.NumberOfPipelineNodes),
			TimeToLive:            intPtrWhenSet(scale.TimeToLive),
		}
	}

	if proxy := ssis.Proxy; proxy != nil {
		args.Proxy = datafactory.IntegrationRuntimeSsisProxyArgs{
			SelfHostedIntegrationRuntimeName: pulumi.String(proxy.SelfHostedIntegrationRuntimeName.GetValue()),
			StagingStorageLinkedServiceName:  pulumi.String(proxy.StagingStorageLinkedServiceName.GetValue()),
			Path:                             stringPtrWhenSet(proxy.Path),
		}
	}

	created, err := datafactory.NewIntegrationRuntimeSsis(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create ssis integration runtime %s", resourceName)
	}

	return &runtimeOutputs{
		id:                        created.ID(),
		name:                      created.Name,
		primaryAuthorizationKey:   pulumi.String(""),
		secondaryAuthorizationKey: pulumi.String(""),
	}, nil
}

// parametersMapWhenSet leaves a Key Vault reference's parameter map
// unsent when empty.
func parametersMapWhenSet(parameters map[string]string) pulumi.StringMapInput {
	if len(parameters) == 0 {
		return nil
	}
	return pulumi.ToStringMap(parameters)
}
