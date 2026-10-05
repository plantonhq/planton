# AWS ALB

Deploys an Application Load Balancer — the Layer-7 entry point that terminates HTTP/HTTPS and hands requests to the routing graph. The ALB itself carries no routing configuration by design: listeners (AwsLbListener) attach to it and own ports, TLS material, and default actions; listener rules (AwsLbListenerRule) own per-service routing; target groups (AwsLbTargetGroup) receive the traffic. This component owns what is truly load-balancer-wide — placement, security groups, and the HTTP behavior attributes.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Application Load Balancer** -- with its scheme, subnets, security groups, address family, and the configured HTTP behavior attributes (timeouts, HTTP/2, header handling, desync mitigation, WAF fail mode, zonal shift)
- **S3 log delivery configuration** -- set on the load balancer for whichever of the three streams (access, connection, health-check logs) has a bucket configured
- **WAF Web ACL association** -- created only when `webAclArn` is set, binding the REGIONAL-scope web ACL to the load balancer
- **Route53 alias A records** -- created only when DNS is enabled, one alias record per hostname pointing at the ALB's DNS name
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the load balancer

Listeners, rules, and target groups are **not** created here — attach AwsLbListener resources through this ALB's `load_balancer_arn` output; TLS certificates live on the listeners.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Subnets** -- at least two AwsSubnet resources in different Availability Zones, referenced by their `subnet_id` outputs (public for internet-facing, private for internal).
- **Security Groups** -- AwsSecurityGroup resources opening exactly the listener ports; without them AWS attaches the VPC's default group (fine for a first boot, wrong for production).
- **S3 Bucket / Route53 Zone** -- optional AwsS3Bucket and AwsRoute53Zone resources for log delivery and alias DNS.

### AWS Account

- **ELB permissions** -- the credentials used by the Provider Connection must have `elasticloadbalancing:CreateLoadBalancer`, `DescribeLoadBalancers`, `ModifyLoadBalancerAttributes`, and `DeleteLoadBalancer`, plus `ec2:Describe*` on the subnets.
- **Two-AZ minimum** -- AWS requires ALB subnets in at least two Availability Zones.
- **Log-delivery bucket policy** -- when enabling any log stream, the S3 bucket must grant the regional ELB log-delivery service write access; delivery fails silently otherwise.

## Deploy

### Console

Open the deployment store, find **AWS ALB**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Internet-Facing ALB** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAlb
metadata:
  name: api-gateway
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  subnets:
    - valueFrom:
        kind: AwsSubnet
        name: public-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: public-subnet-b
        fieldPath: status.outputs.subnet_id
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: alb-ingress
        fieldPath: status.outputs.security_group_id
  deleteProtectionEnabled: true
```

```shell
planton apply -f alb.yaml
```

This creates an internet-facing ALB across two zones with explicit security groups and deletion protection. Attach AwsLbListener resources to route traffic. An Infra Job tracks the provisioning in real time.

### InfraChart

When the ALB deploys alongside its network in one chart, wire the subnet and security group references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  subnets:
    - valueFrom:
        kind: AwsSubnet
        name: public-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: public-subnet-b
        fieldPath: status.outputs.subnet_id
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: alb-ingress
        fieldPath: status.outputs.security_group_id
```

The InfraPipeline resolves the dependency graph, deploys the subnets and security group first, then places the ALB on them.

## Key Configuration

These are the most important decisions when configuring an ALB. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Routing lives elsewhere** -- listeners own ports, TLS certificates, and default actions; rules own per-service routing; target groups receive traffic. The ALB spec is placement and behavior only.

**The scheme is a one-way door** -- `internal` decides internet-facing vs VPC-only and cannot change; AWS replaces the load balancer. Everything else — subnets, security groups, timeouts, header handling — updates in place.

**Security groups: explicit beats default** -- omitted groups fall back to the VPC's default security group. Production ALBs attach explicit groups opening exactly the listener ports.

**Idle timeout causes the classic 504** -- a response slower than the idle timeout (default 60s) times out at the ALB while the target still works. Raise it above the slowest legitimate response, and keep it below the targets' own keep-alive timeouts.

