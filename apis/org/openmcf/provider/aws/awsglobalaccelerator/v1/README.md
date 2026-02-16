# AwsGlobalAccelerator

The **AwsGlobalAccelerator** resource provides a standardized way to provision and manage AWS Global Accelerator through OpenMCF. It creates a Layer 4 anycast networking service that routes traffic through the AWS global backbone to optimal regional endpoints based on health, geography, and routing policies.

## What Gets Created

When you deploy an AwsGlobalAccelerator resource, OpenMCF provisions the complete resource hierarchy:

1. **Global Accelerator**: Two static anycast IPv4 addresses (or dual-stack) that serve as fixed entry points.
2. **Listeners**: Port and protocol configurations that accept inbound traffic on the accelerator.
3. **Endpoint Groups**: Regional sets of endpoints, each with its own health check and traffic dial configuration.
4. **Endpoints**: The actual AWS resources (ALB, NLB, EIP, EC2 instance) that receive traffic.
5. **Flow Logs** (optional): Traffic analysis logs delivered to an S3 bucket.

All three levels (listener → endpoint group → endpoint) are bundled into a single resource because an accelerator without listeners and endpoint groups is functionally useless — just static IPs doing nothing.

## When to Use

**Use AwsGlobalAccelerator when you need:**

- **Static anycast IPs** as fixed entry points for global applications (IP whitelisting, DNS-independent failover)
- **Layer 4 routing** for TCP or UDP traffic across multiple AWS regions
- **Sub-second failover** when a regional endpoint becomes unhealthy
- **Predictable low latency** by entering the AWS backbone at the nearest edge location
- **Traffic shifting** between regions for blue/green or canary deployments

**Use CloudFront instead when:**

- You need HTTP-aware features: caching, header manipulation, Lambda@Edge, WebSocket upgrade handling
- Your workload is purely HTTP/HTTPS and benefits from edge caching
- You need per-path or per-host routing at Layer 7

**Use Route53 latency-based routing instead when:**

- DNS-based failover with TTL-controlled propagation is acceptable
- You don't need static IP addresses
- Cost sensitivity outweighs the need for instant failover (Route53 failover depends on DNS TTLs)

## Prerequisites

- An active AWS account with permissions to create Global Accelerator resources
- At least one endpoint (ALB, NLB, Elastic IP, or EC2 instance) already deployed in a target region
- (Optional) An S3 bucket for flow log storage
- (Optional) BYOIP address pool registered with AWS if using custom IP addresses

## Spec Fields

### Core Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | `bool` | `true` | Whether the accelerator accepts traffic. Set to `false` to pause routing without destroying the resource. |
| `ipAddressType` | `string` | `"IPV4"` | IP address type. `"IPV4"` for two static IPv4 anycast addresses. `"DUAL_STACK"` for IPv4 + IPv6. |
| `ipAddresses` | `string[]` | `[]` (AWS-allocated) | BYOIP addresses to use instead of AWS-allocated IPs. Provide exactly 1 or 2 IPv4 addresses from a registered BYOIP pool. **ForceNew** — changing this destroys and recreates the accelerator. |

### Flow Logs (`flowLogs`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `flowLogs.enabled` | `bool` | `false` | Enable flow log delivery to S3 for traffic analysis. |
| `flowLogs.s3Bucket` | `string \| valueFrom` | — | S3 bucket name for flow log storage. Required when `enabled` is `true`. Supports `valueFrom` referencing an `AwsS3Bucket` resource. |
| `flowLogs.s3Prefix` | `string` | `""` | Key prefix for organizing logs within the bucket. Example: `"ga-logs/prod/"`. |

### Listeners (`listeners`)

At least one listener is required. Each listener defines a port/protocol entry point on the accelerator.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | — | **Required.** Unique identifier for this listener. Lowercase alphanumeric and hyphens, starting with a letter (max 63 chars). Used as key in output maps. |
| `protocol` | `string` | — | **Required.** `"TCP"` or `"UDP"`. Global Accelerator operates at Layer 4. |
| `clientAffinity` | `string` | `"NONE"` | `"NONE"` for stateless routing. `"SOURCE_IP"` to pin a client IP to the same endpoint (required for stateful protocols like gaming or WebSockets). |
| `portRanges` | `PortRange[]` | — | **Required.** 1–10 port ranges. Each has `fromPort` and `toPort` (inclusive, 1–65535). |
| `endpointGroups` | `EndpointGroup[]` | — | **Required.** At least one regional endpoint group. |

### Endpoint Groups (`listeners[].endpointGroups`)

