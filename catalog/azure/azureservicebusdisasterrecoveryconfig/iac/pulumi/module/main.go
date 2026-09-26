package module

import (
	"github.com/pkg/errors"
	azureservicebusdisasterrecoveryconfigv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureservicebusdisasterrecoveryconfig/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/servicebus"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureservicebusdisasterrecoveryconfigv1alpha1.AzureServiceBusDisasterRecoveryConfigStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureServiceBusDisasterRecoveryConfig.Spec

	// The geo-DR pairing: metadata (entities, rules, SAS rules -- not
	// message data) continuously replicates primary -> partner, and the
	// alias DNS name fronts whichever namespace is currently primary.
	//
	// Provider-managed lifecycle choreography worth knowing (all inside the
	// provider, no module-side steps): create waits for the pairing to
	// reach Succeeded; changing the partner breaks the existing pairing,
	// waits, then re-pairs; destroy breaks the pairing, deletes the config,
	// and polls until the alias NAME is released so an immediate re-create
	// with the same alias does not collide. Failover itself is an
	// operational action taken on the SECONDARY during an incident -- never
	// a config change here.
	pairingArgs := &servicebus.NamespaceDisasterRecoveryConfigArgs{
		// ForceNew pair: the alias identity and the primary side are fixed
		// at creation.
		Name:               pulumi.String(spec.AliasName),
		PrimaryNamespaceId: pulumi.String(locals.PrimaryNamespaceId),
		PartnerNamespaceId: pulumi.String(locals.PartnerNamespaceId),
	}

	// Unset defaults the alias connection strings to the namespace's
	// built-in root rule; a scoped rule gives least-privilege alias
	// credentials.
	if locals.AliasAuthorizationRuleId != "" {
		pairingArgs.AliasAuthorizationRuleId = pulumi.StringPtr(locals.AliasAuthorizationRuleId)
	}

	createdPairing, err := servicebus.NewNamespaceDisasterRecoveryConfig(ctx,
		spec.AliasName,
		pairingArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Service Bus disaster-recovery config %s", spec.AliasName)
	}

	// Export stack outputs. The alias connection strings are what DR-aware
	// clients hold: they address the alias DNS name, so a failover needs no
	// client reconfiguration.
	ctx.Export(OpDisasterRecoveryConfigId, createdPairing.ID())
	ctx.Export(OpAliasName, createdPairing.Name)
	ctx.Export(OpPrimaryConnectionStringAlias, pulumi.ToSecret(createdPairing.PrimaryConnectionStringAlias))
	ctx.Export(OpSecondaryConnectionStringAlias, pulumi.ToSecret(createdPairing.SecondaryConnectionStringAlias))
	ctx.Export(OpDefaultPrimaryKey, pulumi.ToSecret(createdPairing.DefaultPrimaryKey))
	ctx.Export(OpDefaultSecondaryKey, pulumi.ToSecret(createdPairing.DefaultSecondaryKey))

	return nil
}
