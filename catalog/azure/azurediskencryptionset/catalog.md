# Azure Disk Encryption Set

Deploys an Azure Disk Encryption Set -- the bridge between Key Vault and server-side disk encryption. Managed disks, VM OS disks, and scale-set disks reference the SET (never the key directly), and the set holds the grantable identity that unwraps the vault key. One set serves many disks: the fan-out point where a single customer-managed key protects a whole environment's storage.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Disk Encryption Set** -- in the specified region and resource group, bound to the referenced Key Vault key with the chosen encryption posture
- **Managed identity** -- required: a system-assigned principal (the grant target surfaced in the outputs) or attached user-assigned identities that unwrap the key
- **Automatic key rotation** -- when enabled: the set re-wraps every dependent disk against each new key version within about a day of rotation
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically, merged with any user tags (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Key Vault with PURGE PROTECTION enabled** -- Azure refuses disk encryption against a vault whose keys could be permanently purged.
- **An AzureKeyVaultKey** with Wrap Key + Unwrap Key in its permitted operations. Reference its `versionless_id` (rotation-following) or `key_id` (pinned).
- **For user-assigned identities** -- AzureUserAssignedIdentity resources already holding vault crypto access ('Key Vault Crypto Service Encryption User' or an access-policy grant).

## Deploy

### Console

Open the deployment store, find **Azure Disk Encryption Set**, and click **Deploy**. The creation wizard walks you through placement (region affinity taught up front), the vault key with the rotation pairing rule carried live to the moment of choice, the required identity, and the fixed encryption posture. Start from the **System-Assigned Set with Auto Key Rotation** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDiskEncryptionSet
metadata:
  name: prod-des
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-security-rg"
  name: prod-des
  keyVaultKeyId:
    valueFrom:
      kind: AzureKeyVaultKey
      name: disk-cmk
      fieldPath: status.outputs.versionless_id
  autoKeyRotationEnabled: true
  identity:
    type: SYSTEM_ASSIGNED
```

```shell
planton apply -f disk-encryption-set.yaml
```

This creates the recommended posture: a system-assigned identity and automatic rotation following the key's versionless ID. Grant the identity vault crypto access after creation (via the `identity_principal_id` output), then point disks at the set. An Infra Job tracks the provisioning in real time.

### InfraChart

Compute kinds wire to the set through its ARM ID in the same InfraPipeline:

```yaml
spec:
  diskEncryptionSetId:
    valueFrom:
      kind: AzureDiskEncryptionSet
      name: prod-des
      fieldPath: status.outputs.disk_encryption_set_id
```

The InfraPipeline resolves the dependency graph -- vault, key, set, then the disks and VMs the set protects.

## Key Configuration

These are the most important decisions when configuring a disk encryption set. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The rotation pairing** -- Azure's one apply-time trap with no authoring-time net: `autoKeyRotationEnabled: true` requires the key's VERSIONLESS ID (what an AzureKeyVaultKey reference resolves by default); rotation off pins a VERSIONED key URL. The mismatch fails at apply.

**Identity ordering** -- SYSTEM_ASSIGNED (recommended) is born with the set, so the vault grant comes AFTER creation via the `identity_principal_id` output. USER_ASSIGNED identities exist before the set, so their grants can too -- the ordering that lets a whole environment stamp out in one pipeline run. Switching a user-assigned flavor to system-assigned replaces the set.

**Encryption type** -- fixed at creation, one set per posture: the customer-key default, double encryption (platform + customer keys, for the strictest regimes), or confidential-VM disks (that posture serves ONLY confidential VMs' OS disks).

**Region affinity** -- disks only reference sets in their OWN region; a multi-region footprint runs one set per region, all sharing the same vault key.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureKeyVaultKey** | `keyVaultKeyId` | `status.outputs.versionless_id` |
| **AzureUserAssignedIdentity** | `identity.identityIds[]` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `disk_encryption_set_id` | Azure Resource Manager ID of the set | The `disk_encryption_set_id` on AzureManagedDisk, AzureVirtualMachine disks, and scale-set disks |
| `disk_encryption_set_name` | Name of the set | Operator tooling |
| `identity_principal_id` | The system-assigned identity's principal ID (empty for user-assigned-only) | The AzureRoleAssignment grant target for vault crypto access |
| `identity_tenant_id` | The Entra tenant of the system-assigned identity | Grant tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**System-assigned rotation** -- the recommended posture: the set's own identity, automatic rotation following the key's versionless ID, and the vault grant issued after creation via `identity_principal_id`. Start from the **System-Assigned Set with Auto Key Rotation** preset.

**User-assigned pre-provisioned** -- attached identities whose vault grants exist before the set deploys, avoiding the system-assigned chicken-and-egg; the cleaner ordering when the identity, the grant, and the set all deploy in one dependency-ordered pass. Start from the **User-Assigned Set with Pre-Provisioned Access** preset.

## Works With

- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- hosts the key; must run purge protection
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the customer-managed key this set unwraps, referenced by `keyVaultKeyId` (its `versionless_id` for the rotating posture)
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- pre-provisioned unwrapping identities attached so vault grants can exist BEFORE the set deploys
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- grants the set's identity vault crypto access
- [**Azure Managed Disk**](/infra-catalog/azure-managed-disk) / [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) / [**Azure Virtual Machine Scale Set**](/infra-catalog/azure-virtual-machine-scale-set) -- reference the set's ARM ID as their `disk_encryption_set_id` to encrypt disks through it
