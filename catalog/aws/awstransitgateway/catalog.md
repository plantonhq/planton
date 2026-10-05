# AWS Transit Gateway

Deploys a Transit Gateway on AWS as a regional networking hub that interconnects VPCs, VPN connections, and Direct Connect through a centralized hub-and-spoke topology, replacing complex VPC peering meshes. The gateway itself is deliberately lean: VPCs join it through the separate **AWS Transit Gateway VPC Attachment** resource, and custom routing domains live on the separate **AWS Transit Gateway Route Table** resource. Decide the routing posture up front -- re-enabling a disabled default-route-table dial replaces the gateway and every attachment on it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Transit Gateway** -- a regional networking hub with configurable BGP ASN, default route table association and propagation, DNS support, VPN ECMP support, optional multicast routing, security-group referencing, an optional in-transit encryption posture, and up to five gateway CIDR blocks for Connect and VPN termination addresses
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

Attachments and custom route tables are **not** part of this resource -- they are first-class sibling resources that reference this gateway's ID.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **Non-overlapping CIDR blocks** across all VPCs that will eventually attach. The default full-mesh routing propagates every attached VPC's CIDRs to the default route table, so overlapping CIDRs cause routing conflicts.
- **A routing posture decision** -- the default-route-table association and propagation dials are effectively one-way: disabling either on a running gateway updates it in place, but re-enabling one later replaces the gateway and every attachment. Decide full-mesh vs segmented before production traffic depends on the hub.

## Deploy

### Console

Open the deployment store, find **AWS Transit Gateway**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Full-Mesh Hub** preset in the [Presets](#presets) tab to pre-populate the zero-routing-configuration posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsTransitGateway
metadata:
  name: network-hub
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  description: Regional hub for prod VPCs
  amazonSideAsn: 65000
```

```shell
planton apply -f transit-gateway.yaml
```

This creates a Transit Gateway with the AWS defaults: full-mesh routing (auto-association and auto-propagation), DNS support, and VPN ECMP enabled. Attach VPCs afterward with `AwsTransitGatewayVpcAttachment` resources. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a Transit Gateway. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Route table association and propagation** -- Both `defaultRouteTableAssociation` and `defaultRouteTablePropagation` default to `true`, giving every attachment full-mesh connectivity. Disable both for segmented topologies where `AwsTransitGatewayRouteTable` resources manage routing domains explicitly. These dials are one-way: disabling later updates in place, but re-enabling replaces the gateway and every attachment.

**BGP ASN** -- `amazonSideAsn` is the private ASN used for BGP sessions with VPN and Direct Connect gateways (16-bit range 64512–65534 or 32-bit range 4200000000–4294967294). Default is 64512. Change this only when integrating with on-premises networks that already use 64512 or when connecting multiple Transit Gateways that need distinct ASNs. Immutable after creation.

**Multicast support** -- Create-time immutable: enabling it after creation requires replacing the entire Transit Gateway. Only enable for specific multicast use cases (financial market data, media streaming).

**Encryption posture** -- `encryptionSupport` is a genuine three-position dial: leave it unset and AWS computes the effective value; pin it on to enforce in-transit encryption through the gateway; pin it off to hold the posture explicitly. Pin it only when a compliance requirement names it.

**Gateway CIDR blocks** -- Up to five `transitGatewayCidrBlocks` (/24 or larger for IPv4, /64 or larger for IPv6, never link-local) provide addresses for Transit Gateway Connect peers and VPN tunnel termination.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies -- the gateway is the root of the private-networking family. Attachments, route tables, and Client VPN endpoints consume its outputs.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `transit_gateway_id` | Transit Gateway ID | VPC attachments, route tables, Client VPN endpoints, VPN connections, Direct Connect gateways |
| `transit_gateway_arn` | ARN of the Transit Gateway | IAM policies, AWS RAM sharing, resource-level permissions |
| `owner_id` | AWS account ID that owns the Transit Gateway | Cross-account sharing verification |
| `association_default_route_table_id` | Default association route table ID | Static route creation when the default association dial is enabled |
| `propagation_default_route_table_id` | Default propagation route table ID | Route propagation management and inspection |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Full-mesh hub** -- The zero-routing-configuration starting point: default association and propagation stay enabled, so every attached VPC reaches every other one out of the box. Start from the **Full-Mesh Hub** preset.

**Segmented hub** -- The isolation-first posture: both default-table dials off, so nothing routes until an `AwsTransitGatewayRouteTable` explicitly associates and propagates attachments. Start from the **Segmented Hub** preset.

**Hybrid connectivity hub** -- A distinct ASN and gateway CIDR blocks prepared for VPN and Direct Connect termination alongside VPC attachments. Start from the **Hybrid Connectivity Hub** preset.

## Works With

- [**AWS Transit Gateway VPC Attachment**](/infra-catalog/aws-transit-gateway-vpc-attachment) -- joins a VPC to this gateway; consumes `transit_gateway_id`
- [**AWS Transit Gateway Route Table**](/infra-catalog/aws-transit-gateway-route-table) -- custom routing domains for segmented topologies; consumes `transit_gateway_id`
- [**AWS Client VPN**](/infra-catalog/aws-client-vpn) -- can terminate remote-access VPN sessions directly on this gateway; consumes `transit_gateway_id`
- [**AWS VPC**](/infra-catalog/aws-vpc) -- provides the VPCs that attachments join to the hub
