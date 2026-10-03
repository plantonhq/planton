# AWS Global Accelerator

Deploys a Global Accelerator with static anycast IP addresses, configurable listeners, regional endpoint groups, and endpoint routing. The accelerator routes traffic through the AWS global network to optimal endpoints based on health, geography, and traffic dial percentages -- Layer 4 (TCP/UDP), not HTTP-aware like CloudFront, with two anycast addresses that never change for the accelerator's lifetime. One operational expectation: every accelerator, listener, or endpoint-group change waits for the accelerator to return to the DEPLOYED state, so applies take minutes, not seconds.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Global Accelerator** -- a networking resource with two static anycast IPv4 addresses (or dual-stack) that serve as fixed entry points, with configurable flow log delivery to S3
- **Listeners** -- one per entry in `listeners`, each accepting traffic on specified port ranges and protocol (TCP or UDP), with optional SOURCE_IP client affinity
- **Endpoint Groups** -- one per listener entry, each targeting a specific AWS region with configurable health checks, traffic dial percentages, and port overrides
- **Endpoints** -- registered within endpoint groups, supporting ALBs, NLBs, Elastic IPs, and EC2 instances as traffic targets with configurable weights and client IP preservation
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **Endpoint resources** -- at least one ALB, NLB, Elastic IP, or EC2 instance to register as a traffic target. Provide the ARN or ID directly or reference another Infra Component via ValueFromRef.
- **An S3 bucket** (optional) for flow log storage when traffic analysis is needed. Provide the bucket name directly or reference an AwsS3Bucket Infra Component.

## Deploy

### Console

Open the deployment store, find **AWS Global Accelerator**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic TCP Accelerator** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsGlobalAccelerator
metadata:
  name: web-accelerator
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  listeners:
    - name: https
      protocol: TCP
      portRanges:
        - fromPort: 443
          toPort: 443
      endpointGroups:
        - name: primary
          endpoints:
            - endpointId:
                value: "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-alb/abc123"
```

```shell
planton apply -f global-accelerator.yaml
```

This creates a Global Accelerator with a single TCP listener on port 443 routing to one ALB endpoint. Flow logs are disabled, health checks use TCP defaults, and traffic dial is set to 100%. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the accelerator to an ALB and S3 bucket deployed in the same InfraPipeline:

```yaml
spec:
  flowLogs:
    enabled: true
    s3Bucket:
      valueFrom:
        kind: AwsS3Bucket
        name: ga-flow-logs
        fieldPath: status.outputs.bucket_id
```

The InfraPipeline resolves the dependency graph, deploys the S3 bucket first, then provisions the Global Accelerator with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Global Accelerator. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**IP address type** -- `ipAddressType` defaults to `IPV4` for two static anycast IPv4 addresses. Set to `DUAL_STACK` for IPv4 + IPv6 when serving clients on IPv6 networks. BYOIP addresses can replace AWS-allocated IPs for IP portability.

**Client affinity** -- Set `clientAffinity` to `SOURCE_IP` on a listener to route all requests from the same client IP to the same endpoint. Required for stateful protocols (gaming, WebSocket, long-lived TCP sessions). Default `NONE` distributes requests across endpoints.

**Traffic dial percentage** -- `trafficDialPercentage` on each endpoint group controls what fraction of traffic reaches that region (0.0-100.0). Use values below 100 for gradual traffic shifting between regions during blue/green or canary deployments. Set to 0 to drain a region without removing endpoints.

**Health check configuration** -- Each endpoint group supports TCP, HTTP, or HTTPS health checks. AWS only supports intervals of exactly 10 or 30 seconds. HTTP/HTTPS checks validate application readiness via a path, while TCP checks only verify port reachability.

**Cross-account endpoints** -- An endpoint that lives in another AWS account needs a Global Accelerator cross-account attachment: create the attachment in the endpoint-owning account (naming this accelerator's account as a principal) and supply its ARN in the endpoint's `attachmentArn`. Same-account endpoints -- the common case -- leave it empty.

**Optional dials stay AWS-owned when unset** -- Fields like `trafficDialPercentage` (default 100) and the health-check port (defaults to the listener's port) are only written when you take a position. Endpoint `weight` is the one exception: AWS documents 128 as the default but the provider would transmit 0 (the "route no traffic" value) for an omitted weight, so the modules materialize 128 themselves. For `weight` and `trafficDialPercentage`, 0 is a real value that drains the endpoint or region -- distinct from leaving the field unset.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsS3Bucket** (optional) | `flowLogs.s3Bucket` | `status.outputs.bucket_id` |
| **AwsAlb** (optional) | `listeners[].endpointGroups[].endpoints[].endpointId` | `status.outputs.load_balancer_arn` |
| **AwsNlb** (optional) | `listeners[].endpointGroups[].endpoints[].endpointId` | `status.outputs.load_balancer_arn` |
| **AwsElasticIp** (optional) | `listeners[].endpointGroups[].endpoints[].endpointId` | `status.outputs.allocation_id` |
| **AwsEc2Instance** (optional) | `listeners[].endpointGroups[].endpoints[].endpointId` | `status.outputs.instance_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `accelerator_arn` | ARN of the Global Accelerator | IAM policies, cross-service permissions |
| `accelerator_dns_name` | DNS name for client connections | Route53 alias records, client configuration |
| `accelerator_dual_stack_dns_name` | Dual-stack DNS name (IPv4 + IPv6) | IPv6-capable client configuration |
| `accelerator_hosted_zone_id` | Route53 hosted zone ID for the accelerator | Route53 alias record creation |
| `accelerator_ip_addresses` | Static anycast IP addresses | DNS pinning, firewall allowlists, client configuration |
| `listener_arns` | Map of listener name to listener ARN | Cross-resource listener references |
| `endpoint_group_arns` | Map of listener/group name to endpoint group ARN | Cross-resource endpoint group references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic TCP accelerator** -- Single-region accelerator with TCP on port 443 routing to one ALB endpoint. Provides static anycast IPs and BGP-based failover. Start from the **Basic TCP Accelerator** preset.

**Multi-region production** -- Two regional endpoint groups (e.g., us-east-1 and eu-west-1) with HTTP health checks, flow logs, and weighted traffic distribution. Automatic regional failover within seconds. Start from the **Multi-Region Production Accelerator** preset.

**Gaming UDP accelerator** -- UDP protocol with SOURCE_IP affinity and a port range for game servers. Routes players to the nearest healthy game server via Elastic IP endpoints. Start from the **Gaming UDP Accelerator** preset.

## Works With

- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- provides the storage bucket for flow log delivery
- [**AWS ALB**](/infra-catalog/aws-alb) -- an Application Load Balancer registered as an endpoint traffic target
- [**AWS NLB**](/infra-catalog/aws-nlb) -- a Network Load Balancer registered as an endpoint traffic target
- [**AWS Elastic IP**](/infra-catalog/aws-elastic-ip) -- an Elastic IP allocation registered as an endpoint traffic target
- [**AWS EC2 Instance**](/infra-catalog/aws-ec2-instance) -- an EC2 instance registered as an endpoint traffic target