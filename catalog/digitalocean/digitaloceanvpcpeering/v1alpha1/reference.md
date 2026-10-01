# DigitalOceanVpcPeering

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanVpcPeeringSpec models the full digitalocean_vpc_peering
resource surface: a private-network peering connection between exactly
two DigitalOcean VPCs.

The provider accepts an unordered set of VPC ids, but the API requires
exactly two -- this spec models them as two named references so any other
cardinality is unrepresentable. The peering is symmetric: which VPC is
vpc_1 and which is vpc_2 does not matter.

## Example

```yaml
# Reference manifest for DigitalOceanVpcPeering -- protovalidate-valid,
# embedded as the reference page's Example block, and the document the
# offline tofu plan renders. One document: the kind's total surface is
# three fields.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanVpcPeering
metadata:
  name: app-to-data-peering
spec:
  peeringName: app-to-data
  # Literal VPC UUIDs; use valueFrom to reference DigitalOceanVpc
  # resources instead.
  vpc_1:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  vpc_2:
    value: ffffffff-1111-2222-3333-444444444444
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.peeringName` | `string` | yes |  |  |
| `spec.vpc1` | `string \| valueFrom` | yes |  | DigitalOceanVpc (`status.outputs.vpc_id`) |
| `spec.vpc2` | `string \| valueFrom` | yes |  | DigitalOceanVpc (`status.outputs.vpc_id`) |

## Field Details

### spec.peeringName

`string` · required

Name of the VPC peering connection. This is the only field that
updates in place; everything else replaces the peering.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vpc1

`string | valueFrom` · required

The first of the two VPCs to peer. Use a literal VPC UUID or a
reference to a DigitalOceanVpc resource. Changing it replaces the
peering.

- references: DigitalOceanVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.vpc2

`string | valueFrom` · required

The second of the two VPCs to peer. Use a literal VPC UUID or a
reference to a DigitalOceanVpc resource. Must differ from vpc_1
(API-enforced). Changing it replaces the peering.

- references: DigitalOceanVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanVpcPeering, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.peering_id` | `string` | UUID of the VPC peering connection (the resource's API identity and its import id). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpc1` | DigitalOceanVpc | `status.outputs.vpc_id` |
| `spec.vpc2` | DigitalOceanVpc | `status.outputs.vpc_id` |

## See Also

- [Overview](../README.md)
