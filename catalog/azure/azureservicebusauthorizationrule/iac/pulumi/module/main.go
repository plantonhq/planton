package module

import (
	"github.com/pkg/errors"
	azureservicebusauthorizationrulev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureservicebusauthorizationrule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/servicebus"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureservicebusauthorizationrulev1alpha1.AzureServiceBusAuthorizationRuleStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureServiceBusAuthorizationRule.Spec

	// One kind, three ARM types: SAS authorization rules exist at
	// namespace, queue, and topic scope with byte-identical shapes (name +
	// the listen/send/manage trio + six key/connection-string outputs).
	// The spec's exactly-one-scope XOR picks the resource; the six output
	// faces are identical across all three, so the export block below is
	// shared. After create/delete on a geo-DR-paired Premium namespace the
	// provider waits for pairing replication to settle -- no module-side
	// ordering is needed.
	var (
		ruleId                         pulumi.StringOutput
		ruleName                       pulumi.StringOutput
		primaryKey                     pulumi.StringOutput
		secondaryKey                   pulumi.StringOutput
		primaryConnectionString        pulumi.StringOutput
		secondaryConnectionString      pulumi.StringOutput
		primaryConnectionStringAlias   pulumi.StringOutput
		secondaryConnectionStringAlias pulumi.StringOutput
	)

	switch {
	case locals.NamespaceId != "":
		// Namespace-wide rights: every queue and topic in the namespace.
		createdRule, err := servicebus.NewNamespaceAuthorizationRule(ctx,
			spec.RuleName,
			&servicebus.NamespaceAuthorizationRuleArgs{
				// ForceNew: renaming replaces the rule and regenerates keys.
				Name:        pulumi.String(spec.RuleName),
				NamespaceId: pulumi.String(locals.NamespaceId),
				Listen:      pulumi.Bool(locals.Listen),
				Send:        pulumi.Bool(locals.Send),
				Manage:      pulumi.Bool(locals.Manage),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create namespace-scoped authorization rule %s", spec.RuleName)
		}
		ruleId = createdRule.ID().ToStringOutput()
		ruleName = createdRule.Name
		primaryKey = createdRule.PrimaryKey
		secondaryKey = createdRule.SecondaryKey
		primaryConnectionString = createdRule.PrimaryConnectionString
		secondaryConnectionString = createdRule.SecondaryConnectionString
		primaryConnectionStringAlias = createdRule.PrimaryConnectionStringAlias
		secondaryConnectionStringAlias = createdRule.SecondaryConnectionStringAlias

	case locals.QueueId != "":
		// Single-queue rights -- the least-privilege choice for
		// point-to-point workloads.
		createdRule, err := servicebus.NewQueueAuthorizationRule(ctx,
			spec.RuleName,
			&servicebus.QueueAuthorizationRuleArgs{
				Name:    pulumi.String(spec.RuleName),
				QueueId: pulumi.String(locals.QueueId),
				Listen:  pulumi.Bool(locals.Listen),
				Send:    pulumi.Bool(locals.Send),
				Manage:  pulumi.Bool(locals.Manage),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create queue-scoped authorization rule %s", spec.RuleName)
		}
		ruleId = createdRule.ID().ToStringOutput()
		ruleName = createdRule.Name
		primaryKey = createdRule.PrimaryKey
		secondaryKey = createdRule.SecondaryKey
		primaryConnectionString = createdRule.PrimaryConnectionString
		secondaryConnectionString = createdRule.SecondaryConnectionString
		primaryConnectionStringAlias = createdRule.PrimaryConnectionStringAlias
		secondaryConnectionStringAlias = createdRule.SecondaryConnectionStringAlias

	case locals.TopicId != "":
		// Single-topic rights (sending to the topic, receiving through its
		// subscriptions).
		createdRule, err := servicebus.NewTopicAuthorizationRule(ctx,
			spec.RuleName,
			&servicebus.TopicAuthorizationRuleArgs{
				Name:    pulumi.String(spec.RuleName),
				TopicId: pulumi.String(locals.TopicId),
				Listen:  pulumi.Bool(locals.Listen),
				Send:    pulumi.Bool(locals.Send),
				Manage:  pulumi.Bool(locals.Manage),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create topic-scoped authorization rule %s", spec.RuleName)
		}
		ruleId = createdRule.ID().ToStringOutput()
		ruleName = createdRule.Name
		primaryKey = createdRule.PrimaryKey
		secondaryKey = createdRule.SecondaryKey
		primaryConnectionString = createdRule.PrimaryConnectionString
		secondaryConnectionString = createdRule.SecondaryConnectionString
		primaryConnectionStringAlias = createdRule.PrimaryConnectionStringAlias
		secondaryConnectionStringAlias = createdRule.SecondaryConnectionStringAlias

	default:
		// Unreachable behind the spec's exactly-one-scope CEL; guards a
		// stack input that bypassed validation.
		return errors.New("exactly one of namespace_id, queue_id, or topic_id must be set")
	}

	// Export stack outputs -- identical faces regardless of scope. Alias
	// connection strings are only populated when the namespace carries a
	// geo-DR pairing.
	ctx.Export(OpAuthorizationRuleId, ruleId)
	ctx.Export(OpRuleName, ruleName)
	ctx.Export(OpPrimaryKey, pulumi.ToSecret(primaryKey))
	ctx.Export(OpSecondaryKey, pulumi.ToSecret(secondaryKey))
	ctx.Export(OpPrimaryConnectionString, pulumi.ToSecret(primaryConnectionString))
	ctx.Export(OpSecondaryConnectionString, pulumi.ToSecret(secondaryConnectionString))
	ctx.Export(OpPrimaryConnectionStringAlias, pulumi.ToSecret(primaryConnectionStringAlias))
	ctx.Export(OpSecondaryConnectionStringAlias, pulumi.ToSecret(secondaryConnectionStringAlias))

	return nil
}
