# AWS NLB

Deploys a Network Load Balancer — the Layer-4 entry point for TCP, UDP, and TLS traffic, with static IP addresses per zone and millions of requests per second of headroom. The NLB itself carries no routing configuration by design: listeners (AwsLbListener) attach to it and own ports, protocols, and TLS material; target groups (AwsLbTargetGroup) receive the connections. This component owns what is truly load-balancer-wide — node placement with optional static IPs, security groups, and traffic distribution behavior.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Network Load Balancer** -- one node per subnet mapping, optionally pinned to an Elastic IP (internet-facing) or a specific private IPv4 address (internal), with the configured cross-zone, DNS routing, zonal-shift, and PrivateLink-enforcement attributes
- **S3 access-log configuration** -- set on the load balancer when the access-logs bucket is configured; captures TLS-listener traffic only (an AWS limitation)
- **Route53 alias A records** -- created only when DNS is enabled, one alias record per hostname pointing at the NLB's DNS name
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the load balancer

Listeners and target groups are **not** created here — attach AwsLbListener resources through this NLB's `load_balancer_arn` output and AwsLbTargetGroup resources as their destinations. Listener rules do not apply to NLBs; routing is purely by port and protocol.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Subnets** -- one AwsSubnet per zone the NLB serves, referenced by their `subnet_id` outputs (public subnets for internet-facing, private for internal).
- **Elastic IPs** -- optional AwsElasticIp resources referenced by their `allocation_id` outputs, for static public addresses (internet-facing only).
- **Security Groups / S3 Bucket / Route53 Zone** -- optional AwsSecurityGroup, AwsS3Bucket, and AwsRoute53Zone resources for filtering, access logs, and alias DNS.

### AWS Account

- **ELB permissions** -- the credentials used by the Provider Connection must have `elasticloadbalancing:CreateLoadBalancer`, `DescribeLoadBalancers`, `ModifyLoadBalancerAttributes`, and `DeleteLoadBalancer`, plus `ec2:Describe*` on the mapped subnets.
- **Subnet capacity** -- the NLB needs at least 8 free IP addresses in each mapped subnet.
- **Log-delivery bucket policy** -- when enabling access logs, the S3 bucket must grant the regional ELB log-delivery service write access; delivery fails silently otherwise.

## Deploy

### Console

