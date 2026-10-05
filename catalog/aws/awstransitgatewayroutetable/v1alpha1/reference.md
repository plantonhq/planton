# AwsTransitGatewayRouteTable

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsTransitGatewayRouteTableSpec defines the desired configuration for an
AWS Transit Gateway route table -- an isolated routing domain inside a
Transit Gateway hub.

A route table is deliberately its own resource rather than a field on the
gateway: a gateway carries many route tables (one per routing domain), and
the table is the unit that segmentation policy hangs off. The canonical
segmented topology -- production spokes that reach shared services but not
each other, an inspection VPC that hair-pins all inter-spoke traffic --
is built from exactly this resource plus attachments created with their
default-table membership turned off.

The table OWNS its routing domain's membership and routes, folded here
because none of them exists independently of the table:
- associations: which attachments USE this table for their outbound
  lookups. An attachment is associated with at most ONE table -- AWS
  rejects a second association at apply time, and since documents cannot
  see each other, that uniqueness cannot be validated earlier. Each entry
  can take over an attachment's existing association in one apply
  (replace_existing_association).
- propagations: which attachments ADVERTISE their routes into this table
  (an attachment can propagate to many tables).
- routes: static routes -- a destination CIDR sent to one attachment, or
  blackholed to drop traffic (the segmentation kill switch).
- prefix_list_references: route a managed prefix list's whole CIDR set
  via one attachment (or blackhole it) instead of maintaining per-CIDR
  static routes.
- set_as_default_association_table / set_as_default_propagation_table:
  designate THIS table as the gateway's default association/propagation
  table, redirecting where default-enabled attachments land.

Membership entries reference AwsTransitGatewayVpcAttachment resources by
default, and accept literal attachment IDs for attachments created
outside this resource graph -- VPN connections, Direct Connect gateways,
and peering attachments all surface as Transit Gateway attachments that a
routing domain must be able to include.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsTransitGatewayRouteTable
metadata:
  name: test-tgw-route-table
  org: test-org
  env: dev
  id: test-tgw-route-table-dev
spec:
  region: us-west-2
  transitGatewayId:
    value: tgw-0a1b2c3d4e5f00001
  associations:
    - attachmentId:
        value: tgw-attach-0a1b2c3d4e5f00001
      replaceExistingAssociation: true
  propagations:
    - value: tgw-attach-0a1b2c3d4e5f00001
    - value: tgw-attach-0a1b2c3d4e5f00002
  routes:
    - destinationCidrBlock: 0.0.0.0/0
      attachmentId:
        value: tgw-attach-0a1b2c3d4e5f00002
    - destinationCidrBlock: 10.99.0.0/16
      blackhole: true
  prefixListReferences:
    - prefixListId: pl-0a1b2c3d4e5f00001
      attachmentId:
        value: tgw-attach-0a1b2c3d4e5f00001
  setAsDefaultAssociationTable: true
  setAsDefaultPropagationTable: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.transitGatewayId` | `string \| valueFrom` | yes |  | AwsTransitGateway (`status.outputs.transit_gateway_id`) |
| `spec.associations` | `[]AwsTransitGatewayRouteTableAssociation` |  |  |  |
| `spec.associations[].attachmentId` | `string \| valueFrom` | yes |  | AwsTransitGatewayVpcAttachment (`status.outputs.attachment_id`) |
| `spec.associations[].replaceExistingAssociation` | `bool` |  |  |  |
| `spec.propagations` | `[]string \| valueFrom` |  |  | AwsTransitGatewayVpcAttachment (`status.outputs.attachment_id`) |
| `spec.routes` | `[]AwsTransitGatewayRouteTableRoute` |  |  |  |
| `spec.routes[].destinationCidrBlock` | `string` | yes |  |  |
| `spec.routes[].attachmentId` | `string \| valueFrom` |  |  | AwsTransitGatewayVpcAttachment (`status.outputs.attachment_id`) |
| `spec.routes[].blackhole` | `bool` |  |  |  |
| `spec.prefixListReferences` | `[]AwsTransitGatewayRouteTablePrefixListReference` |  |  |  |
| `spec.prefixListReferences[].prefixListId` | `string` | yes |  |  |
| `spec.prefixListReferences[].attachmentId` | `string \| valueFrom` |  |  | AwsTransitGatewayVpcAttachment (`status.outputs.attachment_id`) |
| `spec.prefixListReferences[].blackhole` | `bool` |  |  |  |
| `spec.setAsDefaultAssociationTable` | `bool` |  |  |  |
| `spec.setAsDefaultPropagationTable` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the route table will be created. Must match the
Transit Gateway's region.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.transitGatewayId

`string | valueFrom` · required

The Transit Gateway this route table belongs to. Create-time immutable:
changing it replaces the table (and everything folded in it). Reference
an AwsTransitGateway's transit_gateway_id output or pass a literal ID.

- references: AwsTransitGateway (`status.outputs.transit_gateway_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsTransitGateway, name: <that resource's name>, fieldPath: status.outputs.transit_gateway_id}} -- a bare string does not parse

