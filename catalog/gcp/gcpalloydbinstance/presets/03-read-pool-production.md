# Production Read Pool

This preset creates a regional three-node READ_POOL with connector enforcement, TLS-only connections, and query insights enabled.

## When to Use

- Production read scaling with security and observability requirements
- Teams that connect through AlloyDB Auth Proxy or Language Connectors

## Key Configuration Choices

- **requireConnectors + ENCRYPTED_ONLY** — rejects direct unauthenticated connections
- **queryInsightsConfig** — captures plans, application tags, and client addresses for slow-query diagnosis
- **nodeCount: 3** — headroom for read-heavy production traffic

## Related Kinds

- [GcpAlloydbUser](/docs/catalog/gcp/gcpalloydbuser) — pair read pools with per-application credentials
