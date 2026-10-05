package module

import (
	"github.com/pkg/errors"
	azureeventgriddomaintopicv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventgriddomaintopic/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventgrid"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureeventgriddomaintopicv1alpha1.AzureEventgridDomainTopicIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventgridDomainTopic.Spec

	// Create one named event stream (domain topic) inside the domain --
	// the per-tenant mailbox of the multi-tenant pattern. Free at rest;
	// everything is create-only (the topic is pure addressing).
	createdDomainTopic, err := eventgrid.NewDomainTopic(ctx,
		locals.AzureEventgridDomainTopic.Metadata.Name,
		&eventgrid.DomainTopicArgs{
			Name:              pulumi.String(spec.Name),
			DomainName:        pulumi.String(locals.DomainName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create eventgrid domain topic %s",
			locals.AzureEventgridDomainTopic.Metadata.Name)
	}

	ctx.Export(OpDomainTopicId, createdDomainTopic.ID())
	ctx.Export(OpDomainTopicName, createdDomainTopic.Name)

	return nil
}
