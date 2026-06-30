# Strict mTLS for One Workload, with a Plaintext Port

Require mTLS for a single selected workload, while exempting one port that must
stay plaintext -- for example a health-check, metrics-scrape, or legacy port that
a non-mesh client probes directly.

## When to Use

- A specific workload should enforce STRICT mTLS, but one of its ports is hit by
  a caller that cannot speak mTLS (a node-local health checker, a Prometheus
  scraper outside the mesh, etc.).
- You want a tighter, workload-scoped policy that overrides a looser
  namespace-wide default.

## Key Configuration Choices

- **`spec.selector.match_labels`** -- targets just the workload's pods, so this
  policy overrides any namespace-wide policy for them.
- **`mtls.mode: STRICT`** -- the workload default: every port requires mTLS...
- **`port_level_mtls["<port>"].mode: DISABLE`** -- ...except this one port, which
  stays plaintext. The key is the **workload** port (the container's port), not
  the Kubernetes Service port. `port_level_mtls` is only honored because a
  selector is present.

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running and the workload has a sidecar or is in the ambient mesh
  (`KubernetesIstio`).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | Namespace the workload runs in (e.g. `finance`). |
| `<workload>` | Value of the `app` label selecting the workload's pods (e.g. `finance`). |
| `<plaintext-port>` | Workload port to keep plaintext (e.g. `8080`). Quote it -- map keys are strings. |
