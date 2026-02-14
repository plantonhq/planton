---
title: "Client VPN"
description: "Client VPN deployment documentation"
icon: "package"
order: 100
componentName: "awsclientvpn"
---

# AWS Client VPN

Deploys an AWS Client VPN endpoint attached to a VPC, with subnet associations, certificate-based mutual TLS authentication, and configurable authorization rules. The component provisions a managed OpenVPN server that enables clients to securely connect to private VPC resources.

## What Gets Created

When you deploy an AwsClientVpn resource, OpenMCF provisions:

- **Client VPN Endpoint** — an `aws:ec2clientvpn:Endpoint` resource configured with mutual TLS authentication, the specified server certificate, client CIDR block, and connection logging settings
- **Network Associations** — one `aws:ec2clientvpn:NetworkAssociation` per subnet in the `subnets` list, linking the VPN endpoint to each subnet's Availability Zone
- **Authorization Rules** — one `aws:ec2clientvpn:AuthorizationRule` per entry in `cidrAuthorizationRules`, granting all VPN clients access to the specified network CIDR ranges

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A VPC** with at least one subnet for the VPN endpoint association
- **An ACM server certificate** for TLS termination (the certificate must be in the same AWS region)
- **A client certificate** signed by the same CA, distributed to each VPN user
- **A non-overlapping CIDR block** for client IP assignment (between /12 and /22)

## Quick Start

Create a file `client-vpn.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsClientVpn
metadata:
  name: my-vpn
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsClientVpn.my-vpn
spec:
  vpcId: vpc-0a1b2c3d4e5f00001
  subnets:
    - subnet-0a1b2c3d4e5f00001
  clientCidrBlock: "10.100.0.0/22"
  serverCertificateArn: arn:aws:acm:us-east-1:123456789012:certificate/abc-12345
```

Deploy:

```shell
openmcf apply -f client-vpn.yaml
```

This creates a Client VPN endpoint with certificate-based authentication, split-tunnel routing, TCP transport on port 443, and a single subnet association.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `vpcId` | `StringValueOrRef` | Target VPC for the Client VPN endpoint. Can reference an AwsVpc resource via `valueFrom`. | Required |
| `subnets` | `StringValueOrRef[]` | Subnet IDs to associate as target networks. Each enables access in that subnet's Availability Zone. Can reference AwsVpc resources via `valueFrom`. | Minimum 1 item |
| `clientCidrBlock` | `string` | IPv4 CIDR for client IP assignment (e.g., `10.100.0.0/22`). Must not overlap with the VPC CIDR. | Valid IPv4 CIDR, between /12 and /22 |
| `serverCertificateArn` | `StringValueOrRef` | ARN of the ACM certificate for the VPN server. Can reference an AwsCertManagerCert resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-friendly description for the Client VPN endpoint, visible in the AWS Console. |
| `authenticationType` | `enum` | `certificate` | Authentication method. Valid values: `certificate`, `directory`, `cognito`. Only `certificate` is supported in v1. |
| `cidrAuthorizationRules` | `string[]` | `[]` | IPv4 CIDRs that VPN clients are authorized to access. One authorization rule is created per entry. Must be unique, valid IPv4 CIDRs. |
| `disableSplitTunnel` | `bool` | `false` | When `false` (default), only traffic for authorized CIDRs routes through the VPN. When `true`, all client traffic routes through the VPN (full-tunnel). |
| `vpnPort` | `int32` | `443` | Port for VPN connections. Allowed values: `443` (TCP) or `1194` (UDP). |
| `transportProtocol` | `enum` | `tcp` | Transport protocol for VPN sessions. Valid values: `udp`, `tcp`. Must match `vpnPort` (TCP with 443, UDP with 1194). |
| `logGroupName` | `string` | — | CloudWatch Logs group name for connection logging. If blank, logging is disabled. |
| `securityGroups` | `StringValueOrRef[]` | `[]` | Security group IDs for the VPN endpoint's network associations. If omitted, the VPC default security group applies. Can reference AwsSecurityGroup resources via `valueFrom`. |
| `dnsServers` | `string[]` | `[]` | Custom DNS server IPs for VPN clients. Maximum 2 entries. If omitted, clients use the VPC's AmazonProvidedDNS. |

