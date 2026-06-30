---
title: "Transit Gateway"
description: "Transit Gateway deployment documentation"
icon: "package"
order: 100
componentName: "awstransitgateway"
---

# AWS Transit Gateway

Deploys an AWS Transit Gateway with inline VPC attachments, replacing complex VPC peering meshes with a hub-and-spoke topology. The component bundles the Transit Gateway and its VPC attachments together because a TGW without attachments serves no purpose. Default routing behavior provides full-mesh connectivity out of the box.

## What Gets Created

When you deploy an AwsTransitGateway resource, Planton provisions:

- **Transit Gateway** — an `ec2transitgateway.TransitGateway` resource with the configured ASN, routing behavior, DNS support, and feature toggles
- **VPC Attachments** — one `ec2transitgateway.VpcAttachment` per entry in `vpcAttachments`, each connecting a VPC to the TGW through specified subnets with per-attachment DNS, IPv6, and appliance mode settings
- **Default Route Tables** — automatically created by AWS for association and propagation (IDs exposed as stack outputs for future static route management)

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **At least one VPC** with subnets in the target Availability Zones
- **Subnets** in each VPC to host the TGW elastic network interfaces (one per AZ recommended for high availability)

## Quick Start

Create a file `tgw.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsTransitGateway
metadata:
  name: my-tgw
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsTransitGateway.my-tgw
spec:
  region: us-east-1
  vpcAttachments:
    - name: app-vpc
      vpcId: vpc-0123456789abcdef0
      subnetIds:
        - subnet-0a1b2c3d4e5f00001
        - subnet-0a1b2c3d4e5f00002
```

Deploy:

```shell
planton apply -f tgw.yaml
```

This creates a Transit Gateway with default full-mesh routing and DNS support, attaching a single VPC through two subnets.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | The AWS region where the Transit Gateway will be created. | Required |
| `vpcAttachments` | `object[]` | VPC attachments connecting VPCs to the Transit Gateway. | Minimum 1 item required |
| `vpcAttachments[].name` | `string` | Unique name for the attachment. Used as key in `vpc_attachment_ids` output. | Lowercase alphanumeric and hyphens, starts with letter, max 63 chars |
| `vpcAttachments[].vpcId` | `StringValueOrRef` | VPC ID to attach. ForceNew. Can reference AwsVpc via `valueFrom`. | Required |
| `vpcAttachments[].subnetIds` | `StringValueOrRef[]` | Subnets for TGW network interfaces. One per AZ recommended. Can reference AwsVpc via `valueFrom`. | Minimum 1 item required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description for the Transit Gateway. |
| `amazonSideAsn` | `int64` | `64512` | Private ASN for BGP sessions (VPN, Direct Connect). Valid ranges: 64512-65534 or 4200000000-4294967294. |
| `defaultRouteTableAssociation` | `bool` | `true` | Auto-associate new attachments with the default route table. |
| `defaultRouteTablePropagation` | `bool` | `true` | Auto-propagate routes from new attachments to the default route table. |
| `dnsSupport` | `bool` | `true` | Resolve public DNS hostnames of instances in attached VPCs to private IPs. |
| `vpnEcmpSupport` | `bool` | `true` | Distribute traffic across multiple VPN tunnels advertising the same routes. |
| `autoAcceptSharedAttachments` | `bool` | `false` | Auto-accept cross-account attachments shared via AWS RAM. |
| `securityGroupReferencingSupport` | `bool` | `false` | Allow security group rules to reference groups in other attached VPCs. |
| `multicastSupport` | `bool` | `false` | Enable multicast routing. ForceNew. |
| `transitGatewayCidrBlocks` | `string[]` | `[]` | CIDR blocks for TGW Connect and GRE-based attachments. Maximum 5. |
| `vpcAttachments[].dnsSupport` | `bool` | `true` | Per-attachment DNS resolution override. |
| `vpcAttachments[].ipv6Support` | `bool` | `false` | Route IPv6 traffic for this attachment. |
| `vpcAttachments[].applianceModeSupport` | `bool` | `false` | Enable symmetric routing for inspection appliances (firewall, IDS/IPS). |
| `vpcAttachments[].defaultRouteTableAssociation` | `bool` | `true` | Associate this attachment with the TGW default route table. |
| `vpcAttachments[].defaultRouteTablePropagation` | `bool` | `true` | Propagate this attachment's routes to the TGW default route table. |

## Examples

### Multi-VPC Hub-and-Spoke

