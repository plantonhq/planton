package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	azureservicebussubscriptionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureservicebussubscription/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/servicebus"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureservicebussubscriptionv1alpha1.AzureServiceBusSubscriptionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureServiceBusSubscription.Spec

	// The topic and namespace names, parsed from the resolved topic ARM id
	// for the outputs -- consumers receive by the
	// namespace/topic/subscription triple.
	namespaceName, topicName, err := parseTopicId(locals.TopicId)
	if err != nil {
		return err
	}
	locals.NamespaceName = namespaceName
	locals.TopicName = topicName

	// The subscription, addressed by the parent topic's ARM id (azurerm's
	// v4 child-addressing grain). Consumer semantics -- locks, delivery
	// counts, sessions, dead-lettering -- live here, not on the topic.
	subscriptionArgs := &servicebus.SubscriptionArgs{
		// ForceNew: renaming replaces the subscription and resets its read
		// position -- undelivered messages in the old subscription are lost.
		Name:    pulumi.String(spec.SubscriptionName),
		TopicId: pulumi.String(locals.TopicId),
		// Required: Azure has no server default for a subscription's
		// delivery tolerance.
		MaxDeliveryCount: pulumi.Int(int(spec.MaxDeliveryCount)),
		Status:           pulumi.String(statusStrings[spec.Status]),
	}

	// Lifecycle dials -- unset leaves Azure's defaults in place (lock PT1M,
	// TTL inherited from the topic, never auto-delete).
	if spec.LockDuration != nil {
		subscriptionArgs.LockDuration = pulumi.StringPtr(spec.GetLockDuration())
	}
	if spec.DefaultMessageTtl != nil {
		subscriptionArgs.DefaultMessageTtl = pulumi.StringPtr(spec.GetDefaultMessageTtl())
	}
	if spec.AutoDeleteOnIdle != nil {
		subscriptionArgs.AutoDeleteOnIdle = pulumi.StringPtr(spec.GetAutoDeleteOnIdle())
	}
	if spec.DeadLetteringOnMessageExpiration != nil {
		subscriptionArgs.DeadLetteringOnMessageExpiration = pulumi.BoolPtr(spec.GetDeadLetteringOnMessageExpiration())
	}

	// Presence-guarded to Azure's default (true): IaC inputs built from a
	// manifest materialize proto defaults, but direct paths do not.
	subscriptionArgs.DeadLetteringOnFilterEvaluationError = pulumi.Bool(presenceGuardedBool(spec.DeadLetteringOnFilterEvaluationError, true))

	// ForceNew: the session model is fixed at creation.
	if spec.RequiresSession != nil {
		subscriptionArgs.RequiresSession = pulumi.BoolPtr(spec.GetRequiresSession())
	}
	if spec.BatchedOperationsEnabled != nil {
		subscriptionArgs.BatchedOperationsEnabled = pulumi.BoolPtr(spec.GetBatchedOperationsEnabled())
	}

	// Routing chains: targets are entity NAMES in the same namespace (not
	// ARM ids). The classic fan-out-then-collect pattern: subscriptions
	// filter, forwarding funnels matches into a work queue. References
	// resolve to the target's queue_name/topic_name output before the
	// module runs.
	if spec.ForwardTo.GetValue() != "" {
		subscriptionArgs.ForwardTo = pulumi.StringPtr(spec.ForwardTo.GetValue())
	}
	if spec.ForwardDeadLetteredMessagesTo.GetValue() != "" {
		subscriptionArgs.ForwardDeadLetteredMessagesTo = pulumi.StringPtr(spec.ForwardDeadLetteredMessagesTo.GetValue())
	}

	// The JMS 2.0 client-affine binding. Azure stores the entity as
	// {name}${client_id}$D internally; the provider round-trips the
	// user-facing name.
	if spec.ClientScopedSubscription != nil {
		subscriptionArgs.ClientScopedSubscriptionEnabled = pulumi.Bool(true)
		clientScopedArgs := &servicebus.SubscriptionClientScopedSubscriptionArgs{}
		if spec.ClientScopedSubscription.ClientId != "" {
			clientScopedArgs.ClientId = pulumi.StringPtr(spec.ClientScopedSubscription.ClientId)
		}
		if spec.ClientScopedSubscription.Shareable != nil {
			clientScopedArgs.IsClientScopedSubscriptionShareable = pulumi.BoolPtr(spec.ClientScopedSubscription.GetShareable())
		}
		subscriptionArgs.ClientScopedSubscription = clientScopedArgs
	}

	createdSubscription, err := servicebus.NewSubscription(ctx,
		spec.SubscriptionName,
		subscriptionArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Service Bus subscription %s", spec.SubscriptionName)
	}

	// Folded filter rules -- a rule has no life outside its subscription
	// and nothing references one, so they ship as part of the subscription
	// document. OR semantics: a message is delivered once if ANY rule
	// matches -- ALONGSIDE Azure's auto-created "$Default" catch-all,
	// which cannot be declared here (the provider's import check refuses
	// to adopt the service-created rule; the spec reserves the name).
	// Restrictive delivery = remove the catch-all once, out-of-band.
	for _, rule := range spec.Rules {
		ruleArgs := &servicebus.SubscriptionRuleArgs{
			Name:           pulumi.String(rule.RuleName),
			SubscriptionId: createdSubscription.ID(),
			FilterType:     pulumi.String(filterTypeStrings[rule.FilterType]),
		}

		// The SQL path: expression required with SqlFilter (spec-enforced
		// XOR). The optional action annotates matched messages before
		// delivery.
		if rule.SqlFilter != nil {
			ruleArgs.SqlFilter = pulumi.StringPtr(rule.GetSqlFilter())
		}
		if rule.Action != nil {
			ruleArgs.Action = pulumi.StringPtr(rule.GetAction())
		}

		// The correlation path: equality matching on correlation properties
		// -- cheaper than SQL at high throughput.
		if rule.CorrelationFilter != nil {
			correlationArgs := &servicebus.SubscriptionRuleCorrelationFilterArgs{}
			if rule.CorrelationFilter.CorrelationId != nil {
				correlationArgs.CorrelationId = pulumi.StringPtr(rule.CorrelationFilter.GetCorrelationId())
			}
			if rule.CorrelationFilter.MessageId != nil {
				correlationArgs.MessageId = pulumi.StringPtr(rule.CorrelationFilter.GetMessageId())
			}
			if rule.CorrelationFilter.To != nil {
				correlationArgs.To = pulumi.StringPtr(rule.CorrelationFilter.GetTo())
			}
			if rule.CorrelationFilter.ReplyTo != nil {
				correlationArgs.ReplyTo = pulumi.StringPtr(rule.CorrelationFilter.GetReplyTo())
			}
			if rule.CorrelationFilter.Label != nil {
				correlationArgs.Label = pulumi.StringPtr(rule.CorrelationFilter.GetLabel())
			}
			if rule.CorrelationFilter.SessionId != nil {
				correlationArgs.SessionId = pulumi.StringPtr(rule.CorrelationFilter.GetSessionId())
			}
			if rule.CorrelationFilter.ReplyToSessionId != nil {
				correlationArgs.ReplyToSessionId = pulumi.StringPtr(rule.CorrelationFilter.GetReplyToSessionId())
			}
			if rule.CorrelationFilter.ContentType != nil {
				correlationArgs.ContentType = pulumi.StringPtr(rule.CorrelationFilter.GetContentType())
			}
			if len(rule.CorrelationFilter.Properties) > 0 {
				correlationArgs.Properties = pulumi.ToStringMap(rule.CorrelationFilter.Properties)
			}
			ruleArgs.CorrelationFilter = correlationArgs
		}

		// The Pulumi LOGICAL name must not carry "$" (legal in ARM rule
		// names; the ARM-facing Name argument keeps it) -- URN-unfriendly
		// characters are stripped from the logical identity only.
		logicalName := fmt.Sprintf("%s-%s", spec.SubscriptionName, strings.ReplaceAll(rule.RuleName, "$", ""))
		_, err := servicebus.NewSubscriptionRule(ctx,
			logicalName,
			ruleArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdSubscription))
		if err != nil {
			return errors.Wrapf(err, "failed to create subscription rule %s", rule.RuleName)
		}
	}

	// Export outputs.
	ctx.Export(OpSubscriptionId, createdSubscription.ID())
	ctx.Export(OpSubscriptionName, createdSubscription.Name)
	ctx.Export(OpTopicName, pulumi.String(locals.TopicName))
	ctx.Export(OpNamespaceName, pulumi.String(locals.NamespaceName))

	return nil
}

// presenceGuardedBool returns the field's value when set and the proto
// default otherwise -- default materialization is middleware behavior, not a
// wire guarantee.
func presenceGuardedBool(field *bool, defaultValue bool) bool {
	if field == nil {
		return defaultValue
	}
	return *field
}
