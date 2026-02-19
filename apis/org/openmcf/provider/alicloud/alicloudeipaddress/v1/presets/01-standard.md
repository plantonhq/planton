# Standard EIP

This preset creates an Elastic IP Address with all defaults: 5 Mbps bandwidth, PayByTraffic metering, and BGP multi-line ISP. This is the most common configuration for development, staging, and lightweight production workloads where the EIP serves as the public endpoint for a NAT gateway, load balancer, or VPN gateway.

## When to Use

- Development and staging environments where bandwidth demand is low
- NAT gateway EIPs where outbound traffic is bursty and unpredictable
- Any scenario where PayByTraffic metering is cost-effective (most use cases under 100 Mbps sustained)

## Key Configuration Choices

- **5 Mbps bandwidth** (default) -- Acts as a ceiling for PayByTraffic; you only pay for actual data transferred. Increase to 10-50 Mbps if your workload has frequent bursts that could be throttled.
- **PayByTraffic** (default) -- Pay per GB of outbound data. At typical Alibaba Cloud rates, PayByTraffic is cheaper than PayByBandwidth until you sustain over ~30-40% utilization of reserved bandwidth.
- **BGP** (default) -- Multi-line BGP is available in all regions and provides the best routing for general-purpose workloads.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-eip-name>` | Descriptive name for the EIP (e.g., `prod-nat-eip`, `staging-alb-eip`) | Your naming convention |
| `<your-team>` | Team or business unit that owns this EIP | Your organizational structure |
| `<nat\|alb\|vpn>` | The intended association target | Choose one: `nat`, `alb`, `nlb`, `vpn`, `ecs` |

## Related Presets

- **02-high-bandwidth** -- Use instead for production workloads requiring guaranteed high-throughput bandwidth
