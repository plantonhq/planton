# Istio-Enabled Namespace

This preset creates a namespace with Istio service mesh sidecar injection enabled. All pods deployed in this namespace will automatically receive an Istio sidecar proxy for mTLS, traffic management, and observability.

## When to Use

- You run Istio and want automatic sidecar injection for workloads in this namespace
- You need mTLS between services without application-level TLS configuration
- You want Istio traffic management features (traffic splitting, retries, circuit breaking)

## Key Configuration Choices

- **Service mesh enabled** (`true`) with Istio -- adds the `istio-injection: enabled` label (or equivalent) to the namespace
- **Resource profile** (`medium`) -- Istio sidecars consume additional CPU/memory; medium provides more headroom than small
- **Baseline pod security** -- compatible with Istio sidecar requirements (Istio init containers need `NET_ADMIN` capability)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-team-name>` | Team or project that owns this namespace | Your organization's team registry |

## Related Presets

- **01-standard** -- Namespace without service mesh
- **02-production-with-quotas** -- Production-hardened namespace with custom quotas
