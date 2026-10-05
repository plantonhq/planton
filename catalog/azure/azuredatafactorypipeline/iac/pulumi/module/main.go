package module

import (
	"github.com/pkg/errors"
	azuredatafactorypipelinev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactorypipeline/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/datafactory"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuredatafactorypipelinev1alpha1.AzureDataFactoryPipelineIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDataFactoryPipeline.Spec

	// Create the pipeline. The activities travel as the raw JSON
	// "activities" array (Azure owns that schema -- the catalog
	// deliberately does not re-model it); the provider normalizes JSON
	// key ordering when diffing. Parameters and variables are
	// String-typed on this surface (the provider round-trips only
	// string-typed entries).
	pipelineArgs := &datafactory.PipelineArgs{
		Name:          pulumi.String(spec.Name),
		DataFactoryId: pulumi.String(spec.DataFactoryId.GetValue()),
	}

	if spec.Description != "" {
		pipelineArgs.Description = pulumi.String(spec.Description)
	}
	if len(spec.Parameters) > 0 {
		pipelineArgs.Parameters = pulumi.ToStringMap(spec.Parameters)
	}
	if len(spec.Variables) > 0 {
		pipelineArgs.Variables = pulumi.ToStringMap(spec.Variables)
	}
	if spec.ActivitiesJson != "" {
		pipelineArgs.ActivitiesJson = pulumi.String(spec.ActivitiesJson)
	}
	if len(spec.Annotations) > 0 {
		pipelineArgs.Annotations = pulumi.ToStringArray(spec.Annotations)
	}
	if spec.Concurrency != nil {
		pipelineArgs.Concurrency = pulumi.IntPtr(int(spec.GetConcurrency()))
	}
	if spec.Folder != "" {
		pipelineArgs.Folder = pulumi.String(spec.Folder)
	}
	// ENGINE-SHAPE: the bridged SDK preserves the provider's historic
	// field-name typo ("Moniter...") for v5's
	// monitor_metrics_after_duration -- a name difference only; both
	// engines write the same ARM elapsed-time-metric policy.
	if spec.MonitorMetricsAfterDuration != "" {
		pipelineArgs.MoniterMetricsAfterDuration = pulumi.String(spec.MonitorMetricsAfterDuration)
	}

	createdPipeline, err := datafactory.NewPipeline(ctx,
		locals.AzureDataFactoryPipeline.Metadata.Name,
		pipelineArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create data factory pipeline %s",
			locals.AzureDataFactoryPipeline.Metadata.Name)
	}

	ctx.Export(OpPipelineId, createdPipeline.ID())
	ctx.Export(OpPipelineName, createdPipeline.Name)

	return nil
}
