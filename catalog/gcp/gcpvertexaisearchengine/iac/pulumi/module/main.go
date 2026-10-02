package module

import (
	"github.com/pkg/errors"
	gcpvertexaisearchenginev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaisearchengine/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpvertexaisearchenginev1alpha1.GcpVertexAiSearchEngineStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	created, err := engine(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create vertex ai search engine")
	}

	createdControls, controlNames, err := controls(ctx, locals, gcpProvider, created)
	if err != nil {
		return errors.Wrap(err, "failed to create engine controls")
	}

	servingConfigName, err := servingConfig(ctx, locals, gcpProvider, created, createdControls)
	if err != nil {
		return errors.Wrap(err, "failed to configure the engine's serving config")
	}

	widgetConfigName, err := widgetConfig(ctx, locals, gcpProvider, created)
	if err != nil {
		return errors.Wrap(err, "failed to configure the engine's widget config")
	}

	assistantNames, err := assistants(ctx, locals, gcpProvider, created)
	if err != nil {
		return errors.Wrap(err, "failed to create engine assistants")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpEngineId, created.EngineId)
	ctx.Export(OpLocation, pulumi.String(locals.GcpVertexAiSearchEngine.Spec.Location))
	ctx.Export(OpCollectionId, pulumi.String(locals.CollectionId))
	ctx.Export(OpEngineType, pulumi.String(locals.EngineType))
	ctx.Export(OpServingConfigName, servingConfigName)
	ctx.Export(OpWidgetConfigName, widgetConfigName)
	ctx.Export(OpDialogflowAgent, created.DialogflowAgent)
	ctx.Export(OpControlNames, controlNames.ToStringArrayOutput())
	ctx.Export(OpAssistantNames, assistantNames.ToStringArrayOutput())
	return nil
}