### spec.associations

`[]AwsTransitGatewayRouteTableAssociation`

Attachments ASSOCIATED with this table: their outbound traffic is looked
up here. An attachment can be associated with at most ONE route table
across the whole gateway (AWS enforces this at apply time), so an
attachment listed here must have its default_route_table_association
turned off, must not appear in any other table's associations -- or must
be taken over explicitly with the entry's replace_existing_association.
Attachment IDs must be unique within the table (both engines key each
association by its attachment ID).

### spec.associations[].attachmentId

`string | valueFrom` · required

The attachment whose outbound traffic is looked up in this table.
Reference an AwsTransitGatewayVpcAttachment's attachment_id output, or
pass a literal attachment ID for VPN / Direct Connect / peering
attachments created outside the resource graph. Unique within the
table: both engines key the association resource by this ID.

- references: AwsTransitGatewayVpcAttachment (`status.outputs.attachment_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsTransitGatewayVpcAttachment, name: <that resource's name>, fieldPath: status.outputs.attachment_id}} -- a bare string does not parse

### spec.associations[].replaceExistingAssociation

`bool`

Take over the attachment's EXISTING association instead of failing on
it. AWS allows one association per attachment gateway-wide, so
associating an attachment that is already associated somewhere else --
the gateway's default table, or another custom table -- is rejected at
apply time unless this flag is set, in which case the engine
disassociates the existing association first and associates here, in
one apply. Its primary use (per the provider's own guidance) is
attachments on a gateway SHARED INTO this account via RAM, where the
attachment's default-table membership is controlled by the sharing
account and cannot simply be turned off. The flag matters only when the
association is CREATED; changing it on an existing association is a
no-op. Leave it false for attachments whose default-table membership is
already off (the segmented-topology norm).

### spec.propagations

`[]string | valueFrom`

Attachments that PROPAGATE their routes into this table: their CIDRs
(VPC CIDRs for VPC attachments, BGP-learned routes for VPN and Direct
Connect) appear here automatically and track changes. An attachment can
propagate to many tables. Reference AwsTransitGatewayVpcAttachment
attachment_id outputs, or pass literal attachment IDs for attachments
created outside the resource graph.

- references: AwsTransitGatewayVpcAttachment (`status.outputs.attachment_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsTransitGatewayVpcAttachment, name: <that resource's name>, fieldPath: status.outputs.attachment_id}} -- a bare string does not parse

### spec.routes

`[]AwsTransitGatewayRouteTableRoute`

Static routes in this table. Statics win over propagated routes for the
same destination -- use them to steer specific prefixes through an
inspection attachment, provide a default route (0.0.0.0/0) toward an
egress VPC, or blackhole traffic that must never cross the hub.
Destination CIDRs must be unique within the table.

- rule: exactly one of attachment_id or blackhole must be set

### spec.routes[].destinationCidrBlock

`string` · required

Destination CIDR block (IPv4 or IPv6), e.g. "10.20.0.0/16" or
"0.0.0.0/0" for a default route. Unique within the table. AWS uses
longest-prefix match across static and propagated routes, with statics
winning ties.

- rule: {"string":{"minLen":"1"}}

### spec.routes[].attachmentId

`string | valueFrom`

The attachment traffic to this destination is forwarded through.
Required unless blackhole is true. Reference an
AwsTransitGatewayVpcAttachment's attachment_id output, or pass a
literal attachment ID (VPN / Direct Connect / peering).

- references: AwsTransitGatewayVpcAttachment (`status.outputs.attachment_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsTransitGatewayVpcAttachment, name: <that resource's name>, fieldPath: status.outputs.attachment_id}} -- a bare string does not parse

### spec.routes[].blackhole

`bool`

Drop traffic to this destination instead of forwarding it. Mutually
exclusive with attachment_id. Blackholes are the segmentation kill
switch: a spoke's CIDR blackholed in another domain's table can never
be reached from that domain, regardless of propagations.

### spec.prefixListReferences

`[]AwsTransitGatewayRouteTablePrefixListReference`

Managed prefix list references. Each entry routes the entire CIDR set
of a customer-managed or AWS-managed prefix list via one attachment (or
blackholes it), tracking the list's membership as it changes --
operationally safer than mirroring a team's CIDR inventory as static
routes. Prefix list IDs must be unique within the table.

- rule: exactly one of attachment_id or blackhole must be set

### spec.prefixListReferences[].prefixListId

`string` · required

The managed prefix list ID (e.g. "pl-0123456789abcdef0"). Unique within
the table. Both customer-managed and AWS-managed prefix lists are
accepted; the list's CIDR entries become routes that track membership
changes automatically.

- rule: {"string":{"minLen":"1"}}

### spec.prefixListReferences[].attachmentId

`string | valueFrom`

The attachment traffic to the prefix list's CIDRs is forwarded through.
Required unless blackhole is true. Reference an
AwsTransitGatewayVpcAttachment's attachment_id output, or pass a
literal attachment ID.

- references: AwsTransitGatewayVpcAttachment (`status.outputs.attachment_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsTransitGatewayVpcAttachment, name: <that resource's name>, fieldPath: status.outputs.attachment_id}} -- a bare string does not parse

### spec.prefixListReferences[].blackhole

`bool`

Drop traffic to the prefix list's CIDRs instead of forwarding it.
Mutually exclusive with attachment_id.

### spec.setAsDefaultAssociationTable

`bool`

Designate this table as the gateway's DEFAULT ASSOCIATION route table:
attachments that keep default_route_table_association enabled (the
gateway-inherited posture) associate HERE instead of the AWS-created
default table. Meaningful only on a gateway whose own
default_route_table_association dial is enabled -- a gateway created
with the dial disabled has no default-association behavior to redirect.
At most one table per gateway may hold this designation; documents
cannot see each other, so AWS state -- not validation -- is the referee
and the most recent apply wins the pointer. Removing the flag RESTORES
the gateway's original default table (recorded at designation time by
both engines' providers).

### spec.setAsDefaultPropagationTable

`bool`

Designate this table as the gateway's DEFAULT PROPAGATION route table:
attachments that keep default_route_table_propagation enabled advertise
their routes HERE instead of the AWS-created default table. Same
contract as set_as_default_association_table: meaningful only on a
gateway with the propagation dial enabled, at most one claimant per
gateway, and removing the flag restores the original table.

## Validation Rules

- `route_destinations_unique`: routes must have unique destination_cidr_block values
- `prefix_list_ids_unique`: prefix_list_references must have unique prefix_list_id values

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsTransitGatewayRouteTable, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.route_table_id` | `string` | The Transit Gateway route table ID (e.g., "tgw-rtb-0123456789abcdef0"). Referenced by tooling that manages routes or inspects the routing domain, and by the gateway-level default-table designation surface. |
| `status.outputs.route_table_arn` | `string` | The Amazon Resource Name (ARN) of the route table. Used for IAM policies and resource-level permissions. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.transitGatewayId` | AwsTransitGateway | `status.outputs.transit_gateway_id` |
| `spec.associations[].attachmentId` | AwsTransitGatewayVpcAttachment | `status.outputs.attachment_id` |
| `spec.propagations` | AwsTransitGatewayVpcAttachment | `status.outputs.attachment_id` |
| `spec.routes[].attachmentId` | AwsTransitGatewayVpcAttachment | `status.outputs.attachment_id` |
| `spec.prefixListReferences[].attachmentId` | AwsTransitGatewayVpcAttachment | `status.outputs.attachment_id` |

## See Also

- [Overview](../README.md)
