# Basic TCP Accelerator

This preset creates a minimal Global Accelerator that accepts TCP traffic on port 443 and routes it to a single ALB endpoint. It uses all default settings: AWS-allocated static anycast IPs, no flow logs, TCP health checks with 30-second intervals, and no client affinity. This is the simplest useful Global Accelerator configuration and serves as a starting point for most deployments.

## When to Use

- You need static anycast IPs as the entry point for a single-region application
- You want instant BGP-based failover instead of DNS TTL-dependent failover
- You're adding Global Accelerator in front of an existing ALB to improve global client latency
- You're evaluating Global Accelerator and want a low-effort starting configuration

## Key Configuration Choices

- **TCP protocol on port 443** — the standard HTTPS entry point, suitable for web applications and APIs
- **Single endpoint group** — uses the default provider region (no explicit `endpointGroupRegion`), appropriate for single-region deployments
- **TCP health checks** — verifies port reachability only; upgrade to HTTP health checks (`healthCheckProtocol: HTTP`, `healthCheckPath: /health`) for application-level health validation
- **Weight 128** — the default midpoint on the 0–255 scale, suitable for a single endpoint

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `arn:aws:elasticloadbalancing:...` | ARN of the ALB, NLB, EIP, or EC2 instance to route traffic to | AWS Console, CLI, or the `status.outputs` of an `AwsAlb` or `AwsNetworkLoadBalancer` resource |

## Related Presets

- **02-multi-region-production** — Use for production workloads requiring multi-region failover, HTTP health checks, and flow logs
- **03-gaming-udp-accelerator** — Use for UDP-based workloads like gaming or real-time media
