# AzureFabricCapacity

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureFabricCapacitySpec** defines a Microsoft Fabric capacity --
the billing and compute anchor of Microsoft Fabric: workspaces
assign themselves to a capacity, and the capacity's F-SKU sets how
much compute every workload on it (lakehouses, warehouses, Power
BI, real-time analytics) shares.

The capacity is azurerm's ENTIRE Fabric surface: workspaces and the
items inside them are managed through Microsoft's dedicated
`fabric` Terraform provider, the Fabric portal, or its APIs -- not
through ARM.

A running capacity bills PER HOUR from the moment it exists, and the
SKU ladder spans three orders of magnitude (F2048 bills a thousand
times F2's hourly rate).
The capacity's SKU scales up and down in place -- start small.

## Example

```yaml
# Deep-shape example for docs and offline validation: an F2 capacity
# with two administrators and tags. References are literal values so
# the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFabricCapacity
metadata:
  name: test-fabric-capacity
  id: test-fabric-capacity
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  name: testorgfabric
  region: eastus
  skuName: F2
  administrationMembers:
    - admin@testorg.example
    - 11111111-2222-3333-4444-555555555555
  tags:
    costCenter: analytics
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.skuName` | `string` | yes |  |  |
| `spec.administrationMembers` | `[]string` | yes |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the capacity lives in. Can be a literal
string or a reference to an AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the capacity.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The capacity's name -- 3-63 characters; lowercase letters and
numbers only; must start with a letter. The name is how Fabric
workspaces see the capacity, so make it recognizable to the teams
assigning workspaces.

**ForceNew**: changing this destroys and recreates the capacity.

- rule: Capacity names must be 3-63 characters of lowercase letters and numbers, starting with a letter
- rule: {"required":true}

### spec.region

`string` · required

The Azure region the capacity is created in, e.g. "eastus".
Fabric is not available in every Azure region -- check the Fabric
region availability list before choosing.

**ForceNew**: changing this destroys and recreates the capacity.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.skuName

`string` · required

The capacity's F-SKU -- how much compute every workload on the
capacity shares, doubling per step: "F2", "F4", "F8", "F16",
"F32", "F64", "F128", "F256", "F512", "F1024", "F2048". Scales up
and down IN PLACE, so start small and grow with real usage. F64
is the first tier that carries Copilot and unlimited Power BI
sharing. (The provider's sku block also carries a tier field with
exactly one legal value, "Fabric" -- deliberately not part of
this spec; both engines send it explicitly.)

- rule: {"required":true,"string":{"in":["F2","F4","F8","F16","F32","F64","F128","F256","F512","F1024","F2048"]}}

### spec.administrationMembers

`[]string` · required

The capacity's administrators: Microsoft Entra user principal
names (e.g. "admin@contoso.com") or service-principal object IDs.
At least one is required -- Azure rejects a capacity created with
no administrator, and a capacity nobody administers is
unmanageable from the Fabric side. Updatable in place. (Azure's
API technically allows clearing the list on update; the catalog
requires at least one entry at all times -- removing every
administrator from a running paid capacity is a lockout, not a
configuration.)

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.tags

`map<string, string>`

Tags to apply to the capacity, merged over the Planton-derived
metadata tags (user values win on key conflicts).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureFabricCapacity, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.fabric_capacity_id` | `string` | The capacity's Azure Resource Manager ID. |
| `status.outputs.fabric_capacity_name` | `string` | The capacity's name -- what Fabric workspaces assign themselves to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## See Also

- [Overview](../README.md)
