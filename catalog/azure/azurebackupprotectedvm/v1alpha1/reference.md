# AzureBackupProtectedVm

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureBackupProtectedVmSpec** defines a protected VM registration
(ARM: Microsoft.RecoveryServices/vaults/{vault}/.../protectedItems/
VM;iaasvmcontainerv2;{vm-rg};{vm-name}) -- the binding that puts one
virtual machine under a backup policy's protection. Creating it only
REGISTERS protection: the first backup runs on the policy's
schedule, not immediately.

**Destroying this resource stops protection AND deletes the backup
data** (the modules keep the engines' default destroy behavior).
The provider offers retain-on-destroy switches at the engine level
(`vm_backup_stop_protection_and_retain_data_on_destroy` /
`vm_backup_suspend_protection_and_retain_data_on_destroy` in the
azurerm features block) for teams that need the data to outlive the
binding.

The spec requires the source VM -- the provider technically allows
clearing it on update (retaining data without a VM link), but that
flow is protection_state's job here (a recorded tightening).

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: the four
# resolved references (vault resource group, vault name, source VM,
# policy), a disk-LUN exclusion list, and an explicit Protected
# posture.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureBackupProtectedVm
metadata:
  name: test-backup-protected-vm
  org: test-org
  env: dev
spec:
  resourceGroup:
    value: test-rg
  recoveryVaultName:
    value: test-backup-vault
  sourceVmId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/vm-rg/providers/Microsoft.Compute/virtualMachines/test-app-vm
  backupPolicyId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.RecoveryServices/vaults/test-backup-vault/backupPolicies/hourly-enhanced-policy
  excludeDiskLuns: [2, 3]
  protectionState: Protected
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.recoveryVaultName` | `string \| valueFrom` | yes |  | AzureRecoveryServicesVault (`status.outputs.recovery_services_vault_name`) |
| `spec.sourceVmId` | `string \| valueFrom` | yes |  | AzureVirtualMachine (`status.outputs.vm_id`) |
| `spec.backupPolicyId` | `string \| valueFrom` | yes |  | AzureBackupPolicyVm (`status.outputs.backup_policy_id`) |
| `spec.excludeDiskLuns` | `[]int32` |  |  |  |
| `spec.includeDiskLuns` | `[]int32` |  |  |  |
| `spec.protectionState` | `string` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the protecting VAULT lives in (NOT
necessarily the VM's group). Can be a literal resource-group name
or a reference to an AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.recoveryVaultName

`string | valueFrom` · required

The Recovery Services vault that protects the VM, by NAME (ARM
addresses protected items as children of a vault). Fixed at
creation.

- references: AzureRecoveryServicesVault (`status.outputs.recovery_services_vault_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureRecoveryServicesVault, name: <that resource's name>, fieldPath: status.outputs.recovery_services_vault_name}} -- a bare string does not parse

### spec.sourceVmId

`string | valueFrom` · required

The virtual machine to protect, by ARM ID. The VM must live in
the vault's region. Changing it replaces the protection (a new
protected item; the old VM's backup data follows the vault's
soft-delete rules).

- references: AzureVirtualMachine (`status.outputs.vm_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualMachine, name: <that resource's name>, fieldPath: status.outputs.vm_id}} -- a bare string does not parse

### spec.backupPolicyId

`string | valueFrom` · required

The VM backup policy that governs schedule and retention, by ARM
ID. Re-pointing to a different policy updates in place.

- references: AzureBackupPolicyVm (`status.outputs.backup_policy_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureBackupPolicyVm, name: <that resource's name>, fieldPath: status.outputs.backup_policy_id}} -- a bare string does not parse

### spec.excludeDiskLuns

`[]int32`

Disk LUNs (logical unit numbers) EXCLUDED from backup -- back up
the whole VM except these data disks. Mutually exclusive with
include_disk_luns.

- rule: {"repeated":{"items":{"int32":{"gte":0}}}}

### spec.includeDiskLuns

`[]int32`

Disk LUNs INCLUDED in backup -- back up ONLY the OS disk and
these data disks. Mutually exclusive with exclude_disk_luns.

- rule: {"repeated":{"items":{"int32":{"gte":0}}}}

### spec.protectionState

`string`

The desired protection posture (the wire values). Unset lets
Azure manage it (normally "Protected"; Azure also reports
transient states like IRPending which the modules treat as
Protected). ProtectionStopped halts backups and retains data;
BackupsSuspended does the same but ONLY works while the vault is
immutable (Unlocked/Locked) -- an apply-time contract Azure
enforces against the live vault.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Protected","BackupsSuspended","ProtectionStopped"]}}

## Validation Rules

- `bprv_disk_luns_exclusive`: exclude_disk_luns and include_disk_luns are mutually exclusive -- pick the disks to skip OR the disks to keep, not both

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureBackupProtectedVm, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.backup_protected_vm_id` | `string` | The Azure Resource Manager ID of the protected item. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.RecoveryServices/vaults/{vault}/backupFabrics/Azure/protectionContainers/iaasvmcontainer;iaasvmcontainerv2;{vm-rg};{vm-name}/protectedItems/VM;iaasvmcontainerv2;{vm-rg};{vm-name} |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.recoveryVaultName` | AzureRecoveryServicesVault | `status.outputs.recovery_services_vault_name` |
| `spec.sourceVmId` | AzureVirtualMachine | `status.outputs.vm_id` |
| `spec.backupPolicyId` | AzureBackupPolicyVm | `status.outputs.backup_policy_id` |

## See Also

- [Overview](../README.md)
