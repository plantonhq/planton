package module

import (
	"github.com/pkg/errors"
	azurerecoveryservicesvaultv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurerecoveryservicesvault/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/recoveryservices"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurerecoveryservicesvaultv1alpha1.AzureRecoveryServicesVaultIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureRecoveryServicesVault.Spec

	// Create the Recovery Services vault -- the safe that classic Azure
	// Backup data and Site Recovery configuration live in. The vault is
	// free at rest; cost follows the protected items and their backup
	// storage.
	//
	// Destroy semantics kept deliberately at the engines' defaults:
	// deleting the vault FAILS while protected items remain inside it
	// (the purge_protected_items_from_vault_on_destroy feature stays
	// off) -- stop and delete protections first, then the vault.
	// The platform default fills the sku ("Standard"); the nil guard
	// keeps direct module invocations safe. "RS0" is the legacy
	// tier-style spelling ARM also accepts, priced the same.
	sku := "Standard"
	if spec.Sku != nil {
		sku = *spec.Sku
	}

	vaultArgs := &recoveryservices.VaultArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Sku:               pulumi.String(sku),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
		// Plain bool: false is the provider's own default, so passing
		// the zero value through is exact. Enabling is in-place;
		// DISABLING replaces the vault (the provider's one-way
		// ForceNew, recorded on the spec field).
		CrossRegionRestoreEnabled: pulumi.Bool(spec.CrossRegionRestoreEnabled),
		// ForceNew on the provider; false is its default.
		ClassicVmwareReplicationEnabled: pulumi.Bool(spec.ClassicVmwareReplicationEnabled),
	}

	// Optional-with-default on the provider: the platform default fills
	// these, so they always carry a value here.
	if spec.StorageModeType != nil {
		vaultArgs.StorageModeType = pulumi.String(*spec.StorageModeType)
	}
	if spec.PublicNetworkAccessEnabled != nil {
		vaultArgs.PublicNetworkAccessEnabled = pulumi.Bool(*spec.PublicNetworkAccessEnabled)
	}

	// Optional+Computed on the provider: unset leaves the service
	// default (Disabled). Transitions are limited (Locked <- Unlocked
	// <-> Disabled) and Locked is permanent -- recorded on the spec
	// field; the provider stages Locked through Unlocked itself.
	if spec.Immutability != "" {
		vaultArgs.Immutability = pulumi.String(spec.Immutability)
	}

	if spec.Identity != nil {
		identityArgs := &recoveryservices.VaultIdentityArgs{
			Type: pulumi.String(identityTypeWire[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		vaultArgs.Identity = identityArgs
	}

	// Customer-managed-key encryption. Once enabled it can never be
	// disabled, infrastructure_encryption_enabled can never change, and
	// the sku freezes (the provider's own update guards -- recorded on
	// the spec fields).
	if spec.Encryption != nil {
		encryptionArgs := &recoveryservices.VaultEncryptionArgs{
			// Versionless key URIs are accepted (the provider validates
			// VersionTypeAny) and rotate automatically -- the reference's
			// default target.
			KeyId: pulumi.String(spec.Encryption.KeyId.GetValue()),
			// ARM requires an explicit choice; the plain bool always
			// ships on the wire.
			InfrastructureEncryptionEnabled: pulumi.Bool(spec.Encryption.InfrastructureEncryptionEnabled),
		}
		if spec.Encryption.UseSystemAssignedIdentity != nil {
			encryptionArgs.UseSystemAssignedIdentity = pulumi.Bool(*spec.Encryption.UseSystemAssignedIdentity)
		}
		if spec.Encryption.UserAssignedIdentityId.GetValue() != "" {
			encryptionArgs.UserAssignedIdentityId = pulumi.String(spec.Encryption.UserAssignedIdentityId.GetValue())
		}
		vaultArgs.Encryption = encryptionArgs
	}

	// The vault's built-in Azure Monitor alert settings. Every switch
	// defaults ON both provider- and service-side, so an unset block is
	// wire-equivalent to an all-true one.
	if spec.Monitoring != nil {
		monitoringArgs := &recoveryservices.VaultMonitoringArgs{}
		if spec.Monitoring.AlertsForAllJobFailuresEnabled != nil {
			monitoringArgs.AlertsForAllJobFailuresEnabled = pulumi.Bool(*spec.Monitoring.AlertsForAllJobFailuresEnabled)
		}
		if spec.Monitoring.AlertsForCriticalOperationFailuresEnabled != nil {
			monitoringArgs.AlertsForCriticalOperationFailuresEnabled = pulumi.Bool(*spec.Monitoring.AlertsForCriticalOperationFailuresEnabled)
		}
		// PARITY-EXCEPTION: alerts_for_all_failover_issues_enabled,
		// alerts_for_all_replication_issues_enabled and
		// email_notifications_for_site_recovery_enabled are new in
		// azurerm v5 and ABSENT from the classic Pulumi SDK v6.38.0 --
		// this engine cannot express turning them off. An explicit
		// FALSE fails loudly here (the Terraform module honors it); an
		// explicit true is wire-equivalent to the service default and
		// passes.
		if spec.Monitoring.AlertsForAllFailoverIssuesEnabled != nil && !*spec.Monitoring.AlertsForAllFailoverIssuesEnabled {
			return errors.New("monitoring.alerts_for_all_failover_issues_enabled: false is not supported by the pulumi module " +
				"(the classic Pulumi Azure SDK predates this azurerm v5 setting) -- leave it unset or use the terraform module")
		}
		if spec.Monitoring.AlertsForAllReplicationIssuesEnabled != nil && !*spec.Monitoring.AlertsForAllReplicationIssuesEnabled {
			return errors.New("monitoring.alerts_for_all_replication_issues_enabled: false is not supported by the pulumi module " +
				"(the classic Pulumi Azure SDK predates this azurerm v5 setting) -- leave it unset or use the terraform module")
		}
		if spec.Monitoring.EmailNotificationsForSiteRecoveryEnabled != nil && !*spec.Monitoring.EmailNotificationsForSiteRecoveryEnabled {
			return errors.New("monitoring.email_notifications_for_site_recovery_enabled: false is not supported by the pulumi module " +
				"(the classic Pulumi Azure SDK predates this azurerm v5 setting) -- leave it unset or use the terraform module")
		}
		vaultArgs.Monitoring = monitoringArgs
	}

	createdVault, err := recoveryservices.NewVault(ctx,
		spec.Name,
		vaultArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create recovery services vault %s", spec.Name)
	}

	ctx.Export(OpRecoveryServicesVaultId, createdVault.ID())
	ctx.Export(OpRecoveryServicesVaultName, createdVault.Name)
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdVault.Identity.PrincipalId())

	// The composed Resource Guard association (Multi-User
	// Authorization): privileged vault operations then require an
	// approval through the guard. ARM pins the association's own name
	// to the literal "VaultProxy" -- one guard per vault.
	if spec.ResourceGuardId.GetValue() != "" {
		createdAssociation, err := recoveryservices.NewVaultResourceGuardAssociation(ctx,
			spec.Name+"-resource-guard",
			&recoveryservices.VaultResourceGuardAssociationArgs{
				VaultId:         createdVault.ID(),
				ResourceGuardId: pulumi.String(spec.ResourceGuardId.GetValue()),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdVault))
		if err != nil {
			return errors.Wrap(err, "failed to create resource guard association")
		}
		ctx.Export(OpResourceGuardAssociationId, createdAssociation.ID())
	} else {
		// Keep the output present (empty) so both engines' output
		// shapes stay identical whether or not the arm is configured.
		ctx.Export(OpResourceGuardAssociationId, pulumi.String(""))
	}

	return nil
}
