# Kubernetes Metrics Server

## When NOT to Use This

**One installation per cluster.** metrics-server registers the cluster-wide
`v1beta1.metrics.k8s.io` APIService — a singleton. Check whether your cluster
already serves it before adding this component (a second installation would
fight the first over the same APIService). **GKE and AKS ship
metrics-server built-in as a managed component — do not install this
component there.** EKS, kind, k3s, kubeadm, and self-managed clusters
need it.

Also not the right component when:

- **You need custom or external metrics** (queue depth, requests per second)
  — that is Prometheus-adapter / KEDA territory; metrics-server serves only
  CPU and memory.
- **You need a monitoring system** — metrics-server keeps NO history and
  serves only instantaneous values; kube-prometheus-stack is the
  observability story.

## Overview

**KubernetesMetricsServer** installs metrics-server — the cluster's
resource-metrics pipeline — from the official Helm chart (`metrics-server` at
`https://kubernetes-sigs.github.io/metrics-server/`). It scrapes each kubelet
and serves live CPU/memory usage through the `metrics.k8s.io` APIService,
which is what `kubectl top` reads and what HorizontalPodAutoscalers need for
utilization targets. Without it, HPAs deploy but never receive metric values
and never scale.

The typed spec covers the chart's meaningful configuration surface, with a
`helm_values` escape hatch (merged last, Helm `-f` semantics, identical on
both engines) for anything beyond it.

**Key design points:**

- **The release name is fixed to `metrics-server`** (and the chart fullname is
  pinned to it) — one installation per cluster is an upstream constraint, so
  the release name never derives from `metadata.name`, and every chart object
  gets a deterministic name.
- **`kubelet_insecure_tls` is THE critical knob**: required on clusters whose
  kubelets serve self-signed certificates (kind, k3s, kubeadm without kubelet
  certificate rotation, many on-prem setups); leave it false on EKS/AKS.
- **The install waits for real readiness**: both engines install atomically
  and wait for the Deployment, and the chart's `/readyz` probe only passes
  once the first kubelet scrape succeeds — a wrong TLS posture fails the
  deploy loudly instead of surfacing later as HPAs that never scale.
- **The module owns the chart's `defaultArgs` list**: it re-renders it with
  the typed substitutions (`kubelet_preferred_address_types`,
  `metric_resolution`) applied, so the pod spec stays canonical instead of
  carrying confusing duplicate flags.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: installation namespace (`kube-system` is the upstream
  convention — the APIService is cluster infrastructure; a dedicated
  `metrics-server` namespace also works) — literal or a KubernetesNamespace
  reference

### Common

- **`spec.create_namespace`**: create (and own) the namespace with the
  release — usually false for `kube-system`, which always exists
- **`spec.chart_version`**: pinned chart version (default `3.13.1`, which
  ships metrics-server 0.8.1)
- **`spec.kubelet_insecure_tls`**: skip kubelet serving-certificate
  verification — required on self-signed-kubelet clusters, false on managed
  clouds
- **`spec.replicas` + `spec.pod_disruption_budget`**: run 2 replicas guarded
  by a PodDisruptionBudget (minAvailable 1) for HA of the metrics API
- **`spec.metric_resolution`**: kubelet scrape interval (default `15s` — the
  value HPA freshness is built around)
- **`spec.api_service`**: APIService registration — `create` (default true),
  `insecure_skip_tls_verify` (default true, matching the default self-signed
  serving certificate), `ca_bundle`
- **`spec.tls`**: serving-certificate provisioning — `self_signed` (default),
  `helm`, `cert_manager` (with an existing Issuer/ClusterIssuer reference),
  or `existing_secret`
- **`spec.host_network`**: run on the host network where the API server
  cannot reach pod IPs over the overlay (the upstream example: Weave CNI on
  EKS)
- **`spec.prometheus`**: expose metrics-server's OWN telemetry and optionally
  a ServiceMonitor (requires the Prometheus operator CRDs — the release fails
  without them)
- **`spec.helm_values`**: escape hatch for chart values beyond the typed
  fields — never the primary interface

## Stack Outputs

| Output | Purpose |
|---|---|
| `namespace` | Installation namespace |
| `release_name` | Helm release name (always `metrics-server`) |
| `service_name` | The Service the APIService routes to (always `metrics-server`) |
| `api_service_name` | `v1beta1.metrics.k8s.io` — empty when `api_service.create` is false |

## Composing in Infra Charts

The standard wiring: this component first, then every
KubernetesHorizontalPodAutoscaler on the cluster just works — HPAs consume
the metrics API anonymously, no reference needed. For a verified serving
chain, deploy KubernetesCertManager and a KubernetesClusterIssuer in the same
chart and point `tls.cert_manager_issuer.cluster_issuer.name` at the
issuer's `status.outputs.cluster_issuer_name` (a namespaced KubernetesIssuer
goes in `tls.cert_manager_issuer.issuer.name` instead).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
