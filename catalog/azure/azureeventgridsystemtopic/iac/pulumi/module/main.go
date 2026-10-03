package module

import (
	"github.com/pkg/errors"
	azureeventgridsystemtopicv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventgridsystemtopic/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventgrid"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// identityTypeStrings maps the spec enum's values to the provider's
// identity tokens. Unlike the Event Grid publisher kinds, a system
// topic supports the combined mode -- the provider's third token
// carries both flavors.
var identityTypeStrings = map[azureeventgridsystemtopicv1alpha1.AzureEventgridSystemTopicIdentityType]string{
	azureeventgridsystemtopicv1alpha1.AzureEventgridSystemTopicIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azureeventgridsystemtopicv1alpha1.AzureEventgridSystemTopicIdentityType_USER_ASSIGNED:            "UserAssigned",
	azureeventgridsystemtopicv1alpha1.AzureEventgridSystemTopicIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

func Resources(ctx *pulumi.Context, iacInput *azureeventgridsystemtopicv1alpha1.AzureEventgridSystemTopicIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventgridSystemTopic.Spec

	// Create the Event Grid system topic. Azure allows ONE per source
	// resource per topic type; the region must match the source's (or
	// be "Global" for global sources). Free at rest. The bridged SDK
	// still carries the deprecated sourceArmResourceId alias -- the v5
	// name below is the one the pinned Terraform provider keeps.
	systemTopicArgs := &eventgrid.SystemTopicArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(spec.ResourceGroup.GetValue()),
		Location:          pulumi.String(spec.Region),
		SourceResourceId:  pulumi.String(spec.SourceResourceId.GetValue()),
		TopicType:         pulumi.String(spec.TopicType),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.Identity != nil {
		identityArgs := &eventgrid.SystemTopicIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		systemTopicArgs.Identity = identityArgs
	}

	createdSystemTopic, err := eventgrid.NewSystemTopic(ctx,
		locals.AzureEventgridSystemTopic.Metadata.Name,
		systemTopicArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create eventgrid system topic %s",
			locals.AzureEventgridSystemTopic.Metadata.Name)
	}

	ctx.Export(OpSystemTopicId, createdSystemTopic.ID())
	ctx.Export(OpSystemTopicName, createdSystemTopic.Name)
	ctx.Export(OpMetricResourceId, createdSystemTopic.MetricResourceId)
	// Empty unless a system-assigned flavor is enabled -- mirrors the TF
	// module's try(identity[0].principal_id, "").
	ctx.Export(OpIdentityPrincipalId, createdSystemTopic.Identity.PrincipalId().ApplyT(func(principalId *string) string {
		if principalId == nil {
			return ""
		}
		return *principalId
	}).(pulumi.StringOutput))

	return nil
}
