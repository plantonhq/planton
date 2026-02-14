# Standard Istio Service Mesh

This preset deploys the Istio control plane (istiod) with recommended default resources. Istio provides mTLS, traffic management, observability, and security for service-to-service communication.

## When to Use

- You need mTLS between services without application-level TLS
- You want traffic management features (canary deployments, traffic splitting, retries, circuit breaking)
- You need service mesh observability (distributed tracing, traffic metrics)

## Key Configuration Choices

- **Namespace** (`istio-system`) -- the standard namespace for Istio components
- **Default resources** -- sufficient for the istiod control plane; data plane (sidecar) resources are configured per namespace
- **Sidecar injection** -- enable per namespace by adding the `istio-injection: enabled` label (see KubernetesNamespace preset 03-istio-enabled)

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.
