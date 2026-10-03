# AWS Transit Gateway

An AWS Transit Gateway is the regional networking hub that interconnects VPCs, VPN connections, and Direct Connect gateways through a single, centralized point. It replaces complex VPC peering meshes with a scalable hub-and-spoke topology.

This component provisions the hub itself. What composes AROUND the hub is modeled as first-class resources:

- **`AwsTransitGatewayVpcAttachment`** connects one VPC (via subnets) to the gateway.
- **`AwsTransitGatewayRouteTable`** defines an isolated routing domain -- its associations, propagations, and static routes.

With the default association and propagation dials enabled (the defaults), every attachment lands in the gateway's default route table and advertises its VPC CIDRs there -- full-mesh connectivity with zero routing configuration. Disable one or both to build segmented topologies from custom route tables.

## When to Use

- **Multi-VPC connectivity**: Connect 2 or more VPCs that need to communicate. A Transit Gateway scales far better than VPC peering (which needs N*(N-1)/2 connections).
- **Hybrid networking**: Connect on-premises networks to multiple VPCs through VPN or Direct Connect via a single attachment point.
- **Centralized inspection**: Route inter-VPC traffic through a firewall appliance VPC (appliance mode on that VPC's attachment keeps flows AZ-symmetric).
- **Shared services**: Expose a shared-services VPC (DNS, monitoring, logging) to all application VPCs while keeping the applications isolated from each other.

## When NOT to Use

- **Single VPC**: No inter-VPC traffic means no hub to route it.
- **Two VPCs only**: Simple VPC peering may be more cost-effective for exactly two VPCs with low traffic.
- **Cross-region**: Each Transit Gateway is regional. Cross-region connectivity uses TGW peering attachments -- a separate surface (see the docs deferral ledger).

## Spec Fields

### Core Configuration

| Field | Type | Default | Description |
|---|---|---|---|
| `region` | string | (required) | AWS region for the gateway |
| `description` | string | - | Human-readable description |
| `amazonSideAsn` | int64 | 64512 | BGP ASN for the Amazon side (64512-65534 or 4200000000-4294967294). Changing it replaces the gateway |

### Routing Behavior

| Field | Type | Default | Description |
|---|---|---|---|
| `defaultRouteTableAssociation` | optional bool | true | Auto-associate new attachments with the default route table. Disable -> enable REPLACES the gateway |
| `defaultRouteTablePropagation` | optional bool | true | Auto-propagate attachment routes into the default route table. Disable -> enable REPLACES the gateway |

### Feature Dials

| Field | Type | Default | Description |
|---|---|---|---|
| `dnsSupport` | optional bool | true | DNS resolution across attached VPCs |
| `vpnEcmpSupport` | optional bool | true | Equal Cost Multi-Path for VPN connections |
| `autoAcceptSharedAttachments` | bool | false | Auto-accept RAM-shared attachment requests |
| `securityGroupReferencingSupport` | bool | false | Cross-VPC security group references |
| `multicastSupport` | bool | false | Multicast routing (create-time immutable) |
| `encryptionSupport` | optional bool | AWS-computed | Enforce in-transit encryption; leave unset unless pinning the posture |

### CIDR Blocks

| Field | Type | Default | Description |
|---|---|---|---|
| `transitGatewayCidrBlocks` | string[] | [] | TGW CIDR blocks for TGW Connect/GRE (max 5; IPv4 /24 or larger, IPv6 /64 or larger, never 169.254.0.0/16) |

## Outputs

| Output | Description |
|---|---|
| `transit_gateway_id` | The gateway ID referenced by attachments, route tables, subnet routes, VPN connections, and Direct Connect gateways |
| `transit_gateway_arn` | ARN for IAM policies and AWS RAM sharing |
| `owner_id` | Owning AWS account ID |
| `association_default_route_table_id` | Default association route table (empty when disabled) |
| `propagation_default_route_table_id` | Default propagation route table (empty when disabled) |

## Composition

- An `AwsTransitGatewayVpcAttachment` references `status.outputs.transit_gateway_id` to join a VPC to this hub.
- An `AwsTransitGatewayRouteTable` references the same output to carve an isolated routing domain.
- An `AwsSubnet` route with `targetType: transit_gateway` sends VPC traffic to this hub (compose `targetId` from this resource's `transit_gateway_id` output).

## Example: full-mesh hub

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsTransitGateway
metadata:
  name: core-hub
spec:
  region: us-east-1
  description: Full-mesh hub for the platform VPCs
```

## Example: segmented hub (custom routing domains)

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsTransitGateway
metadata:
  name: segmented-hub
spec:
  region: us-east-1
  description: Segmented hub - custom route tables own all routing
  defaultRouteTableAssociation: false
  defaultRouteTablePropagation: false
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
