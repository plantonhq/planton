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
  region: us-east-1
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

This creates a Global Accelerator with a TCP listener on port 443 and one endpoint group in the provider's default region. No endpoints are registered yet — add them to the `endpoints` array or register them after deployment.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the global accelerator will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `listeners` | `object[]` | Listener definitions. Each defines a protocol, port ranges, and endpoint groups. | Minimum 1 item |
| `listeners[].name` | `string` | Unique name for the listener. Used as key in output maps. | Lowercase alphanumeric and hyphens, starts with letter, max 63 chars |
| `listeners[].protocol` | `string` | Layer 4 protocol: `TCP` or `UDP`. | Required |
| `listeners[].portRanges` | `object[]` | Port ranges the listener accepts traffic on. | Minimum 1, maximum 10 |
| `listeners[].portRanges[].fromPort` | `int` | First port in range (inclusive). | 1–65535 |
| `listeners[].portRanges[].toPort` | `int` | Last port in range (inclusive). | 1–65535 |
| `listeners[].endpointGroups` | `object[]` | Regional endpoint group definitions. | Minimum 1 |
| `listeners[].endpointGroups[].name` | `string` | Unique name within the listener. Used in composite output key. | Lowercase alphanumeric and hyphens, starts with letter, max 63 chars |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | `bool` | `true` | Whether the accelerator accepts traffic. Set `false` to disable without destroying. |
| `ipAddressType` | `string` | `IPV4` | `IPV4` or `DUAL_STACK` (IPv4 + IPv6). |
| `ipAddresses` | `string[]` | `[]` | BYOIP addresses (max 2). ForceNew. Leave empty for AWS-allocated IPs. |
| `flowLogs.enabled` | `bool` | `false` | Enable flow log delivery to S3. |
| `flowLogs.s3Bucket` | `string` | — | S3 bucket for flow logs. Required when `flowLogs.enabled` is `true`. Can reference AwsS3Bucket via `valueFrom`. |
| `flowLogs.s3Prefix` | `string` | `""` | S3 key prefix for flow logs. |
| `listeners[].clientAffinity` | `string` | `NONE` | `NONE` or `SOURCE_IP`. Use `SOURCE_IP` for stateful protocols. |
| `listeners[].endpointGroups[].endpointGroupRegion` | `string` | Provider region | AWS region for the endpoint group. ForceNew. |
| `listeners[].endpointGroups[].healthCheckPort` | `int` | Listener port | Port for health checks. |
| `listeners[].endpointGroups[].healthCheckProtocol` | `string` | `TCP` | `TCP`, `HTTP`, or `HTTPS`. |
| `listeners[].endpointGroups[].healthCheckPath` | `string` | — | Path for HTTP/HTTPS health checks. Required when protocol is `HTTP` or `HTTPS`. |
| `listeners[].endpointGroups[].healthCheckIntervalSeconds` | `int` | `30` | Health check interval. Must be exactly `10` or `30` (AWS constraint). |
| `listeners[].endpointGroups[].thresholdCount` | `int` | `3` | Consecutive checks to change health status. Range: 1–10. |
| `listeners[].endpointGroups[].trafficDialPercentage` | `float` | `100.0` | Percentage of traffic to route to this group. 0.0–100.0. Set to 0 to drain a region. |
| `listeners[].endpointGroups[].endpoints` | `object[]` | `[]` | Endpoints to register. Can be added later. |
| `listeners[].endpointGroups[].endpoints[].endpointId` | `string` | — | ALB ARN, NLB ARN, EIP allocation ID, or EC2 instance ID. Can reference via `valueFrom`. |
| `listeners[].endpointGroups[].endpoints[].weight` | `int` | `128` | Relative traffic weight. 0–255. Set 0 to stop traffic without removing. |
| `listeners[].endpointGroups[].endpoints[].clientIpPreservationEnabled` | `bool` | `false` | Preserve original client IP. Supported for ALB and EC2 endpoints. |
| `listeners[].endpointGroups[].portOverrides` | `object[]` | `[]` | Remap listener ports to different endpoint ports. Maximum 10. |
| `listeners[].endpointGroups[].portOverrides[].listenerPort` | `int` | — | Source listener port. 1–65535. |
| `listeners[].endpointGroups[].portOverrides[].endpointPort` | `int` | — | Destination endpoint port. 1–65535. |

## Examples

### Single-Region TCP Accelerator

