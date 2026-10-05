# AWS Transit Gateway Route Table

A Transit Gateway route table is one isolated routing domain inside a Transit Gateway hub. It owns its domain's membership and routes:

- **Associations** -- which attachments USE this table for their outbound lookups (at most one association per attachment, gateway-wide). An entry can take over an attachment's existing association in the same apply (`replaceExistingAssociation`).
- **Propagations** -- which attachments ADVERTISE their routes into this table (any number of tables per attachment).
- **Static routes** -- a destination CIDR forwarded through one attachment, or blackholed. Statics win longest-prefix ties over propagated routes.
- **Prefix list references** -- route a managed prefix list's whole CIDR set via one attachment (or blackhole it), tracking list membership automatically.
- **Default-table designation** -- optionally claim this table as the GATEWAY's default association and/or propagation table, so attachments that keep their default-table dials on land here.

The canonical segmented topology -- production spokes that reach shared services but not each other, an inspection VPC that hair-pins inter-spoke traffic -- is built from exactly this resource plus attachments created with their default-table membership turned off.

## When to Use

- Any topology beyond the default full mesh: prod/non-prod isolation, inspection hair-pinning, egress-VPC default routes, blackholed quarantine CIDRs.
- Keep the gateway's `defaultRouteTableAssociation`/`defaultRouteTablePropagation` disabled (or pin the relevant attachments' membership off) so custom tables own the routing -- OR, on a default-enabled gateway, designate one custom table as the default domain (`setAsDefaultAssociationTable`/`setAsDefaultPropagationTable`) and take over already-associated attachments with `replaceExistingAssociation`.

## When NOT to Use

- A pure full-mesh topology needs no custom tables -- the gateway's default route table handles everything.

## Prerequisites

- An `AwsTransitGateway`.
- `AwsTransitGatewayVpcAttachment` resources (or literal attachment IDs for VPN / Direct Connect / peering attachments created outside the Planton graph).

## Spec Fields

| Field | Type | Default | Description |
|---|---|---|---|
| `region` | string | (required) | AWS region; must match the gateway |
| `transitGatewayId` | ref | (required) | The gateway this table belongs to (create-time immutable) |
| `associations` | object[] | [] | Attachments associated with this table: `attachmentId` (unique) + optional `replaceExistingAssociation` to take over an existing association (default table or another custom table) in one apply |
| `propagations` | ref[] | [] | Attachments advertising their routes into this table |
| `routes` | object[] | [] | Static routes: `destinationCidrBlock` (unique) + exactly one of `attachmentId` / `blackhole` |
| `prefixListReferences` | object[] | [] | Prefix list routes: `prefixListId` (unique) + exactly one of `attachmentId` / `blackhole` |
| `setAsDefaultAssociationTable` | bool | false | Designate this table as the gateway's default ASSOCIATION table (default-enabled gateways; one claimant per gateway; removal restores the original) |
| `setAsDefaultPropagationTable` | bool | false | Designate this table as the gateway's default PROPAGATION table (same contract) |

## Outputs

| Output | Description |
|---|---|
| `route_table_id` | The table ID (tgw-rtb-...) |
| `route_table_arn` | ARN for IAM policies |

## Composition

- References an `AwsTransitGateway` and any number of `AwsTransitGatewayVpcAttachment`s.
- Association/propagation/route entries accept literal attachment IDs, so VPN, Direct Connect, and peering attachments created outside the graph participate in the domain too.

## Example: isolated production domain

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsTransitGatewayRouteTable
metadata:
  name: prod-domain
spec:
  region: us-east-1
  transitGatewayId:
    valueFrom:
      kind: AwsTransitGateway
      name: segmented-hub
      fieldPath: status.outputs.transit_gateway_id
  associations:
    - attachmentId:
        valueFrom:
          kind: AwsTransitGatewayVpcAttachment
          name: prod-vpc-attachment
          fieldPath: status.outputs.attachment_id
  propagations:
    - valueFrom:
        kind: AwsTransitGatewayVpcAttachment
        name: shared-services-attachment
        fieldPath: status.outputs.attachment_id
  routes:
    # Everything else leaves through the egress VPC.
    - destinationCidrBlock: 0.0.0.0/0
      attachmentId:
        valueFrom:
          kind: AwsTransitGatewayVpcAttachment
          name: egress-vpc-attachment
          fieldPath: status.outputs.attachment_id
    # The non-prod CIDR can never be reached from this domain.
    - destinationCidrBlock: 10.200.0.0/16
      blackhole: true
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
