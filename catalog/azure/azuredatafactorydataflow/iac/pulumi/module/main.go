package module

import (
	"github.com/pkg/errors"
	azuredatafactorydataflowv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactorydataflow/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/datafactory"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// One kind, two provider resources: Azure stores mapping data flows
// and flowlets in the SAME factory-scoped dataflow namespace,
// differing only in the ARM type token -- so the spec's flowlet flag
// selects which resource is created, and flipping it replaces the
// object. The SDK generates a parallel type set per resource, hence
// the twin builders in builders.go.
func Resources(ctx *pulumi.Context, iacInput *azuredatafactorydataflowv1alpha1.AzureDataFactoryDataFlowIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDataFactoryDataFlow.Spec
	resourceName := locals.AzureDataFactoryDataFlow.Metadata.Name

	var dataFlowId pulumi.StringInput
	var dataFlowName pulumi.StringInput

	if spec.Flowlet {
		flowletArgs := &datafactory.FlowletDataFlowArgs{
			Name:            pulumi.String(spec.Name),
			DataFactoryId:   pulumi.String(spec.DataFactoryId.GetValue()),
			Sources:         buildFlowletSources(spec.Sources),
			Sinks:           buildFlowletSinks(spec.Sinks),
			Transformations: buildFlowletTransformations(spec.Transformations),
		}
		applyFlowletSharedFields(flowletArgs, spec)

		createdFlowlet, err := datafactory.NewFlowletDataFlow(ctx,
			resourceName,
			flowletArgs,
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create data factory flowlet data flow %s", resourceName)
		}

		dataFlowId = createdFlowlet.ID()
		dataFlowName = createdFlowlet.Name
	} else {
		dataFlowArgs := &datafactory.DataFlowArgs{
			Name:            pulumi.String(spec.Name),
			DataFactoryId:   pulumi.String(spec.DataFactoryId.GetValue()),
			Sources:         buildDataFlowSources(spec.Sources),
			Sinks:           buildDataFlowSinks(spec.Sinks),
			Transformations: buildDataFlowTransformations(spec.Transformations),
		}
		applyDataFlowSharedFields(dataFlowArgs, spec)

		createdDataFlow, err := datafactory.NewDataFlow(ctx,
			resourceName,
			dataFlowArgs,
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create data factory data flow %s", resourceName)
		}

		dataFlowId = createdDataFlow.ID()
		dataFlowName = createdDataFlow.Name
	}

	ctx.Export(OpDataFlowId, dataFlowId)
	ctx.Export(OpDataFlowName, dataFlowName)

	return nil
}

// applyDataFlowSharedFields sets the fields both forms share on the
// mapping resource's args.
func applyDataFlowSharedFields(args *datafactory.DataFlowArgs, spec *azuredatafactorydataflowv1alpha1.AzureDataFactoryDataFlowSpec) {
	if spec.Script != "" {
		args.Script = pulumi.String(spec.Script)
	}
	if len(spec.ScriptLines) > 0 {
		args.ScriptLines = pulumi.ToStringArray(spec.ScriptLines)
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if len(spec.Annotations) > 0 {
		args.Annotations = pulumi.ToStringArray(spec.Annotations)
	}
	if spec.Folder != "" {
		args.Folder = pulumi.String(spec.Folder)
	}
}

// applyFlowletSharedFields sets the fields both forms share on the
// flowlet resource's args.
func applyFlowletSharedFields(args *datafactory.FlowletDataFlowArgs, spec *azuredatafactorydataflowv1alpha1.AzureDataFactoryDataFlowSpec) {
	if spec.Script != "" {
		args.Script = pulumi.String(spec.Script)
	}
	if len(spec.ScriptLines) > 0 {
		args.ScriptLines = pulumi.ToStringArray(spec.ScriptLines)
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if len(spec.Annotations) > 0 {
		args.Annotations = pulumi.ToStringArray(spec.Annotations)
	}
	if spec.Folder != "" {
		args.Folder = pulumi.String(spec.Folder)
	}
}