## Examples

### Single-Subnet VPN with Authorization

A minimal VPN endpoint granting access to an internal network:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsClientVpn
metadata:
  name: dev-vpn
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsClientVpn.dev-vpn
spec:
  vpcId: vpc-dev-001
  subnets:
    - subnet-private-az1
  clientCidrBlock: "10.100.0.0/22"
  serverCertificateArn: arn:aws:acm:us-east-1:123456789012:certificate/dev-cert
  cidrAuthorizationRules:
    - "10.0.0.0/16"
```

### Multi-AZ VPN with Logging

A VPN endpoint spanning two Availability Zones with CloudWatch connection logging enabled:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsClientVpn
metadata:
  name: team-vpn
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AwsClientVpn.team-vpn
spec:
  vpcId: vpc-staging-001
  subnets:
    - subnet-private-az1
    - subnet-private-az2
  clientCidrBlock: "10.200.0.0/22"
  serverCertificateArn: arn:aws:acm:us-east-1:123456789012:certificate/staging-cert
  cidrAuthorizationRules:
    - "10.0.0.0/16"
    - "172.16.0.0/12"
  logGroupName: /aws/clientvpn/team-vpn
  description: "Staging environment VPN for engineering team"
```

### Full-Tunnel VPN with Custom DNS

Routes all client traffic through the VPN and overrides DNS resolution with custom servers:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsClientVpn
metadata:
  name: secure-vpn
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsClientVpn.secure-vpn
spec:
  vpcId: vpc-prod-001
  subnets:
    - subnet-private-az1
    - subnet-private-az2
  clientCidrBlock: "10.150.0.0/22"
  serverCertificateArn: arn:aws:acm:us-east-1:123456789012:certificate/prod-cert
  cidrAuthorizationRules:
    - "10.0.0.0/8"
  disableSplitTunnel: true
  dnsServers:
    - "10.0.0.2"
    - "10.0.1.2"
  logGroupName: /aws/clientvpn/secure-vpn
  securityGroups:
    - sg-vpn-endpoint
```

### UDP Transport on Port 1194

A VPN endpoint using UDP for lower-latency connections on the standard OpenVPN port:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsClientVpn
metadata:
  name: low-latency-vpn
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsClientVpn.low-latency-vpn
spec:
  vpcId: vpc-prod-001
  subnets:
    - subnet-private-az1
  clientCidrBlock: "10.250.0.0/22"
  serverCertificateArn: arn:aws:acm:us-east-1:123456789012:certificate/prod-cert
  vpnPort: 1194
  transportProtocol: udp
  cidrAuthorizationRules:
    - "10.0.0.0/16"
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsClientVpn
metadata:
  name: ref-vpn
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsClientVpn.ref-vpn
spec:
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: my-vpc
      field: status.outputs.vpc_id
  subnets:
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets[0]
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets[1]
  clientCidrBlock: "10.100.0.0/22"
  serverCertificateArn:
    valueFrom:
      kind: AwsCertManagerCert
      name: vpn-cert
      field: status.outputs.cert_arn
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: vpn-sg
        field: status.outputs.security_group_id
  cidrAuthorizationRules:
    - "10.0.0.0/16"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `client_vpn_endpoint_id` | `string` | AWS-assigned identifier for the Client VPN endpoint (e.g., `cvpn-endpoint-012345abcdeEXAMPLE`) |
| `security_group_id` | `string` | ID of the security group applied to the endpoint's network associations |
| `subnet_association_ids` | `map<string, string>` | Map of subnet ID to AWS association ID for each associated subnet |
| `endpoint_dns_name` | `string` | DNS name clients use to connect to the VPN endpoint |

## Related Components

- [AwsVpc](/docs/catalog/aws/awsvpc) — provides the VPC and subnets for VPN endpoint placement
- [AwsSecurityGroup](/docs/catalog/aws/awssecuritygroup) — controls traffic between VPN clients and VPC resources
- [AwsCertManagerCert](/docs/catalog/aws/awscertmanagercert) — provides the ACM server certificate for TLS termination
- [AwsEc2Instance](/docs/catalog/aws/awsec2instance) — private instances that VPN clients typically connect to
