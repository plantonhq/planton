# Overview

The AwsAlb API resource provisions an AWS Application Load Balancer: the
Layer-7 entry point that terminates HTTP/HTTPS and hands requests to the
routing graph.

## Why We Created This API Resource

The load balancer carries no routing configuration -- that is deliberate. In
AWS's own model, an ELBv2 load balancer is shared, slow-changing
infrastructure, while ports, certificates, and per-service routes churn with
every deployment. Planton mirrors that split with four composable kinds:

- **`AwsAlb`** (this component) owns what is truly load-balancer-wide:
  placement (two or more subnets, internal vs. internet-facing), security
  groups, and the HTTP behavior knobs AWS models as load balancer attributes.
- **`AwsLbListener`** attaches to the ALB by ARN and owns a port, its TLS
  material, and the default action.
- **`AwsLbListenerRule`** attaches to a listener and owns one service's
  routing conditions.
- **`AwsLbTargetGroup`** receives the traffic.

Deploying a new service means creating a target group and a rule -- never
editing the shared ALB. Rotating a certificate edits one listener. The ALB
manifest itself only changes when placement or HTTP-wide behavior changes,
which keeps the blast radius of every edit exactly as small as the intent
behind it.

## Key Features

### Placement and Security

- **Multi-AZ by construction**: at least two subnets in different
  Availability Zones are required -- the AWS minimum, enforced at the API
  layer rather than discovered at deploy time.
- **Scheme selection**: internet-facing (default) or internal
  (`internal: true`); changing the scheme replaces the load balancer, so the
  spec documents it as immutable.
- **Security groups by reference**: attach `AwsSecurityGroup` outputs (or
  literal IDs) that open exactly the listener ports.
- **IP address family**: `ipv4`, `dualstack`, or
  `dualstack-without-public-ipv4` to serve IPv6 clients without paying for
  public IPv4.
- **BYOIP addressing**: `ipv4IpamPoolId` draws the ALB's public IPv4
  addresses from your own VPC IPAM pool instead of AWS's ranges -- for
  clients that allowlist source ranges.
- **Reserved capacity**: `minimumLoadBalancerCapacityUnits` pre-provisions
  LCUs for a known traffic event instead of waiting for organic scaling;
  the reservation bills while set, so set it for the event window and
  remove it after.

### HTTP Behavior Attributes

- **Timeouts**: idle timeout (1-4000 s) and client keep-alive (60-604800 s).
- **Protocol posture**: HTTP/2 toggle (tri-state -- unset keeps the AWS
  default), desync mitigation mode (`monitor`/`defensive`/`strictest`), and
  dropping invalid request headers for header-smuggling hardening.
- **Header handling**: X-Forwarded-For processing mode, client-port
  propagation, Host-header preservation, and injection of the negotiated TLS
  version/cipher toward targets.
- **Availability calls**: WAF fail-open (availability vs. security when an
  attached WAF is unreachable) and ARC zonal shift.

### Observability

- **Three S3 log streams**: access logs (one entry per request), connection
  logs (TLS handshake details -- where negotiation failures that never become
  requests show up), and health-check logs (per-probe results), each with its
  own bucket and prefix.

### DNS

- **Route53 alias records**: point hostnames at the ALB with alias A records,
  which work at the zone apex, cost nothing per query, and inherit the ALB's
  health.

## Benefits

- **Stable shared infrastructure**: services come and go through listeners,
  rules, and target groups; the ALB and its ARN stay put.
- **Composability**: subnets, security groups, log buckets, and the Route53
  zone are all `valueFrom` references, so the architecture graph shows what
  the ALB depends on.
- **AWS defaults preserved**: only explicitly set attributes are sent to AWS;
  everything else keeps its AWS default instead of a module opinion.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `load_balancer_arn`: ARN of the ALB (what `AwsLbListener` resources attach through)
- `load_balancer_name`: final name assigned to the ALB (metadata.name, truncated to AWS's 32-character limit when necessary)
- `load_balancer_dns_name`: DNS name assigned by AWS
- `load_balancer_hosted_zone_id`: Route53 hosted zone ID for the ALB's DNS entry, for alias records

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
