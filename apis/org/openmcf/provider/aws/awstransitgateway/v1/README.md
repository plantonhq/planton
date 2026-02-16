# AWS Transit Gateway

An AWS Transit Gateway is a regional networking hub that interconnects VPCs, VPN connections, and Direct Connect gateways through a single, centralized point. It replaces complex VPC peering meshes with a scalable hub-and-spoke topology.

## When to Use

- **Multi-VPC connectivity**: Connect 2 or more VPCs that need to communicate with each other. Transit Gateway scales far better than VPC peering (which requires N*(N-1)/2 connections).
- **Hybrid networking**: Connect on-premises networks to multiple VPCs through VPN or Direct Connect via a single attachment point.
- **Centralized inspection**: Route all inter-VPC traffic through a firewall appliance VPC using appliance mode for symmetric routing.
- **Shared services**: Provide a shared-services VPC (DNS, monitoring, logging) accessible from all application VPCs.

## When NOT to Use

- **Single VPC**: If you only have one VPC and no plans to expand, you do not need a Transit Gateway.
- **Two VPCs only**: Simple VPC peering may be more cost-effective for exactly two VPCs with low traffic.
- **Cross-region**: Each Transit Gateway is regional. For cross-region connectivity, you need TGW peering (not yet supported in this component's v1).

## Prerequisites

- At least one AwsVpc resource deployed with private subnets
- AWS credentials with permissions for `ec2:CreateTransitGateway`, `ec2:CreateTransitGatewayVpcAttachment`, and related actions

## Spec Fields

### Core Configuration

| Field | Type | Default | Description |
|---|---|---|---|
| `description` | string | - | Human-readable description |
| `amazonSideAsn` | int64 | 64512 | BGP ASN for the Amazon side (64512-65534 or 4200000000-4294967294) |

### Routing Behavior

| Field | Type | Default | Description |
|---|---|---|---|
| `defaultRouteTableAssociation` | bool | true | Auto-associate new attachments with default route table |
| `defaultRouteTablePropagation` | bool | true | Auto-propagate routes from new attachments |

### Feature Toggles

| Field | Type | Default | Description |
|---|---|---|---|
| `dnsSupport` | bool | true | DNS resolution across attached VPCs |
| `vpnEcmpSupport` | bool | true | Equal Cost Multi-Path for VPN connections |
| `autoAcceptSharedAttachments` | bool | false | Auto-accept RAM-shared attachments |
| `securityGroupReferencingSupport` | bool | false | Cross-VPC security group references |
| `multicastSupport` | bool | false | Multicast routing (ForceNew -- immutable after creation) |

### CIDR Blocks

| Field | Type | Default | Description |
|---|---|---|---|
| `transitGatewayCidrBlocks` | string[] | [] | TGW CIDR blocks for Connect/GRE (max 5) |

### VPC Attachments

Each VPC attachment connects one VPC to the Transit Gateway:

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `name` | string | yes | - | Unique attachment name (lowercase alphanumeric + hyphens) |
| `vpcId` | StringValueOrRef | yes | - | VPC ID to attach |
| `subnetIds` | StringValueOrRef[] | yes | - | Subnet IDs (one per AZ recommended) |
| `dnsSupport` | bool | no | true | Per-attachment DNS override |
| `ipv6Support` | bool | no | false | IPv6 traffic routing |
| `applianceModeSupport` | bool | no | false | Symmetric routing for firewall appliances |
| `defaultRouteTableAssociation` | bool | no | true | Per-attachment route table association override |
| `defaultRouteTablePropagation` | bool | no | true | Per-attachment route propagation override |

## Outputs

| Output | Description |
|---|---|
| `transit_gateway_id` | The TGW ID (primary reference for downstream resources) |
| `transit_gateway_arn` | TGW ARN for IAM policies and RAM sharing |
| `owner_id` | AWS account ID that owns the TGW |
| `association_default_route_table_id` | Default association route table ID |
| `propagation_default_route_table_id` | Default propagation route table ID |
| `vpc_attachment_ids` | Map of attachment name to VPC attachment ID |

## Deliberately Excluded from v1

The following features are intentionally deferred to keep the abstraction clean and focused:

- **Custom route tables**: Default auto-association/propagation covers 80%+ of use cases
- **Static routes**: Future `AwsTransitGatewayRoute` component can reference exported route table IDs
- **Cross-region peering**: Separate lifecycle, planned as a future component
- **Multicast domain configuration**: The multicast toggle is available, but domain/group setup is deferred
- **Connect attachments**: SD-WAN/GRE integration is highly specialized