Each endpoint group represents a set of endpoints in a single AWS region.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | — | **Required.** Unique identifier within the listener. Used in the `endpoint_group_arns` output map as `"listener_name/group_name"`. |
| `endpointGroupRegion` | `string` | Provider region | AWS region for this group (e.g., `"us-east-1"`). **ForceNew** — changing the region replaces the group. |
| `healthCheckPort` | `int32` | Listener port | Port for health checks. Use when the health check port differs from the traffic port. |
| `healthCheckProtocol` | `string` | `"TCP"` | `"TCP"`, `"HTTP"`, or `"HTTPS"`. TCP checks port reachability only; HTTP/HTTPS send GET requests to `healthCheckPath`. |
| `healthCheckPath` | `string` | — | Path for HTTP/HTTPS health checks (e.g., `"/health"`). Required when protocol is `HTTP` or `HTTPS`. Ignored for TCP. |
| `healthCheckIntervalSeconds` | `int32` | `30` | Seconds between health checks. AWS only supports **10** or **30** — no other values are accepted. |
| `thresholdCount` | `int32` | `3` | Consecutive checks that must pass (or fail) to change endpoint health status. Range: 1–10. |
| `trafficDialPercentage` | `double` | `100.0` | Percentage of traffic routed to this group (0.0–100.0). Use for regional traffic shifting. Set to `0` to drain a region. |
| `endpoints` | `Endpoint[]` | `[]` | Endpoints in this group. Optional — you can create the group first and add endpoints later. |
| `portOverrides` | `PortOverride[]` | `[]` | Remap listener ports to different endpoint ports. Max 10 overrides. |

### Endpoints (`listeners[].endpointGroups[].endpoints`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `endpointId` | `string \| valueFrom` | — | **Required.** Resource identifier: ALB ARN, NLB ARN, EIP allocation ID (`eipalloc-...`), or EC2 instance ID (`i-...`). Supports `valueFrom` for cross-resource references. |
| `weight` | `int32` | `128` | Relative traffic weight (0–255). Higher means more traffic. Set to `0` to stop routing without removing. |
| `clientIpPreservationEnabled` | `bool` | `false` | Preserve original client IP. Supported for ALB and EC2 endpoints only. NLB and EIP always preserve client IP. |

### Port Overrides (`listeners[].endpointGroups[].portOverrides`)

| Field | Type | Description |
|-------|------|-------------|
| `listenerPort` | `int32` | The listener port to remap (must be within one of the listener's port ranges). |
| `endpointPort` | `int32` | The port that the endpoint actually serves on. |

## Stack Outputs

After provisioning, the resource exposes the following outputs:

| Output | Type | Description |
|--------|------|-------------|
| `accelerator_arn` | `string` | ARN of the Global Accelerator. Used in IAM policies. |
| `accelerator_dns_name` | `string` | DNS name assigned by AWS (e.g., `a1234567890abcdef.awsglobalaccelerator.com`). Create Route53 alias records pointing here. |
| `accelerator_dual_stack_dns_name` | `string` | Dual-stack DNS name. Populated only when `ipAddressType` is `DUAL_STACK`. |
| `accelerator_hosted_zone_id` | `string` | Route53 hosted zone ID for the accelerator's DNS name. Always `Z2BJ6XQ5FK7U4H` for Global Accelerator. |
| `accelerator_ip_addresses` | `string[]` | Static anycast IP addresses (typically two). These never change for the lifetime of the accelerator. |
| `listener_arns` | `map<string, string>` | Map of listener name → listener ARN. Keys correspond to `spec.listeners[].name`. |
| `endpoint_group_arns` | `map<string, string>` | Map of `"listener_name/group_name"` → endpoint group ARN. |

## Quick Start

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlobalAccelerator
metadata:
  name: my-accelerator
  labels:
    openmcf.org/provisioner: pulumi
spec:
  listeners:
    - name: tcp-443
      protocol: TCP
      portRanges:
        - fromPort: 443
          toPort: 443
      endpointGroups:
        - name: primary
          endpointGroupRegion: us-east-1
          healthCheckProtocol: TCP
          endpoints:
            - endpointId: arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-alb/1234567890abcdef
              weight: 128
```

Apply the manifest:

```shell
openmcf pulumi up --manifest accelerator.yaml --stack <stack-name>
```

## How It Works

1. **Creates the accelerator**: Provisions a Global Accelerator with two static anycast IPs that are advertised from all AWS edge locations worldwide.
2. **Configures listeners**: Sets up port/protocol combinations on the accelerator that accept inbound traffic.
3. **Creates endpoint groups**: Establishes regional groups with health check and traffic dial configuration.
4. **Registers endpoints**: Associates ALBs, NLBs, EIPs, or EC2 instances as traffic destinations within each group.
5. **Enables flow logs** (optional): Configures S3 delivery for traffic analysis.

Traffic flows as: **Client → nearest AWS edge location → AWS global backbone → regional endpoint group → healthy endpoint**.

## References

- [AWS Global Accelerator Documentation](https://docs.aws.amazon.com/global-accelerator/latest/dg/what-is-global-accelerator.html)
- [Global Accelerator FAQ](https://aws.amazon.com/global-accelerator/faqs/)
- [Global Accelerator Pricing](https://aws.amazon.com/global-accelerator/pricing/)
- [AWS Edge Locations](https://aws.amazon.com/cloudfront/features/#Amazon_CloudFront_Infrastructure)
