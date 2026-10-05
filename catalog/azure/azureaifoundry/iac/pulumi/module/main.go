package module

import (
	"github.com/pkg/errors"
	azureaifoundryv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureaifoundry/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/aifoundry"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureaifoundryv1alpha1.AzureAiFoundryIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient
	// chain). The machine_learning features flag makes destroy purge the soft-delete ghost
	// that would otherwise keep holding the hub NAME (hubs are ML workspaces at ARM and the
	// provider default leaves the ghost); a soft-delete recovery window is not part of this
	// module's contract -- mirrors the Terraform module's provider features block.
	azureProvider, err := pulumiazureprovider.GetWithFeatures(ctx, iacInput.ProviderConfig,
		azure.ProviderFeaturesArgs{
			MachineLearning: azure.ProviderFeaturesMachineLearningArgs{
				PurgeSoftDeletedWorkspaceOnDestroy: pulumi.Bool(true),
			},
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureAiFoundry.Spec

	// Create the AI Foundry hub -- ARM-wise an ML workspace of kind
	// "Hub" (the SDK's aifoundry.Hub token creates exactly that
	// object): the shared foundation whose security/storage/network
	// posture every project inside it inherits. The key vault and
	// storage attachments are ForceNew; insights and registry update
	// in place. Deletion is a SOFT delete: the ghost keeps holding the
	// hub name until purged (the provider's machine_learning features
	// flag).
	hubArgs := &aifoundry.HubArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// The two required companion services (both ForceNew).
		KeyVaultId:       pulumi.String(spec.KeyVaultId.GetValue()),
		StorageAccountId: pulumi.String(spec.StorageAccountId.GetValue()),
		Tags:             pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.Identity != nil {
		identityArgs := &aifoundry.HubIdentityArgs{
			Type: pulumi.String(identityTypeWire[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		hubArgs.Identity = identityArgs
	}

	// Attachable and re-pointable in place (unlike the classic ML
	// workspace, where the registry attachment is ForceNew).
	if spec.ApplicationInsightsId.GetValue() != "" {
		hubArgs.ApplicationInsightsId = pulumi.String(spec.ApplicationInsightsId.GetValue())
	}
	if spec.ContainerRegistryId.GetValue() != "" {
		hubArgs.ContainerRegistryId = pulumi.String(spec.ContainerRegistryId.GetValue())
	}

	if spec.PrimaryUserAssignedIdentity.GetValue() != "" {
		hubArgs.PrimaryUserAssignedIdentity = pulumi.String(spec.PrimaryUserAssignedIdentity.GetValue())
	}

	// Optional-with-default-true in the spec; the provider's own
	// default is "Enabled". Presence-guard with the proto default so
	// manifest-driven IaC inputs (nil optional) send the same wire
	// value the Terraform module's optional(bool, true) carries.
	publicNetworkAccess := "Enabled"
	if spec.PublicNetworkAccessEnabled != nil && !*spec.PublicNetworkAccessEnabled {
		publicNetworkAccess = "Disabled"
	}
	hubArgs.PublicNetworkAccess = pulumi.String(publicNetworkAccess)

	// Customer-managed-key encryption; the whole block is ForceNew.
	// The key id is a VERSIONED Key Vault key URL -- the provider's
	// hub contract (versionless is rejected; rotation does not
	// auto-propagate, unlike the classic ML workspace).
	if spec.Encryption != nil {
		encryptionArgs := &aifoundry.HubEncryptionArgs{
			KeyVaultId: pulumi.String(spec.Encryption.KeyVaultId.GetValue()),
			KeyId:      pulumi.String(spec.Encryption.KeyId.GetValue()),
		}
		if spec.Encryption.UserAssignedIdentityId.GetValue() != "" {
			encryptionArgs.UserAssignedIdentityId = pulumi.String(spec.Encryption.UserAssignedIdentityId.GetValue())
		}
		hubArgs.Encryption = encryptionArgs
	}

	// The managed virtual network. isolation_mode is Optional+Computed
	// on the provider -- unspecified omits it and the value is read
	// back.
	if spec.ManagedNetwork != nil {
		managedNetworkArgs := &aifoundry.HubManagedNetworkArgs{}
		if isolationMode, ok := isolationModeWire[spec.ManagedNetwork.IsolationMode]; ok {
			managedNetworkArgs.IsolationMode = pulumi.String(isolationMode)
		}
		hubArgs.ManagedNetwork = managedNetworkArgs
	}

	// Sent only when true (both engines): the property is
	// Optional+Computed and the SERVICE flips it true when encryption
	// is enabled -- a pinned false would fight that read-back. ForceNew.
	if spec.HighBusinessImpactEnabled {
		hubArgs.HighBusinessImpactEnabled = pulumi.Bool(true)
	}

	if spec.Description != "" {
		hubArgs.Description = pulumi.String(spec.Description)
	}
	if spec.FriendlyName != "" {
		hubArgs.FriendlyName = pulumi.String(spec.FriendlyName)
	}

	createdHub, err := aifoundry.NewHub(ctx,
		spec.Name,
		hubArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create ai foundry hub %s", spec.Name)
	}

	ctx.Export(OpAiFoundryId, createdHub.ID())
	ctx.Export(OpAiFoundryName, createdHub.Name)
	ctx.Export(OpWorkspaceGuid, createdHub.WorkspaceId)
	ctx.Export(OpDiscoveryUrl, createdHub.DiscoveryUrl)
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdHub.Identity.PrincipalId())

	return nil
}