Open the deployment store, find **AWS NLB**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Static-IP Internet-Facing NLB** preset in the [Presets](#presets) tab for the headline use case: public IPs that never change.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsNlb
metadata:
  name: edge-nlb
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  subnetMappings:
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: public-subnet-a
          fieldPath: status.outputs.subnet_id
      allocationId:
        valueFrom:
          kind: AwsElasticIp
          name: edge-ip-a
          fieldPath: status.outputs.allocation_id
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: public-subnet-b
          fieldPath: status.outputs.subnet_id
      allocationId:
        valueFrom:
          kind: AwsElasticIp
          name: edge-ip-b
          fieldPath: status.outputs.allocation_id
  deleteProtectionEnabled: true
```

```shell
planton apply -f nlb.yaml
```

This creates an internet-facing NLB with a static public IP in each of two zones. An Infra Job tracks the provisioning in real time.

### InfraChart

When the NLB deploys alongside its network in one chart, wire the subnet and Elastic IP references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  subnetMappings:
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: public-subnet-a
          fieldPath: status.outputs.subnet_id
      allocationId:
        valueFrom:
          kind: AwsElasticIp
          name: edge-ip-a
          fieldPath: status.outputs.allocation_id
```

The InfraPipeline resolves the dependency graph, deploys the subnets and Elastic IPs first, then places the NLB nodes on them.

## Key Configuration

These are the most important decisions when configuring an NLB. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The scheme is a one-way door** -- `internal` decides internet-facing vs VPC-only and cannot change; AWS replaces the load balancer, issuing a new DNS name and orphaning Elastic IP associations. Subnet mappings are replace-on-change for NLBs too.

**Subnet mappings are the placement** -- one mapping per Availability Zone, each placing one NLB node. Internet-facing mappings can pin an Elastic IP (the static-IP story: firewall allowlists, DNS pinning); internal mappings can pin a private IPv4 address inside the subnet's CIDR; dualstack mappings can additionally pin an `ipv6Address` from the subnet's IPv6 range. AWS recommends two or more zones. Modifying an existing mapping replaces the load balancer -- pure additions do not.

**Capacity and connection headroom** -- `minimumLoadBalancerCapacityUnits` reserves LCU capacity ahead of a known traffic event (billed whether used or not); `secondaryIpsAutoAssignedPerSubnet` (0-7) widens each node's source-port budget for very high connection counts to a single target, and decreasing it later replaces the load balancer. For UDP on a dualstack NLB, set `enablePrefixForIpv6SourceNat: "on"` -- prefix-based IPv6 source NAT is a prerequisite for IPv6 UDP flow handling.

**Security groups are optional but permanent** -- without them the NLB accepts all traffic on its listener ports (filtering happens at the targets); once attached, at least one group must remain forever, and a group-less NLB can never gain them. The PrivateLink enforcement setting decides whether inbound rules also filter traffic arriving through VPC endpoints.

**Cross-zone load balancing is off by default** -- each zone's node serves its own zone's targets (unlike ALB, where cross-zone is always on). Enabling it evens out uneven target distributions at the cost of inter-AZ data transfer. The DNS client routing policy is the complementary knob: zone affinity keeps clients in their own zone's node.

**Deletion protection is recommended** -- deleting an NLB silently orphans every listener attached to it, and pinned Elastic IPs start billing as unattached. The wizard preselects protection ON.

**Access logs capture TLS traffic only** -- an AWS limitation on NLB; plain TCP/UDP flows never appear. Pair with VPC Flow Logs for flow-level visibility.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetMappings[].subnetId` | `status.outputs.subnet_id` |
| **AwsElasticIp** | `subnetMappings[].allocationId` | `status.outputs.allocation_id` |
| **AwsSecurityGroup** | `securityGroups[]` | `status.outputs.security_group_id` |
| **AwsS3Bucket** | `accessLogs.bucket` | `status.outputs.bucket_id` |
| **AwsRoute53Zone** | `dns.route53ZoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `load_balancer_arn` | ARN of the NLB | AwsLbListener resources attach through it; IAM policies and CloudWatch alarms take it |
| `load_balancer_name` | The NLB's name (truncated to AWS's 32-character limit) | Console URLs and CLI queries |
| `load_balancer_dns_name` | The AWS-assigned DNS name | CNAME targets and Route53 alias records |
| `load_balancer_hosted_zone_id` | Route53 hosted zone ID for the DNS name | Required when creating alias records manually |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Static-IP internet entry point** -- an internet-facing NLB with an Elastic IP per zone; partners allowlist the fixed addresses. Start from the **Static-IP Internet-Facing NLB** preset.

**Internal service front** -- an internal NLB fronting TCP/gRPC services for other VPC workloads. Start from the **Internal NLB** preset.

**PrivateLink service backend** -- an internal NLB with security groups and PrivateLink enforcement, backing a VPC endpoint service consumed by other accounts. Start from the **PrivateLink-Hardened NLB** preset.

## Works With

- [**AWS LB Listener**](/infra-catalog/aws-lb-listener) -- attaches to this NLB's `load_balancer_arn` output and owns ports, protocols (TCP/UDP/TCP_UDP/TLS), and TLS certificates.
- [**AWS LB Target Group**](/infra-catalog/aws-lb-target-group) -- receives the connections the listeners forward.
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- the zone placement, referenced per subnet mapping.
- [**AWS Elastic IP**](/infra-catalog/aws-elastic-ip) -- static public addresses, referenced per internet-facing mapping.
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- optional inbound filtering, referenced by `securityGroups`.
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- the access-log destination, referenced by `accessLogs.bucket`.
- [**AWS Route 53 Zone**](/infra-catalog/aws-route53-zone) -- the hosted zone for alias records, referenced by `dns.route53ZoneId`.
