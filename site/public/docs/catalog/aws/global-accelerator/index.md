---
title: "Global Accelerator"
description: "Global Accelerator deployment documentation"
icon: "package"
order: 100
componentName: "awsglobalaccelerator"
---

# AWS Global Accelerator

Deploys an AWS Global Accelerator with bundled listeners and regional endpoint groups, providing two static anycast IP addresses that route traffic through the AWS global network to healthy endpoints in one or more AWS regions. The component bundles the full accelerator hierarchy (accelerator, listeners, endpoint groups, endpoints) into a single resource for complete deployment in one manifest.

## What Gets Created

When you deploy an AwsGlobalAccelerator resource, OpenMCF provisions:

- **Global Accelerator** — an `aws_globalaccelerator_accelerator` with static anycast IPs, optional flow log delivery to S3, and support for both IPv4 and dual-stack addressing
- **Listeners** — one `aws_globalaccelerator_listener` per entry in `spec.listeners`, each defining the protocol (TCP or UDP), port ranges, and client affinity setting
- **Endpoint Groups** — one `aws_globalaccelerator_endpoint_group` per entry in each listener's `endpointGroups`, each targeting a specific AWS region with health check configuration, traffic dial percentage, and optional port overrides
- **Endpoints** — registered within each endpoint group, pointing to ALBs, NLBs, Elastic IPs, or EC2 instances with configurable weights and client IP preservation

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **At least one endpoint** (ALB, NLB, Elastic IP, or EC2 instance) deployed in the target region, or plan to register endpoints after the accelerator is created
- **An S3 bucket** if enabling flow logs

## Quick Start

Create a file `global-accelerator.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlobalAccelerator
metadata:
  name: my-ga
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsGlobalAccelerator.my-ga
spec:
  listeners:
    - name: tcp-443
      protocol: TCP
      portRanges:
        - fromPort: 443
          toPort: 443
      endpointGroups:
        - name: primary
```

Deploy:

```shell
openmcf apply -f global-accelerator.yaml
```

This creates a Global Accelerator with a TCP listener on port 443 and one endpoint group in the provider's default region.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `listeners` | `object[]` | Listener definitions. | Minimum 1 item |
| `listeners[].name` | `string` | Unique name for the listener. | Lowercase alphanumeric and hyphens, max 63 chars |
| `listeners[].protocol` | `string` | Layer 4 protocol: `TCP` or `UDP`. | Required |
| `listeners[].portRanges` | `object[]` | Port ranges the listener accepts traffic on. | Minimum 1, maximum 10 |
| `listeners[].portRanges[].fromPort` | `int` | First port in range (inclusive). | 1–65535 |
| `listeners[].portRanges[].toPort` | `int` | Last port in range (inclusive). | 1–65535 |
| `listeners[].endpointGroups` | `object[]` | Regional endpoint group definitions. | Minimum 1 |
| `listeners[].endpointGroups[].name` | `string` | Unique name within the listener. | Lowercase alphanumeric and hyphens, max 63 chars |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | `bool` | `true` | Whether the accelerator accepts traffic. |
| `ipAddressType` | `string` | `IPV4` | `IPV4` or `DUAL_STACK`. |
| `ipAddresses` | `string[]` | `[]` | BYOIP addresses (max 2). ForceNew. |
| `flowLogs.enabled` | `bool` | `false` | Enable flow log delivery to S3. |
| `flowLogs.s3Bucket` | `string` | — | S3 bucket for flow logs. Can reference AwsS3Bucket via `valueFrom`. |
| `flowLogs.s3Prefix` | `string` | `""` | S3 key prefix for flow logs. |
| `listeners[].clientAffinity` | `string` | `NONE` | `NONE` or `SOURCE_IP`. |
| `listeners[].endpointGroups[].endpointGroupRegion` | `string` | Provider region | AWS region. ForceNew. |
| `listeners[].endpointGroups[].healthCheckProtocol` | `string` | `TCP` | `TCP`, `HTTP`, or `HTTPS`. |
| `listeners[].endpointGroups[].healthCheckPath` | `string` | — | Required for HTTP/HTTPS. |
| `listeners[].endpointGroups[].healthCheckIntervalSeconds` | `int` | `30` | Must be exactly `10` or `30`. |
| `listeners[].endpointGroups[].thresholdCount` | `int` | `3` | Range: 1–10. |
| `listeners[].endpointGroups[].trafficDialPercentage` | `float` | `100.0` | 0.0–100.0. |
| `listeners[].endpointGroups[].endpoints[].endpointId` | `string` | — | ALB/NLB ARN, EIP, or EC2 ID. |
| `listeners[].endpointGroups[].endpoints[].weight` | `int` | `128` | 0–255. |
| `listeners[].endpointGroups[].endpoints[].clientIpPreservationEnabled` | `bool` | `false` | ALB and EC2 only. |

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `accelerator_arn` | `string` | ARN of the Global Accelerator |
| `accelerator_dns_name` | `string` | Anycast DNS name |
| `accelerator_dual_stack_dns_name` | `string` | IPv4+IPv6 DNS name |
| `accelerator_hosted_zone_id` | `string` | Route53 hosted zone ID (`Z2BJ6XQ5FK7U4H`) |
| `accelerator_ip_addresses` | `string[]` | Static anycast IP addresses |
| `listener_arns` | `map<string, string>` | Listener name to ARN |
| `endpoint_group_arns` | `map<string, string>` | `listener_name/group_name` to ARN |

## Related Components

- [AwsAlb](/docs/catalog/aws/alb) — common endpoint type for HTTP/HTTPS workloads
- [Network Load Balancer](/docs/catalog/aws/network-load-balancer) — common endpoint type for Layer 4 workloads
- [Elastic IP](/docs/catalog/aws/elastic-ip) — static IP endpoints
- [S3 Bucket](/docs/catalog/aws/s3-bucket) — stores flow logs
- [Route53 Zone](/docs/catalog/aws/route53-zone) — alias records for custom domains
