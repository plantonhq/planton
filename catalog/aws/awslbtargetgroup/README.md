# Overview

The AwsLbTargetGroup API resource provisions an ELBv2 target group: the
routing destination that load balancer listeners and listener rules forward
traffic to, with health checks, stickiness, and connection-handling policy
attached.

## Why We Created This API Resource

The target group is the composition point of AWS load balancing. It has its
own lifecycle (it exists independently of any load balancer), its own ARN, and
it is referenced from many places at once: listener default actions, listener
rule forward actions, ECS services, and auto-scaling groups. Modeling it as a
first-class component -- instead of burying it inside a load balancer
definition -- lets you:

- **Deploy services without touching the load balancer**: an ECS service or a
  listener rule references the group's `target_group_arn` output, so shipping
  a new service never edits the shared ALB or NLB.
- **Run blue/green and canary rollouts**: two groups behind one weighted
  forward action shift traffic gradually; each group keeps its own health
  checks and drain behavior.
- **Serve both load balancer families with one kind**: the protocol decides
  the family (HTTP/HTTPS for ALB; TCP/UDP/TCP_UDP/TLS/QUIC/TCP_QUIC for NLB),
  exactly as AWS models ELBv2.

## Key Features

### Target Addressing

- **Four target types**: `instance` (EC2 instance IDs), `ip` (VPC-routable
  addresses -- the ECS awsvpc and Kubernetes pod-IP shape), `lambda` (direct
  function invocation), and `alb` (the NLB-in-front-of-ALB pattern).
- **Static target registration**: standalone instances, fixed IPs, a Lambda
  function, or an inner ALB register directly in the spec; dynamic registrars
  (ECS, ASG, Kubernetes controllers) keep managing their own targets.
- **IPv4 and IPv6 target groups**, with `ipv6` requiring a dualstack load
  balancer.
- **QUIC / HTTP/3 backends** (NLB): `QUIC` groups carry QUIC natively and
  `TCP_QUIC` groups serve QUIC with TCP fallback on one port; registrations
  carry the `quicServerId` that QUIC's connection-ID routing requires.

### Health and Traffic Policy

- **Full health-check control**: protocol (independent of the traffic
  protocol), port, path, thresholds, interval/timeout, and HTTP or gRPC
  response matchers.
- **Session stickiness**: `lb_cookie` and `app_cookie` for ALB;
  `source_ip`, `source_ip_dest_ip`, and `source_ip_dest_ip_proto` flow-hash
  affinity for NLB.
- **Per-target readiness routing** (ALB): `targetControlPort` enables ALB
  Target Optimizer, routing by each target's self-reported readiness.
- **Rollout-friendly draining**: deregistration delay, slow start (ALB),
  connection termination on drain (NLB), and unhealthy-target connection
  policy.
- **Group-level health policy**: DNS failover and unhealthy-state-routing
  thresholds that act on the group as a whole.

### Load Balancing Behavior

- **Algorithm selection** (ALB): `round_robin`,
  `least_outstanding_requests`, or `weighted_random` with optional automatic
  anomaly mitigation.
- **Cross-zone override**: keep, force, or disable cross-zone distribution
  per group, independent of the load balancer's own setting.
- **Client-address fidelity** (NLB): preserve the original client IP or carry
  connection metadata via Proxy Protocol v2.

## Benefits

- **Composability**: listeners, rules, and ECS services reference the group by
  ARN through `valueFrom`, so the architecture graph shows exactly which
  routes feed which backends.
- **Safe rollouts**: per-group health and drain policy makes weighted traffic
  shifting predictable.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `target_group_arn`: ARN of the target group (what listeners, rules, ECS services, and ASGs reference)
- `target_group_name`: friendly name of the group (metadata.name, truncated to AWS's 32-character limit when necessary)
- `arn_suffix`: ARN suffix used as the TargetGroup dimension in CloudWatch metrics

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
