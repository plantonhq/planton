# Overview

The AwsNlb API resource provisions an AWS Network Load Balancer: the Layer-4
entry point for TCP/UDP/TLS traffic, with static IP addresses per
Availability Zone and millions of connections per second of headroom.

## Why We Created This API Resource

The load balancer carries no routing configuration -- that is deliberate.
Planton models the ELBv2 surface as composable kinds, split exactly where AWS
splits it:

- **`AwsNlb`** (this component) owns what is truly load-balancer-wide: node
  placement with optional static IPs (subnet mappings), optional security
  groups, and traffic distribution behavior (cross-zone, DNS client routing,
  zonal shift).
- **`AwsLbListener`** attaches to the NLB by ARN and owns a port, protocol
  (TCP/UDP/TCP_UDP/TLS), and TLS material. NLB listeners only forward -- AWS
  rejects every other action type at Layer 4 -- and listener rules do not
  apply; routing is purely by port/protocol.
- **`AwsLbTargetGroup`** receives the connections.

Adding a port or rotating a certificate edits one listener; the NLB itself --
and the Elastic IPs partners have allowlisted -- stays put.

## Key Features

### Static IP Placement

- **Subnet mappings**: each mapping pins one NLB node to a subnet, optionally
  with an Elastic IP (`allocationId`, typically an `AwsElasticIp` reference)
  for internet-facing NLBs, a fixed `privateIpv4Address` for internal ones,
  or a fixed `ipv6Address` for dualstack nodes.
- **Static public IPs are the headline NLB feature**: they survive scaling
  events and maintenance, so partners, firewalls, and legacy systems can
  allowlist them permanently -- on IPv4 and IPv6 alike.

### Traffic Distribution and Capacity

- **Cross-zone toggle**: off by AWS default (unlike ALB) because inter-AZ
  data transfer is billed; enable it when target distribution across AZs is
  uneven.
- **DNS client routing policy**: `any_availability_zone`,
  `availability_zone_affinity`, or `partial_availability_zone_affinity` --
  trades latency and cross-zone cost against spillover capacity.
- **Zonal shift**: allows Amazon Application Recovery Controller to drain an
  impaired Availability Zone.
- **Capacity reservation**: `minimumLoadBalancerCapacityUnits` pre-provisions
  LCUs for a known traffic level (launches, failover targets) instead of
  waiting for organic scaling; reserved capacity bills whether used or not.
- **Source-port headroom**: `secondaryIpsAutoAssignedPerSubnet` (0-7) widens
  the per-node source-port budget for very high connection counts to a
  single target; decreasing it later replaces the load balancer.
- **IPv6 prefix source NAT**: `enablePrefixForIpv6SourceNat` switches
  dualstack nodes to a /80 prefix per AZ for source NAT -- required for UDP
  listeners on a dualstack NLB.

### Security

- **Optional security groups**: unlike ALB, an NLB can run without them --
  but attaching any is a one-way door: once attached, the last group can
  never be removed, only replaced.
- **PrivateLink enforcement**: choose whether inbound security-group rules
  apply to traffic arriving through PrivateLink VPC endpoints (`on`/`off`).

### Observability and DNS

- **Access logs to S3**: TLS-listener traffic only (an AWS limitation --
  plain TCP/UDP flows are not logged).
- **Route53 alias records**: point hostnames at the NLB with alias A records,
  which work at the zone apex and cost nothing per query.

## Benefits

- **Stable allowlisted endpoints**: the NLB and its Elastic IPs outlive every
  listener and target-group change behind them.
- **Composability**: subnets, Elastic IPs, security groups, the log bucket,
  and the Route53 zone are all `valueFrom` references, so the architecture
  graph shows what the NLB depends on.
- **AWS defaults preserved**: only explicitly set attributes are sent to AWS;
  everything else keeps its AWS default instead of a module opinion.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `load_balancer_arn`: ARN of the NLB (what `AwsLbListener` resources attach through; also what Global Accelerator endpoints reference)
- `load_balancer_name`: final name assigned to the NLB (metadata.name, truncated to AWS's 32-character limit when necessary)
- `load_balancer_dns_name`: DNS name assigned by AWS
- `load_balancer_hosted_zone_id`: Route53 hosted zone ID for the NLB's DNS name, for alias records

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
