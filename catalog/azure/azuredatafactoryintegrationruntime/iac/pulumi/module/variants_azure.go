package module

import (
	"github.com/pkg/errors"
	azuredatafactoryintegrationruntimev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactoryintegrationruntime/v1alpha1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/datafactory"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The managed data-flow compute builder
// (azurerm_data_factory_integration_runtime_azure).
func createAzure(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactoryintegrationruntimev1alpha1.AzureDataFactoryIntegrationRuntimeSpec,
	azureProvider pulumi.ProviderResource,
) (*runtimeOutputs, error) {
	azure := spec.GetAzure()

	// cleanup_enabled platform-defaults to TRUE (tear the cluster down
	// after every run) -- sent explicitly so the manifest's intent is
	// always on the wire, mirroring the provider's own default.
	cleanupEnabled := true
	if azure.CleanupEnabled != nil {
		cleanupEnabled = *azure.CleanupEnabled
	}

	args := &datafactory.IntegrationRuntimeRuleArgs{
		Name:           pulumi.String(spec.Name),
		DataFactoryId:  pulumi.String(spec.DataFactoryId.GetValue()),
		Location:       pulumi.String(azure.Region),
		Description:    stringPtrWhenSet(spec.Description),
		CleanupEnabled: pulumi.Bool(cleanupEnabled),
		ComputeType:    stringPtrWhenSet(azure.ComputeType),
		CoreCount:      intPtrWhenSet(azure.CoreCount),
		TimeToLiveMin:  intPtrWhenSet(azure.TimeToLiveMin),
		// Applied by the provider through a separate
		// enable-interactive-authoring operation once the runtime is
		// online (not part of the create body).
		InteractiveAuthoringTimeToLiveInMinutes: intPtrWhenSet(azure.InteractiveAuthoringTimeToLiveInMinutes),
		VirtualNetworkEnabled:                   boolPtrWhenTrue(azure.VirtualNetworkEnabled),
	}

	created, err := datafactory.NewIntegrationRuntimeRule(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create azure integration runtime %s", resourceName)
	}

	return &runtimeOutputs{
		id:                        created.ID(),
		name:                      created.Name,
		primaryAuthorizationKey:   pulumi.String(""),
		secondaryAuthorizationKey: pulumi.String(""),
	}, nil
}
