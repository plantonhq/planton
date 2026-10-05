# Public API Service

This preset creates a public, scale-to-zero HTTP API: unauthenticated callers reach the service directly on its run.app URL, instances appear with traffic and disappear when idle.

## When to Use

- Public REST/GraphQL APIs, webhooks, and websites without their own load balancer
- Prototypes and internal tools that should cost nothing while idle
- The starting point most services grow from — add VPC egress, volumes, or traffic splitting as needs appear

## Key Configuration Choices

- **Scale-to-zero (`minInstanceCount: 0`)** — cheapest posture; set 1+ to eliminate cold starts at idle cost
- **`startupCpuBoost`** — temporarily boosts CPU during instance start, cutting cold-start latency for JIT-heavy runtimes
- **Startup probe on `/healthz`** — instances receive no traffic until the app really serves
- **`allowUnauthenticated: true`** — the additive-IAM public grant (`roles/run.invoker` to `allUsers`); delete it and callers need IAM identity tokens
- **`deletionProtection: false`** — kept off for easy teardown while iterating; omit the field (defaults to true) for anything real

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `us-docker.pkg.dev/my-project/my-repo/api:1.0.0` | Your container image | Artifact Registry |

## Related Presets

- **02-private-vpc-service** — the same service locked to internal traffic with VPC egress and Cloud SQL
- **03-gpu-inference** — GPU-backed model serving

## Related Kinds

- [GcpRegionNetworkEndpointGroup](/docs/catalog/gcp/gcpregionnetworkendpointgroup) — bridge this service into a global HTTPS load balancer for custom domains
