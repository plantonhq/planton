package module

import (
	"github.com/pkg/errors"
	azuredatafactoryintegrationruntimev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactoryintegrationruntime/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// One kind, 3 provider resources: Azure stores every integration
// runtime flavor in the SAME factory-scoped namespace
// ({factory_id}/integrationRuntimes/{name}), so the spec's variant
// block selects which resource is created. Shared fields (name,
// factory, description) travel identically on every flavor; each
// builder adds only its variant's own arguments.
//
// The authorization-key outputs exist only on the self-hosted flavor
// (Azure issues keys for a PRIMARY self-hosted registration only);
// the managed builders export them empty so the output contract stays
// uniform across variants.
func Resources(ctx *pulumi.Context, stackInput *azuredatafactoryintegrationruntimev1alpha1.AzureDataFactoryIntegrationRuntimeStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDataFactoryIntegrationRuntime.Spec
	resourceName := locals.AzureDataFactoryIntegrationRuntime.Metadata.Name

	var outputs *runtimeOutputs

	switch {
	case spec.GetAzure() != nil:
		outputs, err = createAzure(ctx, resourceName, spec, azureProvider)
	case spec.GetAzureSsis() != nil:
		outputs, err = createAzureSsis(ctx, resourceName, spec, azureProvider)
	case spec.GetSelfHosted() != nil:
		outputs, err = createSelfHosted(ctx, resourceName, spec, azureProvider)
	default:
		// The variant oneof is required, so this is unreachable; the guard
		// keeps a broken input loud instead of silently exporting nothing.
		return errors.New("exactly one integration runtime variant block must be set")
	}
	if err != nil {
		return err
	}

	ctx.Export(OpIntegrationRuntimeId, outputs.id)
	ctx.Export(OpIntegrationRuntimeName, outputs.name)
	ctx.Export(OpPrimaryAuthorizationKey, outputs.primaryAuthorizationKey)
	ctx.Export(OpSecondaryAuthorizationKey, outputs.secondaryAuthorizationKey)

	return nil
}
