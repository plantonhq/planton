package module

import (
	"github.com/pkg/errors"
	azureeventgridnamespacev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventgridnamespace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventgrid"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// identityTypeStrings maps the spec enum's values to the provider's
// identity tokens. Like the system topic, a namespace supports the
// combined mode -- the third token carries both flavors.
var identityTypeStrings = map[azureeventgridnamespacev1alpha1.AzureEventgridNamespaceIdentityType]string{
	azureeventgridnamespacev1alpha1.AzureEventgridNamespaceIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azureeventgridnamespacev1alpha1.AzureEventgridNamespaceIdentityType_USER_ASSIGNED:            "UserAssigned",
	azureeventgridnamespacev1alpha1.AzureEventgridNamespaceIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

func Resources(ctx *pulumi.Context, iacInput *azureeventgridnamespacev1alpha1.AzureEventgridNamespaceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventgridNamespace.Spec

	// Platform default 1 TU -- always sent, so the rendered plan states
	// the provisioned throughput; updatable in place. (The proto default
	// is applied here, matching the TF module's coalesce.)
	capacity := 1
	if spec.Capacity != nil {
		capacity = int(*spec.Capacity)
	}

	// Platform default true, mapped to the provider's Enabled/Disabled
	// tokens -- always sent (mirrors Azure's own default).
	publicNetworkAccessEnabled := true
	if spec.PublicNetworkAccessEnabled != nil {
		publicNetworkAccessEnabled = *spec.PublicNetworkAccessEnabled
	}

	// "Standard" is the SKU's only legal value at v5 -- deliberately not
	// part of the spec; both engines send it explicitly.
	namespaceArgs := &eventgrid.NamespaceArgs{
		Name:                pulumi.String(spec.Name),
		ResourceGroupName:   pulumi.String(spec.ResourceGroup.GetValue()),
		Location:            pulumi.String(spec.Region),
		Sku:                 pulumi.String("Standard"),
		Capacity:            pulumi.Int(capacity),
		PublicNetworkAccess: pulumi.String(publicNetworkAccessToken(publicNetworkAccessEnabled)),
		Tags:                pulumi.ToStringMap(locals.AzureTags),
	}

	// "Allow" is Azure's only legal rule action on this resource at v5,
	// sent explicitly. The list is built unconditionally so rule
	// deletions propagate (the provider clears rules via its own expand).
	if len(spec.InboundIpRules) > 0 {
		inboundRules := eventgrid.NamespaceInboundIpRuleArray{}
		for _, ipMask := range spec.InboundIpRules {
			inboundRules = append(inboundRules, &eventgrid.NamespaceInboundIpRuleArgs{
				IpMask: pulumi.String(ipMask),
				Action: pulumi.String("Allow"),
			})
		}
		namespaceArgs.InboundIpRules = inboundRules
	}

	if spec.Identity != nil {
		identityArgs := &eventgrid.NamespaceIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		namespaceArgs.Identity = identityArgs
	}

	// The MQTT broker block. Presence is the enable switch (the provider
	// hardcodes state Enabled when the block is sent) and the WHOLE block
	// is create-only -- changing it replaces the namespace. The bridged
	// SDK pluralizes the block name (topicSpacesConfigurations) where v5
	// names it topic_spaces_configuration -- an engine-shape difference
	// only; both engines write the same ARM object.
	if spec.TopicSpacesConfiguration != nil {
		// Session dials carry platform defaults (1 session, 1 hour) --
		// always sent inside a sent block, mirroring the provider's own
		// schema defaults (proto defaults applied here, matching the TF
		// module's coalesce).
		maxClientSessions := 1
		if spec.TopicSpacesConfiguration.MaximumClientSessionsPerAuthenticationName != nil {
			maxClientSessions = int(*spec.TopicSpacesConfiguration.MaximumClientSessionsPerAuthenticationName)
		}
		maxSessionExpiryHours := 1
		if spec.TopicSpacesConfiguration.MaximumSessionExpiryInHours != nil {
			maxSessionExpiryHours = int(*spec.TopicSpacesConfiguration.MaximumSessionExpiryInHours)
		}
		topicSpaces := &eventgrid.NamespaceTopicSpacesConfigurationArgs{
			MaximumClientSessionsPerAuthenticationName: pulumi.IntPtr(maxClientSessions),
			MaximumSessionExpiryInHours:                pulumi.IntPtr(maxSessionExpiryHours),
		}
		if len(spec.TopicSpacesConfiguration.AlternativeAuthenticationNameSources) > 0 {
			topicSpaces.AlternativeAuthenticationNameSources = pulumi.ToStringArray(spec.TopicSpacesConfiguration.AlternativeAuthenticationNameSources)
		}
		if spec.TopicSpacesConfiguration.RouteTopicId.GetValue() != "" {
			topicSpaces.RouteTopicId = pulumi.String(spec.TopicSpacesConfiguration.RouteTopicId.GetValue())
		}
		if len(spec.TopicSpacesConfiguration.DynamicRoutingEnrichments) > 0 {
			enrichments := eventgrid.NamespaceTopicSpacesConfigurationDynamicRoutingEnrichmentArray{}
			for _, enrichment := range spec.TopicSpacesConfiguration.DynamicRoutingEnrichments {
				enrichments = append(enrichments, &eventgrid.NamespaceTopicSpacesConfigurationDynamicRoutingEnrichmentArgs{
					Key:   pulumi.String(enrichment.Key),
					Value: pulumi.String(enrichment.Value),
				})
			}
			topicSpaces.DynamicRoutingEnrichments = enrichments
		}
		if len(spec.TopicSpacesConfiguration.StaticRoutingEnrichments) > 0 {
			// The provider pins every static enrichment's value type to
			// String -- nothing to send from the spec.
			enrichments := eventgrid.NamespaceTopicSpacesConfigurationStaticRoutingEnrichmentArray{}
			for _, enrichment := range spec.TopicSpacesConfiguration.StaticRoutingEnrichments {
				enrichments = append(enrichments, &eventgrid.NamespaceTopicSpacesConfigurationStaticRoutingEnrichmentArgs{
					Key:   pulumi.String(enrichment.Key),
					Value: pulumi.String(enrichment.Value),
				})
			}
			topicSpaces.StaticRoutingEnrichments = enrichments
		}
		namespaceArgs.TopicSpacesConfigurations = eventgrid.NamespaceTopicSpacesConfigurationArray{topicSpaces}
	}

	createdNamespace, err := eventgrid.NewNamespace(ctx,
		locals.AzureEventgridNamespace.Metadata.Name,
		namespaceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create eventgrid namespace %s",
			locals.AzureEventgridNamespace.Metadata.Name)
	}

	ctx.Export(OpNamespaceId, createdNamespace.ID())
	ctx.Export(OpNamespaceName, createdNamespace.Name)
	// Empty unless a system-assigned flavor is enabled -- mirrors the TF
	// module's try(identity[0].principal_id, "").
	ctx.Export(OpIdentityPrincipalId, createdNamespace.Identity.PrincipalId().ApplyT(func(principalId *string) string {
		if principalId == nil {
			return ""
		}
		return *principalId
	}).(pulumi.StringOutput))

	return nil
}

// publicNetworkAccessToken maps the spec's bool onto the provider's
// Enabled/Disabled vocabulary. (Azure also defines SecuredByPerimeter,
// which the provider does not admit at v5.)
func publicNetworkAccessToken(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}
