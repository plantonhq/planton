# AwsVpcPeering

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsVpcPeeringSpec defines one side of a VPC peering connection, as a
request-XOR-accept mode union:

The REQUEST arm creates the peering from its VPC toward a peer VPC.
Same-account, same-region peerings can auto-accept and go active in
one deploy. Cross-account or cross-region peerings stay
pending-acceptance until the accepter side accepts - that side is a
second instance of THIS kind running the accept arm.

The ACCEPT arm adopts an existing pending connection by its pcx- id
and accepts it. Destroying an accept-arm instance abandons
management WITHOUT deleting the peering (AWS only lets Terraform
delete from the requester side) - the requester instance owns the
delete.

Peering is non-transitive and route-less by itself: after the
connection goes active, each side still adds routes (and security
group rules) toward the peer CIDR. One VPC pair supports at most one
peering - AWS returns the EXISTING connection id for a duplicate
request, so never declare the same pair twice.

## Example

```yaml
# Canonical AwsVpcPeering example (hack/dev manifest and refgen Example
# source): a same-account, same-region peering that auto-accepts and
# enables DNS resolution both ways.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsVpcPeering
metadata:
  name: app-to-data
  id: app-to-data
  org: test-org
  env: dev
spec:
  region: us-west-2
  request:
    vpcId:
      value: vpc-0123456789abcdef0
    peerVpcId:
      value: vpc-0fedcba9876543210
    autoAccept: true
    requesterAllowRemoteVpcDnsResolution: true
    accepterAllowRemoteVpcDnsResolution: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.request` | `AwsVpcPeeringRequest` |  |  |  |
| `spec.request.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.request.peerVpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.request.peerOwnerId` | `string` |  |  |  |
| `spec.request.peerRegion` | `string` |  |  |  |
| `spec.request.autoAccept` | `bool` |  |  |  |
| `spec.request.requesterAllowRemoteVpcDnsResolution` | `bool` |  |  |  |
| `spec.request.accepterAllowRemoteVpcDnsResolution` | `bool` |  |  |  |
| `spec.accept` | `AwsVpcPeeringAccept` |  |  |  |
| `spec.accept.vpcPeeringConnectionId` | `string \| valueFrom` | yes |  | AwsVpcPeering (`status.outputs.peering_connection_id`) |
| `spec.accept.autoAccept` | `bool` |  |  |  |
| `spec.accept.accepterAllowRemoteVpcDnsResolution` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region this side of the peering operates in. Example:
"us-east-1". For the request arm this is the requester VPC's
region; for the accept arm the accepter VPC's region.

- rule: {"string":{"minLen":"1"}}

### spec.request

`AwsVpcPeeringRequest`

The request arm: create the peering from this VPC.

- rule: auto_accept cannot be combined with peer_region - a cross-region peering is accepted from the accepter side (a second instance running the accept arm)
- rule: auto_accept cannot be combined with peer_owner_id - a cross-account peering is accepted by the peer account (a second instance running the accept arm)
- rule: accepter_allow_remote_vpc_dns_resolution from the request arm requires auto_accept (same-account) - in cross-account/cross-region topologies the accept-arm instance sets it

### spec.request.vpcId

`string | valueFrom` · required

The requester VPC - the side this instance's credentials own.
Fixed for life. Reference an AwsVpc vpc_id output or pass a
literal vpc-... id.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.request.peerVpcId

`string | valueFrom` · required

The accepter VPC. Fixed for life. Same-account peers reference an
AwsVpc vpc_id output; cross-account peers pass the literal peer
vpc-... id (with peer_owner_id).

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.request.peerOwnerId

`string`

The AWS account that owns the peer VPC. Unset means this account.
Fixed for life.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.request.peerRegion

`string`

The region the peer VPC lives in, for cross-region peering. Unset
means this region. Fixed for life. Cross-region peerings never
auto-accept.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.request.autoAccept

`bool`

Accept the peering immediately after creating it. Only possible
when both VPCs are in this account and region - anything else
leaves the connection pending-acceptance for the accepter side's
instance.

### spec.request.requesterAllowRemoteVpcDnsResolution

`bool`

Let instances in the peer VPC resolve THIS VPC's private DNS
hostnames to private IPs. Requires enable_dns_hostnames on the
participating VPCs, and an ACTIVE connection - on a pending
cross-account peering AWS rejects the modification until accepted
(deploy again after acceptance, or set it from the accept arm's
side).

### spec.request.accepterAllowRemoteVpcDnsResolution

`bool`

Let instances in THIS VPC resolve the peer VPC's private DNS
hostnames. From the request arm this is settable only on an
auto-accepted (same-account) connection; cross-account topologies
set it on the accept-arm instance.

### spec.accept

`AwsVpcPeeringAccept`

The accept arm: accept a pending peering from the accepter side.

### spec.accept.vpcPeeringConnectionId

`string | valueFrom` · required

The pending connection to accept (pcx-...). Fixed for life.
Reference the requester instance's peering_connection_id output
(same-account request→accept wiring) or pass the literal id
shared by the requesting account.

- references: AwsVpcPeering (`status.outputs.peering_connection_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpcPeering, name: <that resource's name>, fieldPath: status.outputs.peering_connection_id}} -- a bare string does not parse

### spec.accept.autoAccept

`bool`

Accept the pending connection on deploy. Leave true - false only
adopts the connection into management without accepting it (it
stays pending).

### spec.accept.accepterAllowRemoteVpcDnsResolution

`bool`

Let instances in the requester VPC resolve THIS (accepter) VPC's
private DNS hostnames. Requires enable_dns_hostnames on the
participating VPCs and an active connection.

## Validation Rules

- `spec.exactly_one_arm`: configure exactly one of request (create a peering from your VPC) and accept (accept a pending peering by id)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsVpcPeering, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.peering_connection_id` | `string` | The peering connection's id (pcx-...) - what route tables and accept-arm instances reference, and the provider's import ID. |
| `status.outputs.accept_status` | `string` | The connection's acceptance status after this side's deploy: "active" for auto-accepted or accepted connections, "pending-acceptance" for a cross-account/cross-region request still waiting on its accepter. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.request.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.request.peerVpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.accept.vpcPeeringConnectionId` | AwsVpcPeering | `status.outputs.peering_connection_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsVpcPeering | `spec.accept.vpcPeeringConnectionId` | `status.outputs.peering_connection_id` |

## See Also

- [Overview](../README.md)
