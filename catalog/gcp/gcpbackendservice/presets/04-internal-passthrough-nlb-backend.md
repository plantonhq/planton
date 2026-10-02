# Internal Passthrough NLB Backend

A REGIONAL backend service for an internal passthrough Network Load Balancer: TCP traffic forwarded to instance groups with the client's source IP preserved, a primary pool and a failover pool, per-session connection tracking, and zonal affinity. A regional `GcpGlobalForwardingRule` on the `INTERNAL` scheme names this service directly through `backendService` — there is no proxy in a passthrough load balancer.

## When to Use

- Load balancing TCP or UDP between services inside a VPC (databases, brokers, internal APIs) where connections must not be terminated
- Active/standby pools: the failover pool takes traffic only when the primary's healthy ratio drops to `failoverRatio`
- Keeping traffic inside the client's zone for latency or egress cost, spilling across zones only when the local zone thins out

## Remix Notes

- `region` is what makes this a regional (passthrough-capable) backend service; `loadBalancingScheme: INTERNAL` is named outright because an unset scheme means `EXTERNAL` on both scopes.
- The health check must be a regional `GcpHealthCheck` with a TCP probe; a global health check is rejected by Google for the regional Application Load Balancer schemes and accepted for passthrough, so a regional one is the safe habit.
- `protocol: UNSPECIFIED` forwards every IP protocol — pair it with a forwarding rule using `ipProtocol: L3_DEFAULT` and `allPorts: true`.
- For a single active leader instead of pools, replace `failoverPolicy`, `connectionTrackingPolicy`, and `healthCheck` with `haPolicy` (`fastIpMove: GARP_RA` for keepalived-style VIP moves); Google forbids combining them.
- Reference the network, health check, and groups via `valueFrom` instead of literal self-links.
