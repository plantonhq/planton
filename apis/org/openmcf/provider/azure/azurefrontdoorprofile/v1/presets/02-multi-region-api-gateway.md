# Multi-Region API Gateway

A Standard-tier Front Door profile configured as an API gateway with multi-origin health-based failover and path-based routing to separate API and static asset backends.

## What This Preset Provides

- **Standard SKU**: Global load balancing with SSL offloading at 99.99% SLA
- **Two origin groups**: Separate backend pools for API services and static assets
- **Multi-region API origins**: Two API origins (east + west) with equal weight for active-active distribution
- **Fast health probes**: 15-second interval HTTPS probes on `/api/healthz` for rapid failure detection
- **Path-based routing**: `/api/*` traffic routed to API backends, `/static/*` to blob storage
- **Static asset caching**: Compression and edge caching enabled for the static route
- **No session affinity**: Disabled for stateless API backends
- **Lower timeout**: 60-second response timeout tuned for API workloads

## When to Use

- API + SPA architecture where frontend and backend have different backends
- Multi-region API deployment needing health-based failover
- Latency-sensitive APIs that benefit from anycast edge routing
- Mixed dynamic API traffic and cacheable static assets

## What to Customize

| Field | What to Set |
|-------|-------------|
| `resourceGroup.value` | Your Azure resource group name |
| `name` | Globally unique profile name |
| `origins[].hostName` (api-east) | Your East US API hostname |
| `origins[].hostName` (api-west) | Your West US API hostname |
| `origins[].hostName` (blob-origin) | Your storage account blob hostname |
| `healthProbe.path` | Your API health check endpoint |
| `responseTimeoutSeconds` | Adjust based on your API latency characteristics |
| `patternsToMatch` | Adjust URL path patterns to match your routing needs |
