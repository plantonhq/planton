---
title: "ALB"
description: "ALB deployment documentation"
icon: "package"
order: 100
componentName: "awsalb"
---

# AWS ALB

Deploys an AWS Application Load Balancer with automatic listener configuration, optional SSL termination via ACM, and optional Route53 DNS record management. The component enforces multi-AZ deployment by requiring at least two subnets.

## What Gets Created

When you deploy an AwsAlb resource, OpenMCF provisions:

- **Application Load Balancer** — an `aws_lb` resource of type `application`, placed in the specified subnets with attached security groups
- **HTTP Listener (port 80)** — if SSL is disabled, serves a fixed `200 OK` response; if SSL is enabled, redirects to HTTPS with a `301`
- **HTTPS Listener (port 443)** — created only when SSL is enabled, terminates TLS using the specified ACM certificate with policy `ELBSecurityPolicy-2016-08`, serves a fixed `200 OK` response as the default action
- **Route53 A Records** — created only when DNS is enabled, one alias record per hostname pointing to the ALB's DNS name with target health evaluation enabled

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **At least two subnets** in different Availability Zones (public subnets for internet-facing, private for internal)
- **A security group** allowing inbound traffic on port 80 and/or 443
- **An ACM certificate ARN** if enabling SSL
- **A Route53 hosted zone** if enabling DNS management

## Quick Start

Create a file `alb.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: my-alb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsAlb.my-alb
spec:
  region: us-west-2
  subnets:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  securityGroups:
    - sg-0a1b2c3d4e5f00001
```

Deploy:

```shell
openmcf apply -f alb.yaml
```

This creates an internet-facing ALB with an HTTP listener on port 80 across two subnets.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the ALB will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `subnets` | `string[]` | Subnet IDs where the ALB is placed. Use public subnets for internet-facing, private for internal. | Minimum 2 items required |
| `subnets[].value` | `string` | Direct subnet ID value | — |
| `subnets[].valueFrom` | `object` | Foreign key reference to an AwsVpc resource | Default kind: `AwsVpc` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `securityGroups` | `string[]` | `[]` | Security group IDs to attach to the ALB. Can reference AwsSecurityGroup resources via `valueFrom`. |
| `internal` | `bool` | `false` | When `true`, creates an internal ALB accessible only within the VPC. When `false`, creates an internet-facing ALB. |
| `deleteProtectionEnabled` | `bool` | `false` | Prevents accidental deletion of the ALB when enabled. |
| `idleTimeoutSeconds` | `int` | `60` | Connection idle timeout in seconds before the ALB closes the connection. |
| `ssl.enabled` | `bool` | `false` | Enables HTTPS listener on port 443 and configures HTTP-to-HTTPS redirect on port 80. |
| `ssl.certificateArn` | `string` | — | ACM certificate ARN for TLS termination. Required when `ssl.enabled` is `true`. Can reference an AwsCertManagerCert resource via `valueFrom`. |
| `dns.enabled` | `bool` | `false` | Enables automatic Route53 DNS record creation for the ALB. |
| `dns.route53ZoneId` | `string` | — | Route53 hosted zone ID where records are created. Required when `dns.enabled` is `true`. Can reference an AwsRoute53Zone resource via `valueFrom`. |
| `dns.hostnames` | `string[]` | `[]` | Domain names that will point to the ALB via Route53 alias records. Must be unique. |

## Examples

### Internal ALB

An ALB accessible only within the VPC, useful for internal microservice routing:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: internal-api-alb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsAlb.internal-api-alb
spec:
  region: us-west-2
  subnets:
    - subnet-private-az1
    - subnet-private-az2
  securityGroups:
    - sg-internal-alb
  internal: true
```

### ALB with SSL Termination

HTTPS-enabled ALB that redirects HTTP to HTTPS automatically:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: web-alb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsAlb.web-alb
spec:
  region: us-west-2
  subnets:
    - subnet-public-az1
    - subnet-public-az2
  securityGroups:
    - sg-web-alb
  deleteProtectionEnabled: true
  idleTimeoutSeconds: 90
  ssl:
    enabled: true
    certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/abc-12345
```

### Full-Featured ALB with DNS

Production configuration with SSL, DNS management, and deletion protection:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: prod-alb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsAlb.prod-alb
spec:
  region: us-west-2
  subnets:
    - subnet-public-az1
    - subnet-public-az2
    - subnet-public-az3
  securityGroups:
    - sg-prod-alb
  deleteProtectionEnabled: true
  idleTimeoutSeconds: 120
  ssl:
    enabled: true
    certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/prod-cert
  dns:
    enabled: true
    route53ZoneId: Z0123456789ABCDEFGHIJ
    hostnames:
      - app.example.com
      - api.example.com
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: ref-alb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsAlb.ref-alb
spec:
  region: us-west-2
  subnets:
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.public_subnets[0].id
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.public_subnets[1].id
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: alb-sg
        field: status.outputs.security_group_id
  ssl:
    enabled: true
    certificateArn:
      valueFrom:
        kind: AwsCertManagerCert
        name: my-cert
        field: status.outputs.cert_arn
  dns:
    enabled: true
    route53ZoneId:
      valueFrom:
        kind: AwsRoute53Zone
        name: my-zone
        field: status.outputs.zone_id
    hostnames:
      - app.example.com
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `load_balancer_arn` | `string` | ARN of the created Application Load Balancer |
| `load_balancer_name` | `string` | Name assigned to the ALB (may differ from `metadata.name`) |
| `load_balancer_dns_name` | `string` | DNS name automatically assigned by AWS (e.g., `my-alb-123456.us-east-1.elb.amazonaws.com`) |
| `load_balancer_hosted_zone_id` | `string` | Route53 hosted zone ID for the ALB's DNS entry, used for creating alias records |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides the subnets for ALB placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls inbound and outbound traffic to the ALB
- [AwsRoute53Zone](/docs/catalog/aws/route53-zone) — hosts the DNS zone for alias records
- [AwsCertManagerCert](/docs/catalog/aws/certificate) — provides the ACM certificate for SSL termination
