# AWS LB Target Group

Deploys an ELBv2 target group — the routing destination that listeners and listener rules forward traffic to, and the composition point of AWS load balancing. A target group has its own lifecycle and its own ARN: one group can receive traffic from several listeners, one listener can spread traffic across several weighted groups, ECS services take a group's ARN as their deployment target, and auto-scaling groups register instances into it. The same kind serves both Application and Network Load Balancers, exactly as AWS models it — the protocol decides the family, and the family decides which tuning fields apply.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Target Group** -- with its target type, port/protocol, health check, stickiness, and family-specific traffic attributes
- **Static target registrations** -- only for rows listed in `targets`; dynamic registrars (ECS, auto-scaling, Kubernetes controllers) manage their own membership
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the group

The load balancer and listeners are **not** created here — an AwsLbListener forwards to this group's `target_group_arn` output.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **VPC** -- the AwsVpc the targets live in, referenced by its `vpc_id` output (required for every target type except lambda).
- **Targets** -- optional AwsEc2Instance resources for static registrations, referenced by their `instance_id` outputs.

### AWS Account

- **ELB permissions** -- the credentials used by the Provider Connection must have `elasticloadbalancing:CreateTargetGroup`, `ModifyTargetGroup`, `ModifyTargetGroupAttributes`, `RegisterTargets`, `DeregisterTargets`, and `DeleteTargetGroup`.
- **Same-region rule** -- the group, its VPC, and any load balancer forwarding to it must share one region.

## Deploy

### Console

Open the deployment store, find **AWS LB Target Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **ECS Service HTTP** preset in the [Presets](#presets) tab for the most common container pattern.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLbTargetGroup
metadata:
  name: web-servers
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: main-vpc
      fieldPath: status.outputs.vpc_id
  targetType: ip
  port: 8080
  protocol: HTTP
  healthCheck:
    path: /healthz
    matcher: "200"
```

```shell
planton apply -f target-group.yaml
```

This creates an HTTP target group for IP targets (the shape ECS awsvpc services register into) with a real readiness probe. An Infra Job tracks the provisioning in real time.

### InfraChart

When the group deploys alongside its VPC in one chart, wire the network reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: main-vpc
      fieldPath: status.outputs.vpc_id
  targetType: ip
  port: 8080
  protocol: HTTP
  healthCheck:
    path: /healthz
    matcher: "200"
```

The InfraPipeline resolves the dependency graph, deploys the VPC first, then creates the target group in it.

## Key Configuration

These are the most important decisions when configuring a target group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The identity is create-only** -- name, VPC, target type, address family, port, protocol, and protocol version are fixed at creation. Changing any replaces the group, and the IaC engine re-creates dependent listener attachments.

**The target type shapes everything** -- `instance` (EC2 fleets), `ip` (ECS awsvpc tasks, pod IPs, peered/on-premises addresses), `lambda` (the load balancer invokes the function — no port, protocol, or VPC), or `alb` (the NLB-in-front-of-ALB pattern combining static Layer-4 IPs with Layer-7 routing).

**The protocol decides the family** -- HTTP/HTTPS groups attach to ALBs and unlock request-level tuning (routing algorithm, anomaly mitigation, slow start, cookie stickiness, Target Optimizer readiness routing); TCP/UDP/TCP_UDP/TLS/QUIC/TCP_QUIC groups attach to NLBs and unlock connection-level tuning (client-IP preservation, Proxy Protocol v2, connection termination, flow-hash stickiness, QUIC server-ID registrations).

**Health checks probe independently of traffic** -- an NLB TCP group can (and usually should) run HTTP probes against a readiness endpoint; a TCP probe only proves the port is open. AWS applies protocol-appropriate defaults when the block is omitted.

**Deregistration delay gates deploys** -- a draining target keeps serving in-flight requests for this window (default 300s); deploys cannot finish faster than it. Size it to the longest legitimate request.

**Static targets are usually empty** -- ECS services, auto-scaling groups, and Kubernetes controllers register targets dynamically. Registrations are folded into this kind (not a separate resource) because a registration is pure glue with no referenceable identity of its own.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsVpc** | `vpcId` | `status.outputs.vpc_id` |
| **AwsEc2Instance** | `targets[].targetId` | `status.outputs.instance_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `target_group_arn` | ARN of the target group | AwsLbListener forward actions, ECS service load-balancer bindings, auto-scaling group attachments |
| `target_group_name` | The group's name (truncated to AWS's 32-character limit) | Console URLs and CLI queries |
| `arn_suffix` | The ARN's final segment | The CloudWatch TargetGroup metric dimension |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**ECS service backend** -- an ip-type HTTP group ECS registers task IPs into. Start from the **ECS Service HTTP** preset.

**NLB TCP passthrough** -- a TCP group behind an NLB for raw connection forwarding. Start from the **NLB TCP Pass-Through** preset.

**Lambda behind an ALB** -- a lambda group giving a function HTTP routing, WAF, and OIDC auth without API Gateway. Start from the **Lambda Function Target** preset.

**HTTP/3 at the edge** -- a TCP_QUIC group behind an NLB serving QUIC natively while the same group takes clients that fall back to TCP on port 443; static registrations carry the `quicServerId` connection-ID routing requires. Start from the **QUIC / HTTP/3 Pass-Through** preset.

## Works With

- [**AWS LB Listener**](/infra-catalog/aws-lb-listener) -- forwards traffic here via its default actions, referencing `target_group_arn`.
- [**AWS LB Listener Rule**](/infra-catalog/aws-lb-listener-rule) -- forwards matched requests here via rule actions.
- [**AWS ALB**](/infra-catalog/aws-alb) -- the Layer-7 load balancer whose listeners deliver HTTP/HTTPS traffic.
- [**AWS NLB**](/infra-catalog/aws-nlb) -- the Layer-4 load balancer whose listeners deliver TCP/UDP/TLS/QUIC traffic.
- [**AWS VPC**](/infra-catalog/aws-vpc) -- the network the targets live in, referenced by `vpcId`.
- [**AWS EC2 Instance**](/infra-catalog/aws-ec2-instance) -- statically registered instance targets, referenced per row.
