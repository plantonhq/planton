package module

import (
	"github.com/pkg/errors"
	azureeventgridtopicv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventgridtopic/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventgrid"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// identityTypeStrings maps the spec enum's values to the provider's
// identity tokens.
var identityTypeStrings = map[azureeventgridtopicv1alpha1.AzureEventgridTopicIdentityType]string{
	azureeventgridtopicv1alpha1.AzureEventgridTopicIdentityType_SYSTEM_ASSIGNED: "SystemAssigned",
	azureeventgridtopicv1alpha1.AzureEventgridTopicIdentityType_USER_ASSIGNED:   "UserAssigned",
}

func Resources(ctx *pulumi.Context, stackInput *azureeventgridtopicv1alpha1.AzureEventgridTopicStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventgridTopic.Spec

	// Create the Event Grid custom topic. The topic's name becomes a
	// PUBLIC DNS hostname ({name}.{region}.eventgrid.azure.net), unique
	// across all Azure customers in the region. Free at rest.
	topicArgs := &eventgrid.TopicArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(spec.ResourceGroup.GetValue()),
		Location:          pulumi.String(spec.Region),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// The provider defaults to EventGridSchema; the platform sends its
	// own default explicitly so both engines render the schema.
	// Create-only: changing it replaces the topic.
	inputSchema := "EventGridSchema"
	if spec.InputSchema != nil && *spec.InputSchema != "" {
		inputSchema = *spec.InputSchema
	}
	topicArgs.InputSchema = pulumi.String(inputSchema)

	// Custom-schema envelope mappings -- sent only when they carry at
	// least one field (the built-in schemas need no mapping), mirroring
	// the Terraform module.
	if fields := spec.InputMappingFields; fields != nil {
		mappingArgs := &eventgrid.TopicInputMappingFieldsArgs{}
		set := false
		if fields.Id != "" {
			mappingArgs.Id = pulumi.String(fields.Id)
			set = true
		}
		if fields.Topic != "" {
			mappingArgs.Topic = pulumi.String(fields.Topic)
			set = true
		}
		if fields.EventTime != "" {
			mappingArgs.EventTime = pulumi.String(fields.EventTime)
			set = true
		}
		if fields.EventType != "" {
			mappingArgs.EventType = pulumi.String(fields.EventType)
			set = true
		}
		if fields.Subject != "" {
			mappingArgs.Subject = pulumi.String(fields.Subject)
			set = true
		}
		if fields.DataVersion != "" {
			mappingArgs.DataVersion = pulumi.String(fields.DataVersion)
			set = true
		}
		if set {
			topicArgs.InputMappingFields = mappingArgs
		}
	}

	if defaults := spec.InputMappingDefaultValues; defaults != nil {
		defaultsArgs := &eventgrid.TopicInputMappingDefaultValuesArgs{}
		set := false
		if defaults.EventType != "" {
			defaultsArgs.EventType = pulumi.String(defaults.EventType)
			set = true
		}
		if defaults.Subject != "" {
			defaultsArgs.Subject = pulumi.String(defaults.Subject)
			set = true
		}
		if defaults.DataVersion != "" {
			defaultsArgs.DataVersion = pulumi.String(defaults.DataVersion)
			set = true
		}
		if set {
			topicArgs.InputMappingDefaultValues = defaultsArgs
		}
	}

	// Always sent (platform defaults mirror Azure's). Local auth inverts
	// to ARM's disableLocalAuth inside the provider.
	publicNetworkAccess := true
	if spec.PublicNetworkAccessEnabled != nil {
		publicNetworkAccess = *spec.PublicNetworkAccessEnabled
	}
	topicArgs.PublicNetworkAccessEnabled = pulumi.Bool(publicNetworkAccess)

	localAuth := true
	if spec.LocalAuthEnabled != nil {
		localAuth = *spec.LocalAuthEnabled
	}
	topicArgs.LocalAuthEnabled = pulumi.Bool(localAuth)

	// The provider clears rules on update by sending an EMPTY list --
	// building the array unconditionally mirrors that (rule deletions
	// propagate). "Allow" is Azure's only legal action on this resource
	// at the pinned provider, sent explicitly.
	inboundRules := eventgrid.TopicInboundIpRuleArray{}
	for _, ipMask := range spec.InboundIpRules {
		inboundRules = append(inboundRules, &eventgrid.TopicInboundIpRuleArgs{
			IpMask: pulumi.String(ipMask),
			Action: pulumi.String("Allow"),
		})
	}
	topicArgs.InboundIpRules = inboundRules

	if spec.Identity != nil {
		identityArgs := &eventgrid.TopicIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		topicArgs.Identity = identityArgs
	}

	createdTopic, err := eventgrid.NewTopic(ctx,
		locals.AzureEventgridTopic.Metadata.Name,
		topicArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create eventgrid topic %s",
			locals.AzureEventgridTopic.Metadata.Name)
	}

	ctx.Export(OpTopicId, createdTopic.ID())
	ctx.Export(OpTopicName, createdTopic.Name)
	ctx.Export(OpEndpoint, createdTopic.Endpoint)
	ctx.Export(OpPrimaryAccessKey, pulumi.ToSecret(createdTopic.PrimaryAccessKey))
	ctx.Export(OpSecondaryAccessKey, pulumi.ToSecret(createdTopic.SecondaryAccessKey))
	// Empty unless SYSTEM_ASSIGNED is enabled -- mirrors the TF module's
	// try(identity[0].principal_id, "").
	ctx.Export(OpIdentityPrincipalId, createdTopic.Identity.PrincipalId().ApplyT(func(principalId *string) string {
		if principalId == nil {
			return ""
		}
		return *principalId
	}).(pulumi.StringOutput))

	return nil
}