Route HTTPS traffic to an ALB through the AWS global network:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlobalAccelerator
metadata:
  name: web-ga
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsGlobalAccelerator.web-ga
spec:
  region: us-east-1
  listeners:
    - name: https
      protocol: TCP
      portRanges:
        - fromPort: 443
          toPort: 443
      endpointGroups:
        - name: us-east-1
          endpointGroupRegion: us-east-1
          healthCheckProtocol: HTTP
          healthCheckPath: /health
          endpoints:
            - endpointId: arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-alb/1234567890abcdef
              weight: 128
              clientIpPreservationEnabled: true
```

### Multi-Region with Traffic Shifting

Route traffic across two regions with a 70/30 split for gradual regional migration:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlobalAccelerator
metadata:
  name: global-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsGlobalAccelerator.global-api
spec:
  region: us-east-1
  ipAddressType: DUAL_STACK
  flowLogs:
    enabled: true
    s3Bucket: my-ga-flow-logs
    s3Prefix: global-api/
  listeners:
    - name: https
      protocol: TCP
      portRanges:
        - fromPort: 443
          toPort: 443
      endpointGroups:
        - name: us-east-1
          endpointGroupRegion: us-east-1
          healthCheckProtocol: HTTP
          healthCheckPath: /health
          healthCheckIntervalSeconds: 10
          thresholdCount: 5
          trafficDialPercentage: 70.0
          endpoints:
            - endpointId: arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/us-alb/1111111111111111
              weight: 200
              clientIpPreservationEnabled: true
        - name: eu-west-1
          endpointGroupRegion: eu-west-1
          healthCheckProtocol: HTTP
          healthCheckPath: /health
          healthCheckIntervalSeconds: 10
          thresholdCount: 5
          trafficDialPercentage: 30.0
          endpoints:
            - endpointId: arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/app/eu-alb/2222222222222222
              weight: 200
              clientIpPreservationEnabled: true
```

### Gaming UDP with Client Affinity

UDP accelerator for a real-time multiplayer game with source IP stickiness:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlobalAccelerator
metadata:
  name: game-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsGlobalAccelerator.game-server
spec:
  region: us-west-2
  listeners:
    - name: game-udp
      protocol: UDP
      clientAffinity: SOURCE_IP
      portRanges:
        - fromPort: 7000
          toPort: 8000
      endpointGroups:
        - name: us-west-2
          endpointGroupRegion: us-west-2
          endpoints:
            - endpointId: eipalloc-0123456789abcdef0
              weight: 128
            - endpointId: eipalloc-fedcba9876543210f
              weight: 128
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding ARNs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlobalAccelerator
metadata:
  name: ref-ga
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsGlobalAccelerator.ref-ga
spec:
  region: us-east-1
  flowLogs:
    enabled: true
    s3Bucket:
      valueFrom:
        kind: AwsS3Bucket
        name: ga-logs
        field: status.outputs.bucket_name
  listeners:
    - name: https
      protocol: TCP
      portRanges:
        - fromPort: 443
          toPort: 443
      endpointGroups:
        - name: primary
          healthCheckProtocol: HTTP
          healthCheckPath: /health
          endpoints:
            - endpointId:
                valueFrom:
                  kind: AwsAlb
                  name: my-alb
                  field: status.outputs.load_balancer_arn
              weight: 200
              clientIpPreservationEnabled: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `accelerator_arn` | `string` | ARN of the Global Accelerator |
| `accelerator_dns_name` | `string` | Anycast DNS name (e.g., `a1234567890abcdef.awsglobalaccelerator.com`) |
| `accelerator_dual_stack_dns_name` | `string` | IPv4+IPv6 DNS name. Only populated when `ipAddressType` is `DUAL_STACK`. |
| `accelerator_hosted_zone_id` | `string` | Route53 hosted zone ID for alias records (always `Z2BJ6XQ5FK7U4H`) |
| `accelerator_ip_addresses` | `string[]` | Static anycast IP addresses assigned to the accelerator |
| `listener_arns` | `map<string, string>` | Map of listener name to listener ARN |
| `endpoint_group_arns` | `map<string, string>` | Map of `listener_name/group_name` to endpoint group ARN |

## Related Components

- [AwsAlb](/docs/catalog/aws/alb) — common endpoint type for HTTP/HTTPS workloads behind the accelerator
- [AwsNetworkLoadBalancer](/docs/catalog/aws/network-load-balancer) — common endpoint type for Layer 4 workloads
- [AwsElasticIp](/docs/catalog/aws/elastic-ip) — provides static IP endpoints for direct server routing
- [AwsS3Bucket](/docs/catalog/aws/s3-bucket) — stores flow logs when flow log delivery is enabled
- [AwsRoute53Zone](/docs/catalog/aws/route53-zone) — create alias records pointing custom domains to the accelerator DNS name
