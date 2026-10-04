# AzureMonitorDataCollectionRuleAssociation

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureMonitorDataCollectionRuleAssociationSpec** attaches ONE
machine to an Azure Monitor data collection rule or data collection
endpoint. The association is how a machine enters monitoring: the
Azure Monitor Agent on the target discovers its associations,
downloads the referenced rule, and starts collecting -- and removing
the association detaches the machine without touching the rule.

An association is an extension resource living ON the target machine
(its ARM ID is scoped under the target's ID), which is why machines
join and leave monitoring independently: one rule serves any number
of machines, and one machine can carry many associations (several
rules, plus at most one endpoint association).

**Exactly one binding**: an association binds its target to EITHER a
data collection rule OR a data collection endpoint, never both. The
endpoint form exists for machines whose configuration access must go
through a Data Collection Endpoint (private-link scenarios) and
carries the fixed, Azure-mandated name "configurationAccessEndpoint".

## Example

```yaml
# Deep-shape example for docs and offline validation: one VM attached
# to a data collection rule. References are literal ARM ids so the
# manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorDataCollectionRuleAssociation
metadata:
  name: test-dcra
  id: test-dcra
  org: test-org
  env: test
spec:
  targetResourceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Compute/virtualMachines/app-vm
  # Name the association after the rule it binds -- association names
  # are what fleet listings show.
  name: linux-baseline-assoc
  dataCollectionRuleId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Insights/dataCollectionRules/linux-baseline
  description: attaches the app VM to the Linux baseline rule
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.targetResourceId` | `string \| valueFrom` | yes |  |  |
| `spec.name` | `string` |  |  |  |
| `spec.dataCollectionRuleId` | `string \| valueFrom` |  |  | AzureMonitorDataCollectionRule (`status.outputs.data_collection_rule_id`) |
| `spec.dataCollectionEndpointId` | `string \| valueFrom` |  |  |  |
| `spec.description` | `string` |  |  |  |

## Field Details

### spec.targetResourceId

`string | valueFrom` · required

The machine the rule or endpoint is attached to, by ARM resource
ID -- an Azure virtual machine, a VM scale set, or an Azure
Arc-enabled server. There is no default kind because several kinds
can be the target: reference the resource's `*_id` output
explicitly with valueFrom (kind + fieldPath), or pass a literal
ARM ID.

**ForceNew**: changing this destroys and recreates the association.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.name

`string`

The association's name, unique among the target's associations.
REQUIRED when binding a data collection rule. LEAVE UNSET when
binding a data collection endpoint: Azure mandates the fixed name
"configurationAccessEndpoint" for endpoint associations, and both
engines apply it automatically.

**ForceNew**: changing this destroys and recreates the association.

### spec.dataCollectionRuleId

`string | valueFrom`

The data collection rule to attach, by ARM resource ID. Can be a
literal ARM ID or a reference to an AzureMonitorDataCollectionRule
output. Set exactly one of data_collection_rule_id and
data_collection_endpoint_id.

- references: AzureMonitorDataCollectionRule (`status.outputs.data_collection_rule_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMonitorDataCollectionRule, name: <that resource's name>, fieldPath: status.outputs.data_collection_rule_id}} -- a bare string does not parse

### spec.dataCollectionEndpointId

`string | valueFrom`

The ARM ID of the Azure Monitor Data Collection Endpoint (DCE) to
attach for configuration access. Format:
/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Insights/dataCollectionEndpoints/{name}
Provide the literal ARM ID of an endpoint managed outside this
catalog. Set exactly one of data_collection_rule_id and
data_collection_endpoint_id.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.description

`string`

A free-text description of the association, shown in the portal.

## Validation Rules

- `dcra_exactly_one_binding`: set exactly one of data_collection_rule_id and data_collection_endpoint_id
- `dcra_rule_binding_requires_name`: name is required when data_collection_rule_id is set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMonitorDataCollectionRuleAssociation, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.data_collection_rule_association_id` | `string` | The association's ARM resource ID, scoped under the target machine ({target_resource_id}/providers/Microsoft.Insights/dataCollectionRuleAssociations/{name}). |
| `status.outputs.data_collection_rule_association_name` | `string` | The association's name on the target. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.dataCollectionRuleId` | AzureMonitorDataCollectionRule | `status.outputs.data_collection_rule_id` |

## See Also

- [Overview](../README.md)
