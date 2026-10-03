package module

import (
	"github.com/pkg/errors"
	azureeventgrideventsubscriptionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventgrideventsubscription/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventgrid"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// identityTypeStrings maps the spec enum's values to the provider's
// identity tokens (delivery and dead-letter identities allow exactly
// these two -- there is no combined mode on a subscription).
var identityTypeStrings = map[azureeventgrideventsubscriptionv1alpha1.AzureEventgridEventSubscriptionIdentityType]string{
	azureeventgrideventsubscriptionv1alpha1.AzureEventgridEventSubscriptionIdentityType_SYSTEM_ASSIGNED: "SystemAssigned",
	azureeventgrideventsubscriptionv1alpha1.AzureEventgridEventSubscriptionIdentityType_USER_ASSIGNED:   "UserAssigned",
}

// The addressing choice (spec-enforced to exactly one) selects which
// of the two provider resources materializes. Azure's two subscription
// resources share ONE configuration grammar (the provider generates
// both from the same schema), but the bridged SDK types them
// separately, so buildScopedArgs and buildSystemTopicArgs below map
// the same spec surface twice -- keep them in lockstep when either
// changes.
//
// ENGINE-SHAPE NOTE: the bridged SDK (azurerm-4.x lineage) still names
// the id-arm destinations eventhubEndpointId / hybridConnectionEndpointId
// / serviceBusQueueEndpointId / serviceBusTopicEndpointId, where the
// pinned Terraform provider renamed them to eventhub_id /
// hybrid_connection_id / service_bus_queue_id / service_bus_topic_id
// at v5. Both engines write the identical ARM destination object.
func Resources(ctx *pulumi.Context, iacInput *azureeventgrideventsubscriptionv1alpha1.AzureEventgridEventSubscriptionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventgridEventSubscription.Spec
	resourceName := locals.AzureEventgridEventSubscription.Metadata.Name

	if spec.Scope != nil {
		createdSubscription, err := eventgrid.NewEventSubscription(ctx,
			resourceName,
			buildScopedArgs(spec),
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create eventgrid event subscription %s", resourceName)
		}
		ctx.Export(OpEventSubscriptionId, createdSubscription.ID())
		ctx.Export(OpEventSubscriptionName, createdSubscription.Name)
		return nil
	}

	resourceGroup, systemTopicName, err := parseSystemTopicId(spec.SystemTopicId.GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to parse system topic id")
	}

	createdSubscription, err := eventgrid.NewSystemTopicEventSubscription(ctx,
		resourceName,
		buildSystemTopicArgs(spec, resourceGroup, systemTopicName),
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create eventgrid system topic event subscription %s", resourceName)
	}
	ctx.Export(OpEventSubscriptionId, createdSubscription.ID())
	ctx.Export(OpEventSubscriptionName, createdSubscription.Name)
	return nil
}

// deliverySchema resolves the platform default (EventGridSchema,
// mirroring Azure's own) so both engines render the schema explicitly.
// The default is only deliverable from EventGridSchema-input sources:
// ARM rejects the create (InvalidRequest, server-side only -- the
// provider does not validate the pairing) when the source topic's
// input_schema is CloudEventSchemaV1_0, so manifests subscribing to a
// CloudEvents topic must set the spec field explicitly.
func deliverySchema(spec *azureeventgrideventsubscriptionv1alpha1.AzureEventgridEventSubscriptionSpec) string {
	if spec.EventDeliverySchema != nil && *spec.EventDeliverySchema != "" {
		return *spec.EventDeliverySchema
	}
	return "EventGridSchema"
}

func intPtrInput(value *int32) pulumi.IntPtrInput {
	if value == nil {
		return nil
	}
	return pulumi.IntPtr(int(*value))
}

func buildScopedArgs(spec *azureeventgrideventsubscriptionv1alpha1.AzureEventgridEventSubscriptionSpec) *eventgrid.EventSubscriptionArgs {
	args := &eventgrid.EventSubscriptionArgs{
		Name:  pulumi.String(spec.Name),
		Scope: pulumi.String(spec.Scope.GetValue()),

		// Always sent (platform default mirrors Azure's). Create-only.
		EventDeliverySchema: pulumi.String(deliverySchema(spec)),

		AdvancedFilteringOnArraysEnabled: pulumi.Bool(spec.AdvancedFilteringOnArraysEnabled != nil && *spec.AdvancedFilteringOnArraysEnabled),

		Labels: pulumi.ToStringArray(spec.Labels),
	}

	// Sent only when set -- the service owns the default (never
	// expires).
	if spec.ExpirationTimeUtc != nil && *spec.ExpirationTimeUtc != "" {
		args.ExpirationTimeUtc = pulumi.StringPtr(*spec.ExpirationTimeUtc)
	}

	// An empty list means ALL event types the source emits -- omitted,
	// mirroring the Terraform module's null.
	if len(spec.IncludedEventTypes) > 0 {
		args.IncludedEventTypes = pulumi.ToStringArray(spec.IncludedEventTypes)
	}

	// The destination union -- exactly one arm (spec-enforced).
	destination := spec.Destination
	if destination.AzureFunction != nil {
		args.AzureFunctionEndpoint = &eventgrid.EventSubscriptionAzureFunctionEndpointArgs{
			FunctionId:                    pulumi.String(destination.AzureFunction.FunctionId.GetValue()),
			MaxEventsPerBatch:             intPtrInput(destination.AzureFunction.MaxEventsPerBatch),
			PreferredBatchSizeInKilobytes: intPtrInput(destination.AzureFunction.PreferredBatchSizeInKilobytes),
		}
	}
	if destination.EventhubId != nil {
		args.EventhubEndpointId = pulumi.StringPtr(destination.EventhubId.GetValue())
	}
	if destination.HybridConnectionId != nil {
		args.HybridConnectionEndpointId = pulumi.StringPtr(destination.HybridConnectionId.GetValue())
	}
	if destination.ServiceBusQueueId != nil {
		args.ServiceBusQueueEndpointId = pulumi.StringPtr(destination.ServiceBusQueueId.GetValue())
	}
	if destination.ServiceBusTopicId != nil {
		args.ServiceBusTopicEndpointId = pulumi.StringPtr(destination.ServiceBusTopicId.GetValue())
	}
	if destination.StorageQueue != nil {
		args.StorageQueueEndpoint = &eventgrid.EventSubscriptionStorageQueueEndpointArgs{
			StorageAccountId:                pulumi.String(destination.StorageQueue.StorageAccountId.GetValue()),
			QueueName:                       pulumi.String(destination.StorageQueue.QueueName),
			QueueMessageTimeToLiveInSeconds: intPtrInput(destination.StorageQueue.QueueMessageTimeToLiveInSeconds),
		}
	}
	if destination.Webhook != nil {
		webhookArgs := &eventgrid.EventSubscriptionWebhookEndpointArgs{
			Url:                           pulumi.String(destination.Webhook.Url),
			MaxEventsPerBatch:             intPtrInput(destination.Webhook.MaxEventsPerBatch),
			PreferredBatchSizeInKilobytes: intPtrInput(destination.Webhook.PreferredBatchSizeInKilobytes),
		}
		// Entra fields are sent only when set -- the service treats an
		// absent field and an empty one identically.
		if destination.Webhook.ActiveDirectoryTenantId != "" {
			webhookArgs.ActiveDirectoryTenantId = pulumi.StringPtr(destination.Webhook.ActiveDirectoryTenantId)
		}
		if destination.Webhook.ActiveDirectoryAppIdOrUri != "" {
			webhookArgs.ActiveDirectoryAppIdOrUri = pulumi.StringPtr(destination.Webhook.ActiveDirectoryAppIdOrUri)
		}
		args.WebhookEndpoint = webhookArgs
	}

	if spec.DeliveryIdentity != nil {
		identityArgs := &eventgrid.EventSubscriptionDeliveryIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.DeliveryIdentity.Type]),
		}
		if spec.DeliveryIdentity.UserAssignedIdentity != nil {
			identityArgs.UserAssignedIdentity = pulumi.StringPtr(spec.DeliveryIdentity.UserAssignedIdentity.GetValue())
		}
		args.DeliveryIdentity = identityArgs
	}

	// NOTE: Azure ignores delivery properties on storage-queue
	// destinations (queue messages carry no custom properties) -- the
	// entries pass through unfiltered so the two engines stay
	// identical; the spec documents the service behavior.
	if len(spec.DeliveryProperties) > 0 {
		properties := eventgrid.EventSubscriptionDeliveryPropertyArray{}
		for _, property := range spec.DeliveryProperties {
			propertyArgs := &eventgrid.EventSubscriptionDeliveryPropertyArgs{
				HeaderName: pulumi.String(property.HeaderName),
				Type:       pulumi.String(property.Type),
				Secret:     pulumi.BoolPtr(property.Secret),
			}
			if property.Value != nil {
				propertyArgs.Value = pulumi.StringPtr(property.Value.GetValue())
			}
			if property.SourceField != "" {
				propertyArgs.SourceField = pulumi.StringPtr(property.SourceField)
			}
			properties = append(properties, propertyArgs)
		}
		args.DeliveryProperties = properties
	}

	if spec.DeadLetter != nil {
		args.StorageBlobDeadLetterDestination = &eventgrid.EventSubscriptionStorageBlobDeadLetterDestinationArgs{
			StorageAccountId:         pulumi.String(spec.DeadLetter.StorageAccountId.GetValue()),
			StorageBlobContainerName: pulumi.String(spec.DeadLetter.StorageBlobContainerName),
		}
	}

	if spec.DeadLetterIdentity != nil {
		identityArgs := &eventgrid.EventSubscriptionDeadLetterIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.DeadLetterIdentity.Type]),
		}
		if spec.DeadLetterIdentity.UserAssignedIdentity != nil {
			identityArgs.UserAssignedIdentity = pulumi.StringPtr(spec.DeadLetterIdentity.UserAssignedIdentity.GetValue())
		}
		args.DeadLetterIdentity = identityArgs
	}

	// Sent only when set -- Azure's defaults (30 attempts / 1440
	// minutes) echo back on read otherwise.
	if spec.RetryPolicy != nil {
		args.RetryPolicy = &eventgrid.EventSubscriptionRetryPolicyArgs{
			MaxDeliveryAttempts: pulumi.Int(int(spec.RetryPolicy.MaxDeliveryAttempts)),
			EventTimeToLive:     pulumi.Int(int(spec.RetryPolicy.EventTimeToLive)),
		}
	}

	if spec.SubjectFilter != nil {
		subjectArgs := &eventgrid.EventSubscriptionSubjectFilterArgs{}
		if spec.SubjectFilter.SubjectBeginsWith != "" {
			subjectArgs.SubjectBeginsWith = pulumi.StringPtr(spec.SubjectFilter.SubjectBeginsWith)
		}
		if spec.SubjectFilter.SubjectEndsWith != "" {
			subjectArgs.SubjectEndsWith = pulumi.StringPtr(spec.SubjectFilter.SubjectEndsWith)
		}
		if spec.SubjectFilter.CaseSensitive != nil {
			subjectArgs.CaseSensitive = pulumi.BoolPtr(*spec.SubjectFilter.CaseSensitive)
		}
		args.SubjectFilter = subjectArgs
	}

	if filter := spec.AdvancedFilter; filter != nil {
		filterArgs := &eventgrid.EventSubscriptionAdvancedFilterArgs{}

		boolEquals := eventgrid.EventSubscriptionAdvancedFilterBoolEqualArray{}
		for _, f := range filter.BoolEquals {
			boolEquals = append(boolEquals, &eventgrid.EventSubscriptionAdvancedFilterBoolEqualArgs{
				Key:   pulumi.String(f.Key),
				Value: pulumi.Bool(f.Value),
			})
		}
		if len(boolEquals) > 0 {
			filterArgs.BoolEquals = boolEquals
		}

		numberGreaterThans := eventgrid.EventSubscriptionAdvancedFilterNumberGreaterThanArray{}
		for _, f := range filter.NumberGreaterThan {
			numberGreaterThans = append(numberGreaterThans, &eventgrid.EventSubscriptionAdvancedFilterNumberGreaterThanArgs{
				Key:   pulumi.String(f.Key),
				Value: pulumi.Float64(f.Value),
			})
		}
		if len(numberGreaterThans) > 0 {
			filterArgs.NumberGreaterThans = numberGreaterThans
		}

		numberGreaterThanOrEquals := eventgrid.EventSubscriptionAdvancedFilterNumberGreaterThanOrEqualArray{}
		for _, f := range filter.NumberGreaterThanOrEquals {
			numberGreaterThanOrEquals = append(numberGreaterThanOrEquals, &eventgrid.EventSubscriptionAdvancedFilterNumberGreaterThanOrEqualArgs{
				Key:   pulumi.String(f.Key),
				Value: pulumi.Float64(f.Value),
			})
		}
		if len(numberGreaterThanOrEquals) > 0 {
			filterArgs.NumberGreaterThanOrEquals = numberGreaterThanOrEquals
		}

		numberLessThans := eventgrid.EventSubscriptionAdvancedFilterNumberLessThanArray{}
		for _, f := range filter.NumberLessThan {
			numberLessThans = append(numberLessThans, &eventgrid.EventSubscriptionAdvancedFilterNumberLessThanArgs{
				Key:   pulumi.String(f.Key),
				Value: pulumi.Float64(f.Value),
			})
		}
		if len(numberLessThans) > 0 {
			filterArgs.NumberLessThans = numberLessThans
		}

		numberLessThanOrEquals := eventgrid.EventSubscriptionAdvancedFilterNumberLessThanOrEqualArray{}
		for _, f := range filter.NumberLessThanOrEquals {
			numberLessThanOrEquals = append(numberLessThanOrEquals, &eventgrid.EventSubscriptionAdvancedFilterNumberLessThanOrEqualArgs{
				Key:   pulumi.String(f.Key),
				Value: pulumi.Float64(f.Value),
			})
		}
		if len(numberLessThanOrEquals) > 0 {
			filterArgs.NumberLessThanOrEquals = numberLessThanOrEquals
		}

		numberIns := eventgrid.EventSubscriptionAdvancedFilterNumberInArray{}
		for _, f := range filter.NumberIn {
			numberIns = append(numberIns, &eventgrid.EventSubscriptionAdvancedFilterNumberInArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToFloat64Array(f.Values),
			})
		}
		if len(numberIns) > 0 {
			filterArgs.NumberIns = numberIns
		}

		numberNotIns := eventgrid.EventSubscriptionAdvancedFilterNumberNotInArray{}
		for _, f := range filter.NumberNotIn {
			numberNotIns = append(numberNotIns, &eventgrid.EventSubscriptionAdvancedFilterNumberNotInArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToFloat64Array(f.Values),
			})
		}
		if len(numberNotIns) > 0 {
			filterArgs.NumberNotIns = numberNotIns
		}

		// The provider's range shape is a list of [from, to] pairs; the
		// spec's named-message shape renders to it here.
		numberInRanges := eventgrid.EventSubscriptionAdvancedFilterNumberInRangeArray{}
		for _, f := range filter.NumberInRange {
			numberInRanges = append(numberInRanges, &eventgrid.EventSubscriptionAdvancedFilterNumberInRangeArgs{
				Key:    pulumi.String(f.Key),
				Values: rangePairs(f.Ranges),
			})
		}
		if len(numberInRanges) > 0 {
			filterArgs.NumberInRanges = numberInRanges
		}

		numberNotInRanges := eventgrid.EventSubscriptionAdvancedFilterNumberNotInRangeArray{}
		for _, f := range filter.NumberNotInRange {
			numberNotInRanges = append(numberNotInRanges, &eventgrid.EventSubscriptionAdvancedFilterNumberNotInRangeArgs{
				Key:    pulumi.String(f.Key),
				Values: rangePairs(f.Ranges),
			})
		}
		if len(numberNotInRanges) > 0 {
			filterArgs.NumberNotInRanges = numberNotInRanges
		}

		stringBeginsWiths := eventgrid.EventSubscriptionAdvancedFilterStringBeginsWithArray{}
		for _, f := range filter.StringBeginsWith {
			stringBeginsWiths = append(stringBeginsWiths, &eventgrid.EventSubscriptionAdvancedFilterStringBeginsWithArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringBeginsWiths) > 0 {
			filterArgs.StringBeginsWiths = stringBeginsWiths
		}

		stringNotBeginsWiths := eventgrid.EventSubscriptionAdvancedFilterStringNotBeginsWithArray{}
		for _, f := range filter.StringNotBeginsWith {
			stringNotBeginsWiths = append(stringNotBeginsWiths, &eventgrid.EventSubscriptionAdvancedFilterStringNotBeginsWithArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringNotBeginsWiths) > 0 {
			filterArgs.StringNotBeginsWiths = stringNotBeginsWiths
		}

		stringEndsWiths := eventgrid.EventSubscriptionAdvancedFilterStringEndsWithArray{}
		for _, f := range filter.StringEndsWith {
			stringEndsWiths = append(stringEndsWiths, &eventgrid.EventSubscriptionAdvancedFilterStringEndsWithArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringEndsWiths) > 0 {
			filterArgs.StringEndsWiths = stringEndsWiths
		}

		stringNotEndsWiths := eventgrid.EventSubscriptionAdvancedFilterStringNotEndsWithArray{}
		for _, f := range filter.StringNotEndsWith {
			stringNotEndsWiths = append(stringNotEndsWiths, &eventgrid.EventSubscriptionAdvancedFilterStringNotEndsWithArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringNotEndsWiths) > 0 {
			filterArgs.StringNotEndsWiths = stringNotEndsWiths
		}

		stringContains := eventgrid.EventSubscriptionAdvancedFilterStringContainArray{}
		for _, f := range filter.StringContains {
			stringContains = append(stringContains, &eventgrid.EventSubscriptionAdvancedFilterStringContainArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringContains) > 0 {
			filterArgs.StringContains = stringContains
		}

		stringNotContains := eventgrid.EventSubscriptionAdvancedFilterStringNotContainArray{}
		for _, f := range filter.StringNotContains {
			stringNotContains = append(stringNotContains, &eventgrid.EventSubscriptionAdvancedFilterStringNotContainArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringNotContains) > 0 {
			filterArgs.StringNotContains = stringNotContains
		}

		stringIns := eventgrid.EventSubscriptionAdvancedFilterStringInArray{}
		for _, f := range filter.StringIn {
			stringIns = append(stringIns, &eventgrid.EventSubscriptionAdvancedFilterStringInArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringIns) > 0 {
			filterArgs.StringIns = stringIns
		}

		stringNotIns := eventgrid.EventSubscriptionAdvancedFilterStringNotInArray{}
		for _, f := range filter.StringNotIn {
			stringNotIns = append(stringNotIns, &eventgrid.EventSubscriptionAdvancedFilterStringNotInArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringNotIns) > 0 {
			filterArgs.StringNotIns = stringNotIns
		}

		isNotNulls := eventgrid.EventSubscriptionAdvancedFilterIsNotNullArray{}
		for _, f := range filter.IsNotNull {
			isNotNulls = append(isNotNulls, &eventgrid.EventSubscriptionAdvancedFilterIsNotNullArgs{
				Key: pulumi.String(f.Key),
			})
		}
		if len(isNotNulls) > 0 {
			filterArgs.IsNotNulls = isNotNulls
		}

		isNullOrUndefineds := eventgrid.EventSubscriptionAdvancedFilterIsNullOrUndefinedArray{}
		for _, f := range filter.IsNullOrUndefined {
			isNullOrUndefineds = append(isNullOrUndefineds, &eventgrid.EventSubscriptionAdvancedFilterIsNullOrUndefinedArgs{
				Key: pulumi.String(f.Key),
			})
		}
		if len(isNullOrUndefineds) > 0 {
			filterArgs.IsNullOrUndefineds = isNullOrUndefineds
		}

		args.AdvancedFilter = filterArgs
	}

	return args
}

// rangePairs renders the spec's named [from, to] messages into the
// provider's list-of-pairs shape.
func rangePairs(ranges []*azureeventgrideventsubscriptionv1alpha1.AzureEventgridEventSubscriptionNumberRange) pulumi.Float64ArrayArrayInput {
	pairs := pulumi.Float64ArrayArray{}
	for _, r := range ranges {
		pairs = append(pairs, pulumi.Float64Array{pulumi.Float64(r.From), pulumi.Float64(r.To)})
	}
	return pairs
}

// buildSystemTopicArgs is buildScopedArgs's twin over the SDK's
// SystemTopicEventSubscription* types -- the same spec surface mapped
// onto the system-topic resource's addressing (resource group + topic
// name). Keep the two builders in lockstep when either changes.
func buildSystemTopicArgs(
	spec *azureeventgrideventsubscriptionv1alpha1.AzureEventgridEventSubscriptionSpec,
	resourceGroup string,
	systemTopicName string,
) *eventgrid.SystemTopicEventSubscriptionArgs {
	args := &eventgrid.SystemTopicEventSubscriptionArgs{
		Name:              pulumi.String(spec.Name),
		SystemTopic:       pulumi.String(systemTopicName),
		ResourceGroupName: pulumi.String(resourceGroup),

		// Always sent (platform default mirrors Azure's). Create-only.
		EventDeliverySchema: pulumi.String(deliverySchema(spec)),

		AdvancedFilteringOnArraysEnabled: pulumi.Bool(spec.AdvancedFilteringOnArraysEnabled != nil && *spec.AdvancedFilteringOnArraysEnabled),

		Labels: pulumi.ToStringArray(spec.Labels),
	}

	// Sent only when set -- the service owns the default (never
	// expires).
	if spec.ExpirationTimeUtc != nil && *spec.ExpirationTimeUtc != "" {
		args.ExpirationTimeUtc = pulumi.StringPtr(*spec.ExpirationTimeUtc)
	}

	// An empty list means ALL event types the source emits -- omitted,
	// mirroring the Terraform module's null.
	if len(spec.IncludedEventTypes) > 0 {
		args.IncludedEventTypes = pulumi.ToStringArray(spec.IncludedEventTypes)
	}

	// The destination union -- exactly one arm (spec-enforced).
	destination := spec.Destination
	if destination.AzureFunction != nil {
		args.AzureFunctionEndpoint = &eventgrid.SystemTopicEventSubscriptionAzureFunctionEndpointArgs{
			FunctionId:                    pulumi.String(destination.AzureFunction.FunctionId.GetValue()),
			MaxEventsPerBatch:             intPtrInput(destination.AzureFunction.MaxEventsPerBatch),
			PreferredBatchSizeInKilobytes: intPtrInput(destination.AzureFunction.PreferredBatchSizeInKilobytes),
		}
	}
	if destination.EventhubId != nil {
		args.EventhubEndpointId = pulumi.StringPtr(destination.EventhubId.GetValue())
	}
	if destination.HybridConnectionId != nil {
		args.HybridConnectionEndpointId = pulumi.StringPtr(destination.HybridConnectionId.GetValue())
	}
	if destination.ServiceBusQueueId != nil {
		args.ServiceBusQueueEndpointId = pulumi.StringPtr(destination.ServiceBusQueueId.GetValue())
	}
	if destination.ServiceBusTopicId != nil {
		args.ServiceBusTopicEndpointId = pulumi.StringPtr(destination.ServiceBusTopicId.GetValue())
	}
	if destination.StorageQueue != nil {
		args.StorageQueueEndpoint = &eventgrid.SystemTopicEventSubscriptionStorageQueueEndpointArgs{
			StorageAccountId:                pulumi.String(destination.StorageQueue.StorageAccountId.GetValue()),
			QueueName:                       pulumi.String(destination.StorageQueue.QueueName),
			QueueMessageTimeToLiveInSeconds: intPtrInput(destination.StorageQueue.QueueMessageTimeToLiveInSeconds),
		}
	}
	if destination.Webhook != nil {
		webhookArgs := &eventgrid.SystemTopicEventSubscriptionWebhookEndpointArgs{
			Url:                           pulumi.String(destination.Webhook.Url),
			MaxEventsPerBatch:             intPtrInput(destination.Webhook.MaxEventsPerBatch),
			PreferredBatchSizeInKilobytes: intPtrInput(destination.Webhook.PreferredBatchSizeInKilobytes),
		}
		// Entra fields are sent only when set -- the service treats an
		// absent field and an empty one identically.
		if destination.Webhook.ActiveDirectoryTenantId != "" {
			webhookArgs.ActiveDirectoryTenantId = pulumi.StringPtr(destination.Webhook.ActiveDirectoryTenantId)
		}
		if destination.Webhook.ActiveDirectoryAppIdOrUri != "" {
			webhookArgs.ActiveDirectoryAppIdOrUri = pulumi.StringPtr(destination.Webhook.ActiveDirectoryAppIdOrUri)
		}
		args.WebhookEndpoint = webhookArgs
	}

	if spec.DeliveryIdentity != nil {
		identityArgs := &eventgrid.SystemTopicEventSubscriptionDeliveryIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.DeliveryIdentity.Type]),
		}
		if spec.DeliveryIdentity.UserAssignedIdentity != nil {
			identityArgs.UserAssignedIdentity = pulumi.StringPtr(spec.DeliveryIdentity.UserAssignedIdentity.GetValue())
		}
		args.DeliveryIdentity = identityArgs
	}

	// NOTE: Azure ignores delivery properties on storage-queue
	// destinations (queue messages carry no custom properties) -- the
	// entries pass through unfiltered so the two engines stay
	// identical; the spec documents the service behavior.
	if len(spec.DeliveryProperties) > 0 {
		properties := eventgrid.SystemTopicEventSubscriptionDeliveryPropertyArray{}
		for _, property := range spec.DeliveryProperties {
			propertyArgs := &eventgrid.SystemTopicEventSubscriptionDeliveryPropertyArgs{
				HeaderName: pulumi.String(property.HeaderName),
				Type:       pulumi.String(property.Type),
				Secret:     pulumi.BoolPtr(property.Secret),
			}
			if property.Value != nil {
				propertyArgs.Value = pulumi.StringPtr(property.Value.GetValue())
			}
			if property.SourceField != "" {
				propertyArgs.SourceField = pulumi.StringPtr(property.SourceField)
			}
			properties = append(properties, propertyArgs)
		}
		args.DeliveryProperties = properties
	}

	if spec.DeadLetter != nil {
		args.StorageBlobDeadLetterDestination = &eventgrid.SystemTopicEventSubscriptionStorageBlobDeadLetterDestinationArgs{
			StorageAccountId:         pulumi.String(spec.DeadLetter.StorageAccountId.GetValue()),
			StorageBlobContainerName: pulumi.String(spec.DeadLetter.StorageBlobContainerName),
		}
	}

	if spec.DeadLetterIdentity != nil {
		identityArgs := &eventgrid.SystemTopicEventSubscriptionDeadLetterIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.DeadLetterIdentity.Type]),
		}
		if spec.DeadLetterIdentity.UserAssignedIdentity != nil {
			identityArgs.UserAssignedIdentity = pulumi.StringPtr(spec.DeadLetterIdentity.UserAssignedIdentity.GetValue())
		}
		args.DeadLetterIdentity = identityArgs
	}

	// Sent only when set -- Azure's defaults (30 attempts / 1440
	// minutes) echo back on read otherwise.
	if spec.RetryPolicy != nil {
		args.RetryPolicy = &eventgrid.SystemTopicEventSubscriptionRetryPolicyArgs{
			MaxDeliveryAttempts: pulumi.Int(int(spec.RetryPolicy.MaxDeliveryAttempts)),
			EventTimeToLive:     pulumi.Int(int(spec.RetryPolicy.EventTimeToLive)),
		}
	}

	if spec.SubjectFilter != nil {
		subjectArgs := &eventgrid.SystemTopicEventSubscriptionSubjectFilterArgs{}
		if spec.SubjectFilter.SubjectBeginsWith != "" {
			subjectArgs.SubjectBeginsWith = pulumi.StringPtr(spec.SubjectFilter.SubjectBeginsWith)
		}
		if spec.SubjectFilter.SubjectEndsWith != "" {
			subjectArgs.SubjectEndsWith = pulumi.StringPtr(spec.SubjectFilter.SubjectEndsWith)
		}
		if spec.SubjectFilter.CaseSensitive != nil {
			subjectArgs.CaseSensitive = pulumi.BoolPtr(*spec.SubjectFilter.CaseSensitive)
		}
		args.SubjectFilter = subjectArgs
	}

	if filter := spec.AdvancedFilter; filter != nil {
		filterArgs := &eventgrid.SystemTopicEventSubscriptionAdvancedFilterArgs{}

		boolEquals := eventgrid.SystemTopicEventSubscriptionAdvancedFilterBoolEqualArray{}
		for _, f := range filter.BoolEquals {
			boolEquals = append(boolEquals, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterBoolEqualArgs{
				Key:   pulumi.String(f.Key),
				Value: pulumi.Bool(f.Value),
			})
		}
		if len(boolEquals) > 0 {
			filterArgs.BoolEquals = boolEquals
		}

		numberGreaterThans := eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberGreaterThanArray{}
		for _, f := range filter.NumberGreaterThan {
			numberGreaterThans = append(numberGreaterThans, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberGreaterThanArgs{
				Key:   pulumi.String(f.Key),
				Value: pulumi.Float64(f.Value),
			})
		}
		if len(numberGreaterThans) > 0 {
			filterArgs.NumberGreaterThans = numberGreaterThans
		}

		numberGreaterThanOrEquals := eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberGreaterThanOrEqualArray{}
		for _, f := range filter.NumberGreaterThanOrEquals {
			numberGreaterThanOrEquals = append(numberGreaterThanOrEquals, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberGreaterThanOrEqualArgs{
				Key:   pulumi.String(f.Key),
				Value: pulumi.Float64(f.Value),
			})
		}
		if len(numberGreaterThanOrEquals) > 0 {
			filterArgs.NumberGreaterThanOrEquals = numberGreaterThanOrEquals
		}

		numberLessThans := eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberLessThanArray{}
		for _, f := range filter.NumberLessThan {
			numberLessThans = append(numberLessThans, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberLessThanArgs{
				Key:   pulumi.String(f.Key),
				Value: pulumi.Float64(f.Value),
			})
		}
		if len(numberLessThans) > 0 {
			filterArgs.NumberLessThans = numberLessThans
		}

		numberLessThanOrEquals := eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberLessThanOrEqualArray{}
		for _, f := range filter.NumberLessThanOrEquals {
			numberLessThanOrEquals = append(numberLessThanOrEquals, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberLessThanOrEqualArgs{
				Key:   pulumi.String(f.Key),
				Value: pulumi.Float64(f.Value),
			})
		}
		if len(numberLessThanOrEquals) > 0 {
			filterArgs.NumberLessThanOrEquals = numberLessThanOrEquals
		}

		numberIns := eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberInArray{}
		for _, f := range filter.NumberIn {
			numberIns = append(numberIns, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberInArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToFloat64Array(f.Values),
			})
		}
		if len(numberIns) > 0 {
			filterArgs.NumberIns = numberIns
		}

		numberNotIns := eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberNotInArray{}
		for _, f := range filter.NumberNotIn {
			numberNotIns = append(numberNotIns, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberNotInArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToFloat64Array(f.Values),
			})
		}
		if len(numberNotIns) > 0 {
			filterArgs.NumberNotIns = numberNotIns
		}

		// The provider's range shape is a list of [from, to] pairs; the
		// spec's named-message shape renders to it here.
		numberInRanges := eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberInRangeArray{}
		for _, f := range filter.NumberInRange {
			numberInRanges = append(numberInRanges, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberInRangeArgs{
				Key:    pulumi.String(f.Key),
				Values: rangePairs(f.Ranges),
			})
		}
		if len(numberInRanges) > 0 {
			filterArgs.NumberInRanges = numberInRanges
		}

		numberNotInRanges := eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberNotInRangeArray{}
		for _, f := range filter.NumberNotInRange {
			numberNotInRanges = append(numberNotInRanges, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterNumberNotInRangeArgs{
				Key:    pulumi.String(f.Key),
				Values: rangePairs(f.Ranges),
			})
		}
		if len(numberNotInRanges) > 0 {
			filterArgs.NumberNotInRanges = numberNotInRanges
		}

		stringBeginsWiths := eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringBeginsWithArray{}
		for _, f := range filter.StringBeginsWith {
			stringBeginsWiths = append(stringBeginsWiths, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringBeginsWithArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringBeginsWiths) > 0 {
			filterArgs.StringBeginsWiths = stringBeginsWiths
		}

		stringNotBeginsWiths := eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringNotBeginsWithArray{}
		for _, f := range filter.StringNotBeginsWith {
			stringNotBeginsWiths = append(stringNotBeginsWiths, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringNotBeginsWithArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringNotBeginsWiths) > 0 {
			filterArgs.StringNotBeginsWiths = stringNotBeginsWiths
		}

		stringEndsWiths := eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringEndsWithArray{}
		for _, f := range filter.StringEndsWith {
			stringEndsWiths = append(stringEndsWiths, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringEndsWithArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringEndsWiths) > 0 {
			filterArgs.StringEndsWiths = stringEndsWiths
		}

		stringNotEndsWiths := eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringNotEndsWithArray{}
		for _, f := range filter.StringNotEndsWith {
			stringNotEndsWiths = append(stringNotEndsWiths, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringNotEndsWithArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringNotEndsWiths) > 0 {
			filterArgs.StringNotEndsWiths = stringNotEndsWiths
		}

		stringContains := eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringContainArray{}
		for _, f := range filter.StringContains {
			stringContains = append(stringContains, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringContainArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringContains) > 0 {
			filterArgs.StringContains = stringContains
		}

		stringNotContains := eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringNotContainArray{}
		for _, f := range filter.StringNotContains {
			stringNotContains = append(stringNotContains, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringNotContainArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringNotContains) > 0 {
			filterArgs.StringNotContains = stringNotContains
		}

		stringIns := eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringInArray{}
		for _, f := range filter.StringIn {
			stringIns = append(stringIns, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringInArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringIns) > 0 {
			filterArgs.StringIns = stringIns
		}

		stringNotIns := eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringNotInArray{}
		for _, f := range filter.StringNotIn {
			stringNotIns = append(stringNotIns, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterStringNotInArgs{
				Key:    pulumi.String(f.Key),
				Values: pulumi.ToStringArray(f.Values),
			})
		}
		if len(stringNotIns) > 0 {
			filterArgs.StringNotIns = stringNotIns
		}

		isNotNulls := eventgrid.SystemTopicEventSubscriptionAdvancedFilterIsNotNullArray{}
		for _, f := range filter.IsNotNull {
			isNotNulls = append(isNotNulls, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterIsNotNullArgs{
				Key: pulumi.String(f.Key),
			})
		}
		if len(isNotNulls) > 0 {
			filterArgs.IsNotNulls = isNotNulls
		}

		isNullOrUndefineds := eventgrid.SystemTopicEventSubscriptionAdvancedFilterIsNullOrUndefinedArray{}
		for _, f := range filter.IsNullOrUndefined {
			isNullOrUndefineds = append(isNullOrUndefineds, &eventgrid.SystemTopicEventSubscriptionAdvancedFilterIsNullOrUndefinedArgs{
				Key: pulumi.String(f.Key),
			})
		}
		if len(isNullOrUndefineds) > 0 {
			filterArgs.IsNullOrUndefineds = isNullOrUndefineds
		}

		args.AdvancedFilter = filterArgs
	}

	return args
}
