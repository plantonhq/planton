package module

import (
	"github.com/pkg/errors"
	azureeventgriddomainv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventgriddomain/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventgrid"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// identityTypeStrings maps the spec enum's values to the provider's
// identity tokens.
var identityTypeStrings = map[azureeventgriddomainv1alpha1.AzureEventgridDomainIdentityType]string{
	azureeventgriddomainv1alpha1.AzureEventgridDomainIdentityType_SYSTEM_ASSIGNED: "SystemAssigned",
	azureeventgriddomainv1alpha1.AzureEventgridDomainIdentityType_USER_ASSIGNED:   "UserAssigned",
}

func Resources(ctx *pulumi.Context, iacInput *azureeventgriddomainv1alpha1.AzureEventgridDomainIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventgridDomain.Spec

	// Create the Event Grid domain -- one publishing endpoint and one
	// pair of access keys serving many event streams (domain topics),
	// the multi-tenant pattern. The domain's name becomes a PUBLIC DNS
	// hostname ({name}.{region}.eventgrid.azure.net), unique across all
	// Azure customers in the region. Free at rest.
	domainArgs := &eventgrid.DomainArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(spec.ResourceGroup.GetValue()),
		Location:          pulumi.String(spec.Region),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// The provider defaults to EventGridSchema; the platform sends its
	// own default explicitly so both engines render the schema -- one
	// schema for every topic in the domain. Create-only: changing it
	// replaces the domain.
	inputSchema := "EventGridSchema"
	if spec.InputSchema != nil && *spec.InputSchema != "" {
		inputSchema = *spec.InputSchema
	}
	domainArgs.InputSchema = pulumi.String(inputSchema)

	// Custom-schema envelope mappings -- sent only when they carry at
	// least one field (the built-in schemas need no mapping), mirroring
	// the Terraform module.
	if fields := spec.InputMappingFields; fields != nil {
		mappingArgs := &eventgrid.DomainInputMappingFieldsArgs{}
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
			domainArgs.InputMappingFields = mappingArgs
		}
	}

	if defaults := spec.InputMappingDefaultValues; defaults != nil {
		defaultsArgs := &eventgrid.DomainInputMappingDefaultValuesArgs{}
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
			domainArgs.InputMappingDefaultValues = defaultsArgs
		}
	}

	// Domain-topic lifecycle flags -- always sent (platform defaults
	// mirror Azure's auto-managed posture). Both false is the
	// pinned-topics governance posture (declared
	// AzureEventgridDomainTopic resources only).
	autoCreate := true
	if spec.AutoCreateTopicWithFirstSubscription != nil {
		autoCreate = *spec.AutoCreateTopicWithFirstSubscription
	}
	domainArgs.AutoCreateTopicWithFirstSubscription = pulumi.Bool(autoCreate)

	autoDelete := true
	if spec.AutoDeleteTopicWithLastSubscription != nil {
		autoDelete = *spec.AutoDeleteTopicWithLastSubscription
	}
	domainArgs.AutoDeleteTopicWithLastSubscription = pulumi.Bool(autoDelete)

	// Always sent (platform defaults mirror Azure's). Local auth inverts
	// to ARM's disableLocalAuth inside the provider.
	publicNetworkAccess := true
	if spec.PublicNetworkAccessEnabled != nil {
		publicNetworkAccess = *spec.PublicNetworkAccessEnabled
	}
	domainArgs.PublicNetworkAccessEnabled = pulumi.Bool(publicNetworkAccess)

	localAuth := true
	if spec.LocalAuthEnabled != nil {
		localAuth = *spec.LocalAuthEnabled
	}
	domainArgs.LocalAuthEnabled = pulumi.Bool(localAuth)

	// The provider clears rules on update by sending an EMPTY list --
	// building the array unconditionally mirrors that (rule deletions
	// propagate). "Allow" is Azure's only legal action on this resource
	// at the pinned provider, sent explicitly.
	inboundRules := eventgrid.DomainInboundIpRuleArray{}
	for _, ipMask := range spec.InboundIpRules {
		inboundRules = append(inboundRules, &eventgrid.DomainInboundIpRuleArgs{
			IpMask: pulumi.String(ipMask),
			Action: pulumi.String("Allow"),
		})
	}
	domainArgs.InboundIpRules = inboundRules

	if spec.Identity != nil {
		identityArgs := &eventgrid.DomainIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		domainArgs.Identity = identityArgs
	}

	createdDomain, err := eventgrid.NewDomain(ctx,
		locals.AzureEventgridDomain.Metadata.Name,
		domainArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create eventgrid domain %s",
			locals.AzureEventgridDomain.Metadata.Name)
	}

	ctx.Export(OpDomainId, createdDomain.ID())
	ctx.Export(OpDomainName, createdDomain.Name)
	ctx.Export(OpEndpoint, createdDomain.Endpoint)
	ctx.Export(OpPrimaryAccessKey, pulumi.ToSecret(createdDomain.PrimaryAccessKey))
	ctx.Export(OpSecondaryAccessKey, pulumi.ToSecret(createdDomain.SecondaryAccessKey))
	// Empty unless SYSTEM_ASSIGNED is enabled -- mirrors the TF module's
	// try(identity[0].principal_id, "").
	ctx.Export(OpIdentityPrincipalId, createdDomain.Identity.PrincipalId().ApplyT(func(principalId *string) string {
		if principalId == nil {
			return ""
		}
		return *principalId
	}).(pulumi.StringOutput))

	return nil
}