Connect application and database VPCs through a central Transit Gateway:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsTransitGateway
metadata:
  name: hub-tgw
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsTransitGateway.hub-tgw
spec:
  region: us-east-1
  description: Production hub connecting application and database VPCs
  vpcAttachments:
    - name: app-vpc
      vpcId: vpc-app-0123456789
      subnetIds:
        - subnet-app-az1
        - subnet-app-az2
    - name: db-vpc
      vpcId: vpc-db-0123456789
      subnetIds:
        - subnet-db-az1
        - subnet-db-az2
    - name: shared-services
      vpcId: vpc-shared-0123456789
      subnetIds:
        - subnet-shared-az1
        - subnet-shared-az2
```

### Inspection VPC with Appliance Mode

Route all inter-VPC traffic through a centralized firewall/IDS VPC using appliance mode for symmetric routing:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsTransitGateway
metadata:
  name: inspection-tgw
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsTransitGateway.inspection-tgw
spec:
  region: us-east-1
  description: Transit Gateway with centralized traffic inspection
  securityGroupReferencingSupport: true
  vpcAttachments:
    - name: inspection-vpc
      vpcId: vpc-inspection-0123456789
      subnetIds:
        - subnet-inspection-az1
        - subnet-inspection-az2
      applianceModeSupport: true
    - name: workload-vpc
      vpcId: vpc-workload-0123456789
      subnetIds:
        - subnet-workload-az1
        - subnet-workload-az2
```

### Full-Featured with Custom ASN and CIDR Blocks

Production Transit Gateway with custom BGP ASN, TGW CIDR blocks for Connect integration, and cross-account attachment acceptance:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsTransitGateway
metadata:
  name: enterprise-tgw
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsTransitGateway.enterprise-tgw
spec:
  region: us-east-1
  description: Enterprise Transit Gateway with hybrid connectivity
  amazonSideAsn: 65100
  autoAcceptSharedAttachments: true
  securityGroupReferencingSupport: true
  transitGatewayCidrBlocks:
    - 100.64.0.0/24
  vpcAttachments:
    - name: production
      vpcId: vpc-prod-0123456789
      subnetIds:
        - subnet-prod-az1
        - subnet-prod-az2
        - subnet-prod-az3
    - name: staging
      vpcId: vpc-staging-0123456789
      subnetIds:
        - subnet-staging-az1
        - subnet-staging-az2
      defaultRouteTablePropagation: false
```

### Using Foreign Key References

Reference Planton-managed VPCs instead of hardcoding IDs:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsTransitGateway
metadata:
  name: ref-tgw
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsTransitGateway.ref-tgw
spec:
  region: us-east-1
  vpcAttachments:
    - name: app-vpc
      vpcId:
        valueFrom:
          kind: AwsVpc
          name: app-vpc
          field: status.outputs.vpc_id
      subnetIds:
        - valueFrom:
            kind: AwsSubnet
            name: app-private-subnet-a
            fieldPath: status.outputs.subnet_id
        - valueFrom:
            kind: AwsSubnet
            name: app-private-subnet-b
            fieldPath: status.outputs.subnet_id
    - name: db-vpc
      vpcId:
        valueFrom:
          kind: AwsVpc
          name: db-vpc
          field: status.outputs.vpc_id
      subnetIds:
        - valueFrom:
            kind: AwsSubnet
            name: db-private-subnet-a
            fieldPath: status.outputs.subnet_id
        - valueFrom:
            kind: AwsSubnet
            name: db-private-subnet-b
            fieldPath: status.outputs.subnet_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `transit_gateway_id` | `string` | Transit Gateway ID (e.g., `tgw-0123456789abcdef0`). Used by VPN connections, Direct Connect gateways, and peering attachments. |
| `transit_gateway_arn` | `string` | Amazon Resource Name for IAM policies and AWS RAM sharing |
| `owner_id` | `string` | AWS account ID that owns the Transit Gateway |
| `association_default_route_table_id` | `string` | ID of the default association route table |
| `propagation_default_route_table_id` | `string` | ID of the default propagation route table |
| `vpc_attachment_ids` | `map<string, string>` | Map of attachment name to VPC attachment ID. Reference specific attachments via `status.outputs.vpc_attachment_ids.{name}`. |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides the VPCs and subnets attached to the Transit Gateway
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls traffic in attached VPCs; cross-VPC referencing available when `securityGroupReferencingSupport` is enabled
- [AwsClientVpn](/docs/catalog/aws/client-vpn) — VPN connectivity into the Transit Gateway network
