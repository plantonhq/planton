package module

import (
	"github.com/pkg/errors"
	azureeventhubauthorizationrulev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubauthorizationrule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventhub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureeventhubauthorizationrulev1alpha1.AzureEventHubAuthorizationRuleIacInput) error {
	locals, err := initializeLocals(ctx, iacInput)
	if err != nil {
		return err
	}

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventHubAuthorizationRule.Spec

	// One kind, two ARM types: SAS authorization rules exist at namespace
	// and event hub scope with byte-identical shapes (name + the
	// listen/send/manage trio + six key/connection-string outputs). The
	// spec's exactly-one-scope XOR picks the resource; the output faces
	// are identical across both, so the export block below is shared.
	listen := pulumi.Bool(presenceGuardedBool(spec.Listen))
	send := pulumi.Bool(presenceGuardedBool(spec.Send))
	manage := pulumi.Bool(presenceGuardedBool(spec.Manage))

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
	case locals.IsNamespaceScoped:
		// Namespace-wide rights: every hub in the namespace.
		createdRule, err := eventhub.NewEventHubNamespaceAuthorizationRule(ctx,
			spec.RuleName,
			&eventhub.EventHubNamespaceAuthorizationRuleArgs{
				// ForceNew: renaming replaces the rule and regenerates keys.
				Name:              pulumi.String(spec.RuleName),
				NamespaceName:     pulumi.String(locals.NamespaceName),
				ResourceGroupName: pulumi.String(locals.ResourceGroupName),
				Listen:            listen,
				Send:              send,
				Manage:            manage,
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

	case locals.IsHubScoped:
		// Single-hub rights -- the least-privilege choice for per-stream
		// producers and consumers.
		createdRule, err := eventhub.NewEventHubAuthorizationRule(ctx,
			spec.RuleName,
			&eventhub.EventHubAuthorizationRuleArgs{
				Name:              pulumi.String(spec.RuleName),
				NamespaceName:     pulumi.String(locals.NamespaceName),
				EventhubName:      pulumi.String(locals.EventHubName),
				ResourceGroupName: pulumi.String(locals.ResourceGroupName),
				Listen:            listen,
				Send:              send,
				Manage:            manage,
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create hub-scoped authorization rule %s", spec.RuleName)
		}
		ruleId = createdRule.ID().ToStringOutput()
		ruleName = createdRule.Name
		primaryKey = createdRule.PrimaryKey
		secondaryKey = createdRule.SecondaryKey
		primaryConnectionString = createdRule.PrimaryConnectionString
		secondaryConnectionString = createdRule.SecondaryConnectionString
		primaryConnectionStringAlias = createdRule.PrimaryConnectionStringAlias
		secondaryConnectionStringAlias = createdRule.SecondaryConnectionStringAlias
	}

	// Export outputs -- identical faces regardless of scope. Alias
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
