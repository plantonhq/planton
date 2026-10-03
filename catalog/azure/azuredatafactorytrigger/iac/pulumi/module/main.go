package module

import (
	"github.com/pkg/errors"
	azuredatafactorytriggerv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactorytrigger/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// One kind, four provider resources: Azure stores all four trigger
// types in the SAME factory-scoped trigger namespace, so the spec's
// variant block selects which resource is created. All four share the
// started/stopped lifecycle: the provider stops a started trigger
// before any update or delete, then starts it again when `activated`
// is true (the platform default, sent explicitly).
func Resources(ctx *pulumi.Context, iacInput *azuredatafactorytriggerv1alpha1.AzureDataFactoryTriggerIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDataFactoryTrigger.Spec
	resourceName := locals.AzureDataFactoryTrigger.Metadata.Name

	var triggerId pulumi.StringInput
	var triggerName pulumi.StringInput

	switch {
	case spec.Schedule != nil:
		triggerId, triggerName, err = createScheduleTrigger(ctx, resourceName, spec, azureProvider)
	case spec.TumblingWindow != nil:
		triggerId, triggerName, err = createTumblingWindowTrigger(ctx, resourceName, spec, azureProvider)
	case spec.BlobEvent != nil:
		triggerId, triggerName, err = createBlobEventTrigger(ctx, resourceName, spec, azureProvider)
	case spec.CustomEvent != nil:
		triggerId, triggerName, err = createCustomEventTrigger(ctx, resourceName, spec, azureProvider)
	default:
		// The spec's exactly-one CEL makes this unreachable; the guard
		// keeps a broken input loud instead of silently exporting nothing.
		return errors.New("exactly one trigger variant (schedule, tumbling_window, blob_event, custom_event) must be set")
	}
	if err != nil {
		return err
	}

	ctx.Export(OpTriggerId, triggerId)
	ctx.Export(OpTriggerName, triggerName)

	return nil
}

// activatedOrDefault applies the platform default: a trigger deploys
// STARTED unless the spec says otherwise.
func activatedOrDefault(spec *azuredatafactorytriggerv1alpha1.AzureDataFactoryTriggerSpec) bool {
	if spec.Activated != nil {
		return spec.GetActivated()
	}
	return true
}
