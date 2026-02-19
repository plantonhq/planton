---
title: "Elastic IP"
description: "Elastic IP deployment documentation"
icon: "package"
order: 100
componentName: "awselasticip"
---

# AWS Elastic IP

Deploys a static public IPv4 address from Amazon's pool or a Bring-Your-Own-IP range. The allocated Elastic IP persists until explicitly released, providing a stable public endpoint for Network Load Balancers, NAT Gateways, and EC2 instances.

## What Gets Created

When you deploy an AwsElasticIp resource, OpenMCF provisions:

- **Elastic IP Address** — an `aws_eip` resource in the VPC domain, allocated from Amazon's default IPv4 pool or from a BYOIP pool when `publicIpv4Pool` is specified

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A BYOIP address range** registered with AWS if using `publicIpv4Pool` (optional — most deployments use Amazon's default pool)

## Quick Start

Create a file `eip.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: my-eip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsElasticIp.my-eip
spec:
  region: us-east-1
```

Deploy:

```shell
openmcf apply -f eip.yaml
```

This allocates a VPC Elastic IP from Amazon's default pool in `us-east-1`. The `allocation_id` and `public_ip` outputs are immediately available for downstream references.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | `string` | AWS region where the Elastic IP will be allocated (e.g., `us-east-1`). |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `publicIpv4Pool` | `string` | Amazon's pool | BYOIP pool ID to allocate from (e.g., `ipv4pool-ec2-xxx`). ForceNew — changing this replaces the EIP. |
| `address` | `string` | — | Specific IPv4 address to allocate from the BYOIP pool. Requires `publicIpv4Pool` to be set. ForceNew. |
| `networkBorderGroup` | `string` | Region default | Location scope for the EIP. Set to a Local Zone or Wavelength zone identifier to constrain the EIP to that zone. ForceNew. |

## Examples

### Standard EIP for NLB

Allocate Elastic IPs and bind them to a Network Load Balancer for static ingress IPs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: nlb-eip-az1
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsElasticIp.nlb-eip-az1
spec:
  region: us-east-1
```

### Using Foreign Key References

Wire the Elastic IP into an NLB subnet mapping via `valueFrom`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: api-nlb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsNetworkLoadBalancer.api-nlb
spec:
  subnetMappings:
    - subnetId:
        valueFrom:
          kind: AwsVpc
          name: prod-vpc
          field: status.outputs.public_subnets.[0].id
      allocationId:
        valueFrom:
          kind: AwsElasticIp
          name: nlb-eip-az1
          field: status.outputs.allocation_id
  listeners:
    - name: https
      port: 443
      protocol: TLS
      tlsConfig:
        certificateArn:
          valueFrom:
            kind: AwsCertManagerCert
            name: api-cert
            field: status.outputs.certificate_arn
      targetGroup:
        port: 8443
        protocol: TCP
        targetType: ip
```

### BYOIP Pool Allocation

Allocate from your organization's registered IP address range:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: byoip-eip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsElasticIp.byoip-eip
spec:
  region: us-east-1
  publicIpv4Pool: ipv4pool-ec2-0123456789abcdef0
  address: "198.51.100.10"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `allocation_id` | `string` | Allocation ID of the Elastic IP (e.g., `eipalloc-0123456789abcdef0`). Primary identifier for NLB subnet mappings and NAT Gateways. |
| `public_ip` | `string` | The public IPv4 address assigned to this Elastic IP. |
| `arn` | `string` | ARN of the Elastic IP, used for IAM policy resource conditions. |
| `public_dns` | `string` | Public DNS hostname (e.g., `ec2-1-2-3-4.compute-1.amazonaws.com`). |

## Related Components

- [AwsNetworkLoadBalancer](/docs/catalog/aws/network-load-balancer) — uses `allocationId` in subnet mappings for static public IPs
- [AwsVpc](/docs/catalog/aws/vpc) — provides the VPC and subnets where EIP consumers are deployed
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls traffic to resources associated with the EIP
