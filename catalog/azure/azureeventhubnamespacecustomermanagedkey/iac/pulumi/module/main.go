package module

import (
	"github.com/pkg/errors"
	azureeventhubnamespacecustomermanagedkeyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubnamespacecustomermanagedkey/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventhub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureeventhubnamespacecustomermanagedkeyv1alpha1.AzureEventHubNamespaceCustomerManagedKeyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventHubNamespaceCustomerManagedKey.Spec

	// The Key Vault keys, resolved to literal data-plane IDs before the
	// module runs. Versionless key IDs make vault-side rotation propagate
	// automatically; pin versioned IDs only when a compliance regime
	// demands immutable key versions.
	keyVaultKeyIds := pulumi.StringArray{}
	for _, keyVaultKeyId := range spec.KeyVaultKeyIds {
		keyVaultKeyIds = append(keyVaultKeyIds, pulumi.String(keyVaultKeyId.GetValue()))
	}

	// Customer-managed-key (BYOK) encryption applied ONTO an existing
	// Event Hubs namespace -- Azure models CMK as a namespace property
	// configured after creation, not a create-time block, and this
	// resource mirrors that grain. The namespace must have single-tenant
	// capacity (a dedicated cluster or PREMIUM) or Azure rejects the
	// encryption patch.
	//
	// ADD-ONLY lifecycle (Azure's own contract): once CMK is enabled it
	// can never be removed -- Azure has no decrypt-back path. The
	// provider's Delete is deliberately a NO-OP: destroying this resource
	// changes NOTHING on the namespace, and returning to Microsoft-managed
	// keys requires replacing the namespace itself.
	cmkArgs := &eventhub.NamespaceCustomerManagedKeyArgs{
		// ForceNew: the configuration is bound to its namespace for life.
		EventhubNamespaceId: pulumi.String(locals.EventhubNamespaceId),
		KeyVaultKeyIds:      keyVaultKeyIds,
	}

	// ForceNew: the second encryption layer is fixed the moment CMK is
	// first configured. Sent only when set so Azure's default (false)
	// applies otherwise.
	if spec.InfrastructureEncryptionEnabled != nil {
		cmkArgs.InfrastructureEncryptionEnabled = pulumi.BoolPtr(spec.GetInfrastructureEncryptionEnabled())
	}

	// Identity contract: a user-assigned identity named here must ALREADY
	// be attached to the parent namespace's identity block, with
	// wrap/unwrap access on the keys' vault -- Azure rejects the patch
	// otherwise. Unset falls back to the namespace's system-assigned
	// identity (grant IT the vault access).
	if spec.UserAssignedIdentityId.GetValue() != "" {
		cmkArgs.UserAssignedIdentityId = pulumi.StringPtr(spec.UserAssignedIdentityId.GetValue())
	}

	createdCustomerManagedKey, err := eventhub.NewNamespaceCustomerManagedKey(ctx,
		locals.AzureEventHubNamespaceCustomerManagedKey.Metadata.Name,
		cmkArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create Event Hubs namespace customer-managed key")
	}

	// The configuration has no ARM object of its own -- the provider's
	// resource id IS the namespace's ARM id.
	ctx.Export(OpCustomerManagedKeyId, createdCustomerManagedKey.ID())

	return nil
}
