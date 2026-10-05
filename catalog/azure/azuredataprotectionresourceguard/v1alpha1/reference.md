# AzureDataProtectionResourceGuard

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureDataProtectionResourceGuardSpec** defines a Data Protection
Resource Guard (ARM: Microsoft.DataProtection/resourceGuards) --
the approval gate behind Multi-User Authorization (MUA): once a
backup vault references a guard, privileged vault operations
(disabling soft delete, deleting protection, reducing retention)
require an approval THROUGH the guard before they execute.

**The guard's protection comes from separation of scope**: place it
in a resource group (or subscription) a DIFFERENT administrator
controls than the vaults it guards. An attacker who compromises the
vault's scope then still cannot approve their own destructive
operations. A guard in the same scope as its vaults is a speed
bump, not a control.

**One guard serves many vaults** -- vaults reference the guard by
its ARM ID (the classic Recovery Services vault's resource_guard_id
field). The guard itself is a free configuration object.

## Example

```yaml
# Offline-plan test manifest. Exercises the full surface: the
# exclusion list (present here to prove the wire shape -- production
# guidance is an EMPTY list, which guards everything; entries must be
# operations ARM allows excluding, never the MUA-mandatory set) and
# user tags merged over the derived ones.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataProtectionResourceGuard
metadata:
  name: test-data-protection-resource-guard
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: test-backup-mua-guard
  vaultCriticalOperationExclusionList:
    - Microsoft.RecoveryServices/vaults/backupResourceGuardProxies/write
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.vaultCriticalOperationExclusionList` | `[]string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the guard lives in, e.g. "eastus". Changing the
region replaces the guard.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the guard is created in -- ideally one a
DIFFERENT administrator controls than the vaults it guards (that
separation IS the security model). Can be a literal
resource-group name or a reference to an AzureResourceGroup's
name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The guard's name, unique within its resource group: 1-260
characters (the provider's own rule). Changing the name replaces
the guard.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"260"}}

### spec.vaultCriticalOperationExclusionList

`[]string`

Critical vault operations EXCLUDED from the guard's approval
requirement, as ARM operation names (e.g.
"Microsoft.RecoveryServices/vaults/backupResourceGuardProxies/write").
An excluded operation executes without an approval. Leave empty to
guard every critical operation (the strongest posture). Updates in
place. Azure keeps a MANDATORY set no guard may ever exclude and
rejects the create outright
(BMSUserErrorInvalidCriticalOperationExclusionList) when the list
names one: the operations that would disarm the guard itself
(backupconfig/write, both vault families'
backupResourceGuardProxies/delete,
backupVaults/write#reduceSoftDeleteSecurity) and the cross-tenant
vault-mapping operations. The provider does not validate this --
the rule below mirrors ARM's own rejection list so the failure
lands at admission, not at deploy.

- rule: the exclusion list names an operation Azure never allows excluding -- the MUA-mandatory set (backupconfig/write, backupResourceGuardProxies/delete, backupVaults/write#reduceSoftDeleteSecurity, and the backupCrossTenantVaultMappings operations) is always guarded
- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.tags

`map<string, string>`

Free-form tags applied to the guard, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDataProtectionResourceGuard, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.resource_guard_id` | `string` | The Azure Resource Manager ID of the guard -- what backup vaults reference to put themselves under the guard's Multi-User Authorization. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.DataProtection/resourceGuards/{name} |
| `status.outputs.resource_guard_name` | `string` | The guard's name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureRecoveryServicesVault | `spec.resourceGuardId` | `status.outputs.resource_guard_id` |

## See Also

- [Overview](../README.md)
