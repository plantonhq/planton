# AzureAvailabilitySet

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureAvailabilitySetSpec** defines an availability set -- the
classic pre-zones placement grouping that spreads VMs across
separate fault domains (power/network/rack) and update domains
(planned-maintenance batches) so one hardware failure or
maintenance window cannot take them all down. VMs join the set at
creation (AzureVirtualMachine's availability.availability_set_id).

An availability set is free and its whole configuration is fixed at
creation (only tags update in place). Prefer availability ZONES in
zoned regions -- the set remains the right tool in regions without
zones and for classic lift-and-shift topologies.

## Example

```yaml
# Deep-shape example for docs and offline validation: an availability
# set with explicit domain counts, managed alignment, and a proximity
# placement group. References are literal values so the manifest
# validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureAvailabilitySet
metadata:
  name: test-availability-set
  id: test-availability-set
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  name: web-avset
  region: eastus
  platformUpdateDomainCount: 5
  platformFaultDomainCount: 3
  managed: true
  proximityPlacementGroupId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Compute/proximityPlacementGroups/web-ppg
  tags:
    tier: web
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.platformUpdateDomainCount` | `int32` |  |  |  |
| `spec.platformFaultDomainCount` | `int32` |  |  |  |
| `spec.managed` | `bool` |  |  |  |
| `spec.proximityPlacementGroupId` | `string \| valueFrom` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the availability set lives in. Can be a
literal string or a reference to an AzureResourceGroup output.
VMs joining the set must live in the same resource group and
region.

**ForceNew**: changing this destroys and recreates the set.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The availability set's name -- up to 80 characters of letters,
numbers, dots, dashes, and underscores; starts with a letter or
number and ends with a letter, number, or underscore.

**ForceNew**: changing this destroys and recreates the set.

- rule: Availability set names are up to 80 letters, numbers, dots, dashes, and underscores, starting with a letter or number and ending with a letter, number, or underscore
- rule: {"required":true}

### spec.region

`string` · required

The Azure region the set is created in, e.g. "eastus". VMs
joining the set must be in the same region.

**ForceNew**: changing this destroys and recreates the set.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.platformUpdateDomainCount

`int32` · optional (explicit presence)

How many update domains (planned-maintenance batches) the set
spreads VMs across, 1-20. Unset means the provider default, 5.
During planned maintenance Azure reboots one update domain at a
time.

**ForceNew**: changing this destroys and recreates the set.

- rule: {"int32":{"lte":20,"gte":1}}

### spec.platformFaultDomainCount

`int32` · optional (explicit presence)

How many fault domains (independent power/network/rack groups)
the set spreads VMs across, 1-3. Unset means the provider
default, 3. Some regions support fewer than 3 -- Azure rejects a
count the region cannot provide.

**ForceNew**: changing this destroys and recreates the set.

- rule: {"int32":{"lte":3,"gte":1}}

### spec.managed

`bool` · optional (explicit presence)

Whether the set is MANAGED (aligned with managed disks: fault
domains also isolate the VMs' disk storage). Unset means the
provider default, true -- the right value for every managed-disk
VM, i.e. essentially all of them.

**ForceNew**: changing this destroys and recreates the set.

### spec.proximityPlacementGroupId

`string | valueFrom`

Co-locates the set's VMs with a proximity placement group for
minimal inter-VM latency. Plain ARM ID: proximity placement
groups are not modeled as a Planton kind.

**ForceNew**: changing this destroys and recreates the set.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Tags to apply to the availability set, merged over the
Planton-derived metadata tags (user values win on key conflicts).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureAvailabilitySet, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.availability_set_id` | `string` | The availability set's Azure Resource Manager ID -- what VMs reference to join the set. |
| `status.outputs.availability_set_name` | `string` | The availability set's name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureVirtualMachine | `spec.availability.availabilitySetId` | `status.outputs.availability_set_id` |

## See Also

- [Overview](../README.md)
