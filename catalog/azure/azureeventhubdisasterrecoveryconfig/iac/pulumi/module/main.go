package module

import (
	"github.com/pkg/errors"
	azureeventhubdisasterrecoveryconfigv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubdisasterrecoveryconfig/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventhub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureeventhubdisasterrecoveryconfigv1alpha1.AzureEventHubDisasterRecoveryConfigIacInput) error {
	locals, err := initializeLocals(ctx, iacInput)
	if err != nil {
		return errors.Wrap(err, "failed to initialize locals")
	}

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventHubDisasterRecoveryConfig.Spec

	// The geo-DR pairing: metadata (hubs, consumer groups, authorization
	// rules -- not event data) continuously replicates primary -> partner,
	// and the alias DNS name fronts whichever namespace is currently
	// primary.
	//
	// Provider-managed lifecycle choreography worth knowing (all inside the
	// provider, no module-side steps): create waits for the pairing to
	// reach the Succeeded provisioning state (polling Accepted ->
	// Succeeded); changing the partner BREAKS the existing pairing first,
	// waits, then re-pairs to the new partner; destroy breaks the pairing,
	// deletes the config, then waits BOTH for the config to 404 AND for the
	// alias NAME to be released by Azure's name-availability check -- the
	// alias name stays reserved briefly after deletion, so destroys take
	// minutes by the service's own design. Failover itself is an
	// operational action performed from the SECONDARY side
	// (portal/CLI/SDK), never a config change on this resource.
	createdPairing, err := eventhub.NewEventhubNamespaceDisasterRecoveryConfig(ctx,
		spec.AliasName,
		&eventhub.EventhubNamespaceDisasterRecoveryConfigArgs{
			// ForceNew trio: the alias identity and the primary side
			// (addressed by discrete names parsed from the primary's ARM
			// ID) are fixed at creation.
			Name:              pulumi.String(spec.AliasName),
			NamespaceName:     pulumi.String(locals.PrimaryNamespaceName),
			ResourceGroupName: pulumi.String(locals.PrimaryResourceGroupName),

			PartnerNamespaceId: pulumi.String(locals.PartnerNamespaceId),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Event Hubs disaster-recovery config %s", spec.AliasName)
	}

	// Export outputs. No credential outputs here: Azure's Event Hubs
	// DR resource exposes none. Alias-addressed connection strings surface
	// on the namespace and authorization-rule kinds instead.
	ctx.Export(OpDisasterRecoveryConfigId, createdPairing.ID())
	ctx.Export(OpAliasName, createdPairing.Name)

	return nil
}
