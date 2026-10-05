# AwsTransitGatewayVpcAttachment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsTransitGatewayVpcAttachmentSpec defines the desired configuration for an
AWS Transit Gateway VPC attachment -- the connection that plugs one VPC
into a Transit Gateway hub.

An attachment is deliberately its own resource rather than a field on the
gateway: a gateway carries many attachments (one per spoke VPC), each with
its own lifecycle, and the attachment ID is what Transit Gateway route
tables associate, propagate, and route against. AWS provisions an elastic
network interface in each chosen subnet; traffic between the VPC and the
gateway flows through those ENIs.

Design notes:
- The gateway and the VPC are create-time immutable: changing either
  replaces the attachment. The subnet set updates in place, so AZs can be
  added or removed without replacement.
- Provide one subnet per Availability Zone you want reachable through the
  gateway -- the gateway only routes traffic to/from AZs it has an ENI in.
  A dedicated small subnet (/28) per AZ for attachments is the
  AWS-recommended pattern; it keeps routing independent of workload
  subnets.
- Whether this attachment lands in the gateway's DEFAULT route table
  (association and propagation) is controlled here per-attachment,
  overriding the gateway-level dials. Custom routing domains are built by
  turning these off and composing AwsTransitGatewayRouteTable resources.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsTransitGatewayVpcAttachment
metadata:
  name: test-tgw-attachment
  org: test-org
  env: dev
  id: test-tgw-attachment-dev
spec:
  region: us-west-2
  transitGatewayId:
    value: tgw-0a1b2c3d4e5f00001
  vpcId:
    value: vpc-0a1b2c3d4e5f00001
  subnetIds:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
  dnsSupport: true
  applianceModeSupport: false
  defaultRouteTableAssociation: false
  defaultRouteTablePropagation: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.transitGatewayId` | `string \| valueFrom` | yes |  | AwsTransitGateway (`status.outputs.transit_gateway_id`) |
| `spec.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.dnsSupport` | `bool` |  | `true` |  |
| `spec.ipv6Support` | `bool` |  |  |  |
| `spec.applianceModeSupport` | `bool` |  |  |  |
| `spec.securityGroupReferencingSupport` | `bool` |  |  |  |
| `spec.defaultRouteTableAssociation` | `bool` |  |  |  |
| `spec.defaultRouteTablePropagation` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the attachment will be created. Must match the
region of both the Transit Gateway and the VPC.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.transitGatewayId

`string | valueFrom` · required

The Transit Gateway to attach to. Create-time immutable: changing it
replaces the attachment. Reference an AwsTransitGateway's
transit_gateway_id output or pass a literal ID.

- references: AwsTransitGateway (`status.outputs.transit_gateway_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsTransitGateway, name: <that resource's name>, fieldPath: status.outputs.transit_gateway_id}} -- a bare string does not parse

### spec.vpcId

`string | valueFrom` · required

The VPC to attach. Create-time immutable: changing it replaces the
attachment. Reference an AwsVpc's vpc_id output or pass a literal ID.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.subnetIds

`[]string | valueFrom` · required

Subnets in which AWS provisions the attachment's network interfaces --
at most one per Availability Zone, and all in the VPC being attached.
The gateway routes traffic only to/from AZs it has an ENI in, so cover
every AZ your workloads run in. Updatable in place (adding/removing an
AZ does not replace the attachment). Accepts direct subnet IDs or
references to AwsSubnet resources.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.dnsSupport

`bool` · optional (explicit presence)

Enable DNS resolution for this attachment. When enabled (the default
when omitted), DNS queries from the attached VPC to public hostnames of
instances in other attached VPCs resolve to private IPs. Overrides the
gateway-level dial for this attachment only.

- default: `true`

### spec.ipv6Support

`bool`

Enable IPv6 traffic over this attachment. Requires the VPC and the
chosen subnets to carry IPv6 CIDRs. Disabled by default.

### spec.applianceModeSupport

`bool`

Enable appliance mode for this attachment. Required when routing
traffic through a stateful virtual appliance (firewall, IDS/IPS) hosted
in the attached VPC: appliance mode keeps a flow's return traffic in
the same Availability Zone as the original flow, preserving symmetric
routing for stateful inspection. Only enable on the shared-services VPC
that hosts the appliances.

### spec.securityGroupReferencingSupport

`bool` · optional (explicit presence)

Enable cross-VPC security group referencing for this attachment,
overriding the gateway-level dial. When left unset, the attachment
inherits the gateway's setting (AWS computes the effective value); set
it only to pin this attachment's posture explicitly.

### spec.defaultRouteTableAssociation

`bool` · optional (explicit presence)

Associate this attachment with the gateway's default route table. When
left unset, the gateway's own default-association dial decides (AWS
computes the effective value). Set to false when this attachment is
associated with a custom AwsTransitGatewayRouteTable instead -- an
attachment can be associated with at most ONE route table, so a default
association and a custom association conflict at the AWS API.

### spec.defaultRouteTablePropagation

`bool` · optional (explicit presence)

Propagate this attachment's VPC CIDRs into the gateway's default route
table. When left unset, the gateway's own default-propagation dial
decides (AWS computes the effective value). Set to false for isolated
routing domains where propagations are declared on custom
AwsTransitGatewayRouteTable resources (an attachment CAN propagate to
many tables, so this is about keeping the default table clean, not a
hard conflict).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsTransitGatewayVpcAttachment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.attachment_id` | `string` | The Transit Gateway attachment ID (e.g., "tgw-attach-0123456789abcdef0"). Referenced by route table associations, propagations, and static routes. |
| `status.outputs.attachment_arn` | `string` | The Amazon Resource Name (ARN) of the attachment. Used for IAM policies and resource-level permissions. |
| `status.outputs.vpc_owner_id` | `string` | The AWS account ID that owns the attached VPC. Differs from the gateway owner in cross-account (RAM-shared) topologies. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.transitGatewayId` | AwsTransitGateway | `status.outputs.transit_gateway_id` |
| `spec.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsTransitGatewayRouteTable | `spec.associations[].attachmentId` | `status.outputs.attachment_id` |
| AwsTransitGatewayRouteTable | `spec.propagations` | `status.outputs.attachment_id` |
| AwsTransitGatewayRouteTable | `spec.routes[].attachmentId` | `status.outputs.attachment_id` |
| AwsTransitGatewayRouteTable | `spec.prefixListReferences[].attachmentId` | `status.outputs.attachment_id` |

## See Also

- [Overview](../README.md)
