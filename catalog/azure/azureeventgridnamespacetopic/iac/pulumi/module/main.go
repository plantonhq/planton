package module

import (
	"github.com/pkg/errors"
	azureeventgridnamespacetopicv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventgridnamespacetopic/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventgrid"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureeventgridnamespacetopicv1alpha1.AzureEventgridNamespaceTopicIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventgridNamespaceTopic.Spec

	// Retention carries the platform default (7 days, the provider's
	// own) and is the topic's only updatable property. The proto default
	// is applied here, matching the TF module's coalesce.
	eventRetentionInDays := 7
	if spec.EventRetentionInDays != nil {
		eventRetentionInDays = int(*spec.EventRetentionInDays)
	}

	// Create the namespace topic. Azure pins the event schema to
	// CloudEvents v1.0 and the publisher type to Custom (the provider
	// sends both; neither is configurable).
	namespaceTopicArgs := &eventgrid.NamespaceTopicArgs{
		Name:                 pulumi.String(spec.Name),
		EventgridNamespaceId: pulumi.String(spec.NamespaceId.GetValue()),
		EventRetentionInDays: pulumi.IntPtr(eventRetentionInDays),
	}

	createdNamespaceTopic, err := eventgrid.NewNamespaceTopic(ctx,
		locals.AzureEventgridNamespaceTopic.Metadata.Name,
		namespaceTopicArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create eventgrid namespace topic %s",
			locals.AzureEventgridNamespaceTopic.Metadata.Name)
	}

	ctx.Export(OpNamespaceTopicId, createdNamespaceTopic.ID())
	ctx.Export(OpNamespaceTopicName, createdNamespaceTopic.Name)

	return nil
}