**Header handling matters behind proxies** -- an ALB behind CloudFront or a corporate gateway should usually preserve (not append to) the X-Forwarded-For chain; applications that route by hostname need the Host header preserved.

**Hardening is editable** -- desync mitigation ("defensive" by default; "monitor" to observe before enforcing "strictest"), invalid-header dropping, and the WAF fail-open/fail-closed call all update in place.

**WAF attaches here** -- `webAclArn` associates a REGIONAL-scope WAFv2 web ACL (same region, at most one per ALB) so every request is inspected before listener rules run; the fail-open toggle decides behavior while the WAF is unreachable.

**Deletion protection is recommended** -- deleting an ALB silently orphans every listener and rule attached to it. The wizard preselects protection ON.

**Capacity can be reserved for events** -- `minimumLoadBalancerCapacityUnits` pre-provisions LCUs for a dated traffic surge (the declarative replacement for pre-warming tickets). The reservation bills while set: size it from the event estimate, then remove the field to release it. `ipv4IpamPoolId` pairs with it for BYOIP addressing, keeping the ALB's public addresses inside ranges your clients have allowlisted.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnets[]` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `securityGroups[]` | `status.outputs.security_group_id` |
| **AwsS3Bucket** | `accessLogs.bucket` / `connectionLogs.bucket` / `healthCheckLogs.bucket` | `status.outputs.bucket_id` |
| **AwsRoute53Zone** | `dns.route53ZoneId` | `status.outputs.zone_id` |
| **AwsWafWebAcl** | `webAclArn` | `status.outputs.web_acl_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `load_balancer_arn` | ARN of the ALB | AwsLbListener resources attach through it |
| `load_balancer_name` | The ALB's name (truncated to AWS's 32-character limit) | Console URLs and CLI queries |
| `load_balancer_dns_name` | The AWS-assigned DNS name | CNAME targets and Route53 alias records |
| `load_balancer_hosted_zone_id` | Route53 hosted zone ID for the DNS name | Required when creating alias records manually |
| `arn_suffix` | The ARN's final segment | The CloudWatch LoadBalancer metric dimension — alarms, dashboards, request-count autoscaling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public HTTP entry point** -- an internet-facing ALB with explicit security groups; listeners add HTTPS with ACM certificates. Start from the **Internet-Facing ALB** preset.

**Internal service front** -- an internal ALB for service-to-service HTTP inside the VPC. Start from the **Internal Hardened ALB** preset.

**Event-day capacity reservation** -- pre-provision LCUs ahead of a dated traffic surge, then remove the field after the event to release the (billed) reservation. Start from the **Reserved Capacity for a Traffic Event** preset.

**Static IPs in front** -- put an AwsNlb with Elastic IPs in front and register this ALB in an `alb`-type AwsLbTargetGroup — static Layer-4 addresses, full Layer-7 routing.

## Works With

- [**AWS LB Listener**](/infra-catalog/aws-lb-listener) -- attaches to this ALB's `load_balancer_arn` output and owns ports, TLS certificates, and default actions.
- [**AWS LB Listener Rule**](/infra-catalog/aws-lb-listener-rule) -- attaches to listeners for path/host/header routing.
- [**AWS LB Target Group**](/infra-catalog/aws-lb-target-group) -- receives the routed traffic; an `alb`-type group also lets an NLB front this ALB.
- [**AWS NLB**](/infra-catalog/aws-nlb) -- fronts this ALB when static Layer-4 addresses are required, via an `alb`-type target group.
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- placement across at least two Availability Zones, referenced by `subnets`.
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- traffic filtering on the listener ports, referenced by `securityGroups`.
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- the log destinations, referenced per stream.
- [**AWS Route 53 Zone**](/infra-catalog/aws-route53-zone) -- the hosted zone for alias records, referenced by `dns.route53ZoneId`.
- [**AWS WAF Web ACL**](/infra-catalog/aws-waf-web-acl) -- a REGIONAL-scope web ACL inspecting requests in front of the listeners, referenced by `webAclArn`.
