package module

import (
	"github.com/pkg/errors"
	azuredatafactoryintegrationruntimev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactoryintegrationruntime/v1alpha1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/datafactory"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The self-hosted agent registration builder
// (azurerm_data_factory_integration_runtime_self_hosted).
func createSelfHosted(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactoryintegrationruntimev1alpha1.AzureDataFactoryIntegrationRuntimeSpec,
	azureProvider pulumi.ProviderResource,
) (*runtimeOutputs, error) {
	selfHosted := spec.GetSelfHosted()

	args := &datafactory.IntegrationRuntimeSelfHostedArgs{
		Name:                                     pulumi.String(spec.Name),
		DataFactoryId:                            pulumi.String(spec.DataFactoryId.GetValue()),
		Description:                              stringPtrWhenSet(spec.Description),
		SelfContainedInteractiveAuthoringEnabled: boolPtrWhenTrue(selfHosted.SelfContainedInteractiveAuthoringEnabled),
	}

	if selfHosted.RbacAuthorization != nil {
		args.RbacAuthorizations = datafactory.IntegrationRuntimeSelfHostedRbacAuthorizationArray{
			datafactory.IntegrationRuntimeSelfHostedRbacAuthorizationArgs{
				ResourceId: pulumi.String(selfHosted.RbacAuthorization.ResourceId.GetValue()),
			},
		}
	}

	created, err := datafactory.NewIntegrationRuntimeSelfHosted(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create self-hosted integration runtime %s", resourceName)
	}

	// Azure returns the authorization keys readable and the provider
	// does not flag them -- wrapped as secrets here so the catalog's
	// sensitive-output contract holds. Empty for a LINKED registration
	// (Azure issues keys for primary registrations only).
	return &runtimeOutputs{
		id:                        created.ID(),
		name:                      created.Name,
		primaryAuthorizationKey:   pulumi.ToSecret(created.PrimaryAuthorizationKey).(pulumi.StringOutput),
		secondaryAuthorizationKey: pulumi.ToSecret(created.SecondaryAuthorizationKey).(pulumi.StringOutput),
	}, nil
}
