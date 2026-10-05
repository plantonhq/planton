# AWS Transit Gateway Route Table

Creates an isolated routing domain inside an AWS Transit Gateway hub. A route table is the unit segmentation policy hangs off: it owns which attachments USE it for outbound lookups (associations), which attachments FEED routes into it (propagations), and its static routes and prefix-list references — including blackholes, the segmentation kill switch. The canonical segmented topology — production spokes that reach shared services but not each other, an inspection VPC that hair-pins inter-spoke traffic — is built from exactly this resource plus attachments created with their default-table membership turned off.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Transit Gateway Route Table** -- the routing domain itself
- **Route Table Associations** -- one per entry in `associations`, binding each attachment's outbound lookups to this table (optionally taking over an existing association in the same apply)
- **Route Table Propagations** -- one per entry in `propagations`, advertising each attachment's routes into this table
- **Static Routes** -- one per entry in `routes`, each forwarding a destination CIDR to an attachment or blackholing it
- **Prefix List References** -- one per entry in `prefixListReferences`, each routing a managed prefix list's whole CIDR set
- **Default-Table Designations** -- when `setAsDefaultAssociationTable` / `setAsDefaultPropagationTable` are set, the gateway's default association/propagation pointers are repointed at this table
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A Transit Gateway** in the same region -- for segmented designs, typically one whose default-association and default-propagation dials are off.
- **Attachments to include** -- `AwsTransitGatewayVpcAttachment` resources (reference their `attachment_id` outputs), or literal `tgw-attach-…` IDs for VPN, Direct Connect, and peering attachments created outside Planton.
- **Association exclusivity** -- an attachment associates with at most ONE route table across the whole gateway. An attachment listed in this table's `associations` must have `defaultRouteTableAssociation` off and must not appear in any other table's associations; AWS rejects a second association at apply time. To move an attachment that is ALREADY associated somewhere else (the gateway's default table, another custom table, or an attachment on a RAM-shared gateway whose default membership you cannot switch off), set the entry's `replaceExistingAssociation` -- the old association is removed and the new one created in one apply.

## Deploy

### Console

Open the deployment store, find **AWS Transit Gateway Route Table**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Isolated Domain** preset in the [Presets](#presets) tab for the segmented-topology starting point.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsTransitGatewayRouteTable
metadata:
  name: prod-domain
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  transitGatewayId:
    value: "tgw-0123456789abcdef0"
  associations:
    - attachmentId:
        value: "tgw-attach-0123456789abcdef0"
  propagations:
    - value: "tgw-attach-0fedcba9876543210"
  routes:
    - destinationCidrBlock: "10.66.0.0/16"
      blackhole: true
```

```shell
planton apply -f route-table.yaml
```

This creates a domain where one attachment routes by this table, one shared-services attachment advertises its routes in, and one CIDR is blackholed so this domain can never reach it. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the table to the gateway and attachments deployed in the same InfraPipeline:

```yaml
spec:
  transitGatewayId:
    valueFrom:
      kind: AwsTransitGateway
      name: network-hub
      fieldPath: status.outputs.transit_gateway_id
  associations:
    - attachmentId:
        valueFrom:
          kind: AwsTransitGatewayVpcAttachment
          name: prod-a-attachment
          fieldPath: status.outputs.attachment_id
  propagations:
    - valueFrom:
        kind: AwsTransitGatewayVpcAttachment
        name: shared-services-attachment
        fieldPath: status.outputs.attachment_id
```

The InfraPipeline resolves the dependency graph, deploys the gateway and attachments first, then provisions the table with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring a route table. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The immutable gateway** -- `transitGatewayId` is create-time immutable: a route table cannot move between gateways. Changing it replaces the table and everything folded into it.

**Associations vs propagations** -- Association picks the rulebook (an attachment's outbound traffic is looked up here; exclusive — one table per attachment). Propagation fills the rulebook (the attachment's CIDRs appear here automatically; additive — many tables at once).

**Static routes and blackholes** -- Statics win over propagated routes for the same destination via longest-prefix match. Each route targets exactly one thing: an attachment, or the blackhole. Destinations (IPv4 or IPv6) must be unique within the table.

**Prefix list references** -- Route a managed prefix list's entire CIDR set through one attachment (or blackhole it), tracking membership changes automatically — safer than mirroring a team's CIDR inventory as statics. Each prefix list appears at most once per table.

**Association takeover** -- `associations[].replaceExistingAssociation` moves an attachment that is already associated elsewhere into this table in one apply. Its primary use is attachments on RAM-shared gateways (the sharing account controls their default membership) and brownfield adoption of flat hubs. The flag matters only when the association is created.

**Default-table designation** -- `setAsDefaultAssociationTable` / `setAsDefaultPropagationTable` make this table the gateway's default association/propagation table: attachments that keep their default-table dials on land here. Meaningful only on a gateway whose corresponding default dial is enabled; at most one table per gateway may claim each designation (AWS state is the referee — the most recent apply wins), and removing a flag restores the gateway's original default table.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsTransitGateway** | `transitGatewayId` | `status.outputs.transit_gateway_id` |
| **AwsTransitGatewayVpcAttachment** | `associations[].attachmentId`, `propagations[]`, `routes[].attachmentId`, `prefixListReferences[].attachmentId` | `status.outputs.attachment_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `route_table_id` | The Transit Gateway route table ID | Tooling that manages routes or inspects the routing domain |
| `route_table_arn` | ARN of the route table | IAM policies, resource-level permissions |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Isolated domain** -- Associate the spokes of one environment and propagate only shared services: spokes reach shared services but never each other. Start from the **Isolated Domain** preset.

**Inspection domain** -- A static 0.0.0.0/0 (or a broad summary route) toward the inspection VPC's attachment overrides propagated spoke-to-spoke routes, hair-pinning all inter-spoke traffic through the firewall. Start from the **Inspection Domain** preset.

**Managed default domain** -- A custom table designated as the gateway's default association and propagation table (`setAsDefaultAssociationTable` / `setAsDefaultPropagationTable`), so attachments that keep their default dials on land in a table you own and can add statics to. Start from the **Managed Default Domain** preset.

**Kill switch** -- Blackhole a CIDR this domain must never reach; the blackhole holds regardless of propagations.

## Works With

- [**AWS Transit Gateway**](/infra-catalog/aws-transit-gateway) -- the hub this routing domain lives inside; provides `transit_gateway_id`
- [**AWS Transit Gateway VPC Attachment**](/infra-catalog/aws-transit-gateway-vpc-attachment) -- the attachments this table associates, propagates, and routes against; provide `attachment_id`
