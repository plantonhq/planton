# AwsTransitGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsTransitGatewaySpec defines the desired configuration for an AWS Transit
Gateway -- the regional networking hub that interconnects VPCs, VPN
connections, and Direct Connect gateways through a central point.

Transit Gateway replaces complex VPC peering meshes with a hub-and-spoke
topology. Attach VPCs to a single gateway (each attachment is its own
AwsTransitGatewayVpcAttachment resource) and they communicate through the
gateway's route tables without individual peering connections.

The gateway itself is a pure hub: it owns the BGP/routing defaults, the
feature dials, and the default route table pair. What composes AROUND it
is modeled as first-class resources that reference the gateway's outputs:
- AwsTransitGatewayVpcAttachment connects one VPC (via subnets) to the hub.
- AwsTransitGatewayRouteTable defines an isolated routing domain with its
  own associations, propagations, and static routes.
With default association + propagation enabled (the defaults), every
attachment lands in the default route table and advertises its VPC CIDRs
there -- full-mesh connectivity with zero routing configuration. Disable
one or both to build segmented topologies (prod/non-prod isolation,
inspection VPC hair-pinning) with custom route tables.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsTransitGateway
metadata:
  name: test-tgw
  org: test-org
  env: dev
  id: test-tgw-dev
spec:
  region: us-west-2
  description: Test Transit Gateway hub
  amazonSideAsn: 64513
  defaultRouteTableAssociation: false
  defaultRouteTablePropagation: false
  dnsSupport: true
  vpnEcmpSupport: true
  securityGroupReferencingSupport: true
  transitGatewayCidrBlocks:
    - 10.255.0.0/24
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.amazonSideAsn` | `int64` |  | `64512` |  |
| `spec.defaultRouteTableAssociation` | `bool` |  | `true` |  |
| `spec.defaultRouteTablePropagation` | `bool` |  | `true` |  |
| `spec.dnsSupport` | `bool` |  | `true` |  |
| `spec.vpnEcmpSupport` | `bool` |  | `true` |  |
| `spec.autoAcceptSharedAttachments` | `bool` |  |  |  |
| `spec.securityGroupReferencingSupport` | `bool` |  |  |  |
| `spec.multicastSupport` | `bool` |  |  |  |
| `spec.encryptionSupport` | `bool` |  |  |  |
| `spec.transitGatewayCidrBlocks` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the Transit Gateway will be created. All attached
VPCs must be in this region (cross-region connectivity uses TGW peering,
a separate surface).
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description for the Transit Gateway. Appears in the AWS
console and CLI output.

### spec.amazonSideAsn

`int64`

Private Autonomous System Number (ASN) for the Amazon side of BGP
sessions. Used when connecting VPNs or Direct Connect gateways; pick a
number that does not collide with your on-premises ASNs. Changing it
after creation replaces the gateway.

Valid ranges: 64512-65534 (16-bit private) or 4200000000-4294967294
(32-bit private). When omitted, AWS assigns the default 64512.

- default: `64512`

### spec.defaultRouteTableAssociation

`bool` · optional (explicit presence)

Automatically associate new attachments with the gateway's default route
table. When enabled (the default when omitted), every new attachment is
associated without manual intervention -- the full-mesh starting point.

Disable for segmented topologies where every attachment is associated
explicitly with an AwsTransitGatewayRouteTable. AWS QUIRK: flipping this
dial from disabled back to enabled REPLACES the gateway (and with it,
every attachment); disabling an enabled gateway updates in place.

- default: `true`

### spec.defaultRouteTablePropagation

`bool` · optional (explicit presence)

Automatically propagate routes from new attachments to the default route
table. When enabled (the default when omitted), every attached VPC's
CIDR blocks are advertised into the default table, creating full-mesh
reachability.

Disable for isolated routing domains where propagations are declared
per-route-table on AwsTransitGatewayRouteTable resources. AWS QUIRK:
like the association dial, flipping from disabled back to enabled
REPLACES the gateway; disabling updates in place.

- default: `true`

### spec.dnsSupport

`bool` · optional (explicit presence)

Enable DNS resolution for instances in attached VPCs. When enabled (the
default when omitted), queries to public DNS hostnames of instances in
other attached VPCs resolve to their private IP addresses.

- default: `true`

### spec.vpnEcmpSupport

`bool` · optional (explicit presence)

Enable Equal Cost Multi-Path (ECMP) routing for VPN connections. When
enabled (the default when omitted) and multiple VPN tunnels advertise
the same routes, traffic is distributed across all tunnels for higher
aggregate throughput.

- default: `true`

### spec.autoAcceptSharedAttachments

`bool`

Automatically accept cross-account attachment requests shared via AWS
Resource Access Manager (RAM). Disabled by default: shared attachments
require explicit acceptance, which is the safer posture for a hub that
other accounts can request to join.

### spec.securityGroupReferencingSupport

`bool`

Enable cross-VPC security group referencing. When enabled, security
group rules in one attached VPC can reference security groups in
another VPC connected through this gateway, replacing broad CIDR-based
rules with precise group-to-group rules. Individual attachments can
override this dial.

### spec.multicastSupport

`bool`

Enable multicast traffic routing through the Transit Gateway. This is
create-time immutable: changing it replaces the entire gateway. Only
enable for a clear multicast use case (financial market data feeds,
media streaming); the multicast domain surface itself is a separate
resource family.

### spec.encryptionSupport

`bool` · optional (explicit presence)

Enforce encryption of in-transit traffic through the Transit Gateway.
When left unset, AWS applies its own default and the effective value is
computed after creation. Only set this when you explicitly need to pin
the encryption posture -- the tri-state (unset / enabled / disabled) is
sent to AWS only when present.

### spec.transitGatewayCidrBlocks

`[]string`

CIDR blocks to associate with the Transit Gateway. Used for advanced
features like TGW Connect (SD-WAN/third-party appliance integration
over GRE) where the gateway itself needs routable addresses.

Most deployments do not need TGW CIDR blocks -- leave empty unless you
are using TGW Connect. Maximum 5 blocks. IPv4 blocks must be /24 or
larger (a numerically smaller prefix length), IPv6 blocks /64 or
larger, and the link-local range 169.254.0.0/16 is rejected by AWS.

## Validation Rules

- `transit_gateway_cidr_blocks_max_5`: transit_gateway_cidr_blocks supports a maximum of 5 CIDR blocks
- `transit_gateway_cidr_blocks_valid`: each CIDR block must include a prefix length (IPv4 /24 or larger, IPv6 /64 or larger) and must not be in 169.254.0.0/16
- `amazon_side_asn_valid_range`: amazon_side_asn must be in range 64512-65534 (16-bit) or 4200000000-4294967294 (32-bit) when set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsTransitGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.transit_gateway_id` | `string` | The Transit Gateway ID (e.g., "tgw-0123456789abcdef0"). This is the primary identifier used by VPC attachments, route tables, subnet routes, VPN connections, Direct Connect gateways, and peering attachments. |
| `status.outputs.transit_gateway_arn` | `string` | The Amazon Resource Name (ARN) of the Transit Gateway. Used for IAM policies, resource-level permissions, and AWS RAM sharing. |
| `status.outputs.owner_id` | `string` | The AWS account ID that owns this Transit Gateway. |
| `status.outputs.association_default_route_table_id` | `string` | The ID of the default association route table. Attachments created with default route table association enabled are automatically associated with this table. Empty when the gateway is created with default association disabled. |
| `status.outputs.propagation_default_route_table_id` | `string` | The ID of the default propagation route table. Attachments created with default route table propagation enabled automatically advertise their routes into this table. Empty when the gateway is created with default propagation disabled. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsClientVpn | `spec.transitGatewayConfiguration.transitGatewayId` | `status.outputs.transit_gateway_id` |
| AwsTransitGatewayRouteTable | `spec.transitGatewayId` | `status.outputs.transit_gateway_id` |
| AwsTransitGatewayVpcAttachment | `spec.transitGatewayId` | `status.outputs.transit_gateway_id` |

## See Also

- [Overview](../README.md)
