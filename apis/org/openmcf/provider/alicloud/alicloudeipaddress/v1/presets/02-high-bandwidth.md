# High-Bandwidth Production EIP

This preset creates a production-grade EIP with 100 Mbps guaranteed bandwidth, PayByBandwidth metering, and BGP_PRO premium routing. Use this for internet-facing load balancers, high-traffic NAT gateways, or any workload where consistent throughput matters more than cost optimization.

## When to Use

- Production load balancers (ALB/NLB) serving sustained internet traffic
- NAT gateways for clusters with many nodes pulling images or sending telemetry
- Any workload where you need guaranteed bandwidth rather than a traffic-based ceiling
- China mainland deployments requiring optimized BGP routing (BGP_PRO)

## Key Configuration Choices

- **100 Mbps bandwidth** -- Guaranteed allocation. Increase to 200-1000 Mbps for very high-traffic scenarios. The cost scales linearly with bandwidth for PayByBandwidth.
- **PayByBandwidth** -- You pay for the reserved 100 Mbps regardless of actual usage. This is cost-effective when sustained utilization exceeds ~30-40% of the bandwidth allocation. For bursty workloads, consider the standard preset with PayByTraffic instead.
- **BGP_PRO** -- Premium BGP with optimized routing in mainland China. Provides better latency and reliability for domestic traffic. Available only in China mainland regions (cn-*). For non-China regions, use `BGP` instead.
- **Resource group** -- Production EIPs should be assigned to a resource group for access control and billing isolation.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (must be `cn-*` for BGP_PRO) | Your deployment region strategy |
| `<your-eip-name>` | Descriptive name (e.g., `prod-alb-eip`, `prod-nat-eip`) | Your naming convention |
| `<your-resource-group-id>` | Resource group ID (e.g., `rg-prod-123`) | Alibaba Cloud console > Resource Management |
| `<your-team>` | Team or business unit | Your organizational structure |
| `<your-cost-center>` | Cost center code | Your finance or cloud operations team |

## Related Presets

- **01-standard** -- Use instead for development/staging or bursty workloads where PayByTraffic is more cost-effective
