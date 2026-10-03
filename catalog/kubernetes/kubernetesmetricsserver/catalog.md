# Metrics Server

Installs metrics-server — the cluster's resource-metrics pipeline — from the official Helm chart. It scrapes kubelets and serves live CPU/memory usage through the `metrics.k8s.io` API: what `kubectl top` reads and what HorizontalPodAutoscalers need to scale on resource metrics. It serves only instantaneous CPU/memory — no history, no custom metrics. One installation per cluster: the `v1beta1.metrics.k8s.io` APIService is a cluster-wide singleton.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Helm Release** -- the metrics-server Deployment, Service, and RBAC (the release name is fixed to `metrics-server`), pinned to the chart version you choose
- **APIService registration** -- the cluster-wide `v1beta1.metrics.k8s.io` APIService that routes resource-metrics queries to this installation
- **Namespace** (optional) -- created with standard governance labels when `createNamespace` is true (`kube-system` installs leave it false)

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Kubernetes Cluster

- A cluster WITHOUT a metrics API already in place — EKS, AKS, kind, k3s, kubeadm, or self-managed. **GKE ships metrics-server built-in — do not install this component there.**
- On clusters whose kubelets serve self-signed certificates (kind, k3s, kubeadm without kubelet certificate rotation, many on-prem setups), plan to accept kubelet TLS without verification or the install fails its readiness wait — the release only reports ready after the first successful kubelet scrape.

## Deploy

### Console

Open the deployment store, find **Metrics Server**, and click **Deploy**. The creation wizard walks you through placement, the chart pin, kubelet TLS posture, availability, observability, and scheduling. Start from the **Managed Cloud (EKS / AKS)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesMetricsServer
metadata:
  name: metrics-server
  org: acme-corp
  env: prod
spec:
  namespace:
    value: kube-system
```

```shell
planton apply -f metrics-server.yaml
```

This installs metrics-server into `kube-system` with the chart defaults -- verified kubelet TLS, one replica, the APIService registered. On a self-signed-kubelet cluster add `kubeletInsecureTls: true`, or the release never turns ready. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire metrics-server to a namespace or TLS material managed by other Infra Components:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: metrics-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then installs metrics-server into it.

## Key Configuration

These are the most important decisions when configuring metrics-server. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Readiness is the first scrape** -- the Helm release waits for the metrics-server pod to turn ready, and readiness requires a successful kubelet scrape. Every TLS misconfiguration therefore surfaces as a stuck install, not a running-but-broken one.

**Kubelet TLS posture** -- managed clouds (EKS, AKS) verify kubelet certificates cleanly with the defaults. Self-signed-kubelet clusters (kind, k3s, kubeadm) need the insecure-TLS dial — a deliberate, taught trade-off — or a properly distributed kubelet CA for the verified posture.

**Availability vs the drain wedge** -- a PodDisruptionBudget with a single replica wedges node drains (the one pod can never be evicted voluntarily). Raise replicas when enabling the PDB.

**The escape hatch** -- `helmValues` carries additional chart values as a YAML document, merged LAST over everything the typed fields render — for the chart surface beyond the typed fields, never the substitute for them, and never a place for secrets.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesSecret** | `tls.existingSecretName` | `metadata.name` |
| **KubernetesIssuer** | `tls.certManagerIssuer.issuer.name` | `status.outputs.issuer_name` |
| **KubernetesClusterIssuer** | `tls.certManagerIssuer.clusterIssuer.name` | `status.outputs.cluster_issuer_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Installation namespace | Debugging and composition |
| `release_name` | Helm release name (always `metrics-server`) | Debugging the release (`helm status`) |
| `service_name` | The Service the APIService routes to (always `metrics-server`) | Diagnosing the metrics pipeline |
| `api_service_name` | `v1beta1.metrics.k8s.io`; empty when the APIService registration is disabled | Verifying the metrics API registration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Managed cloud** -- EKS/AKS clusters with verifiable kubelet certificates; chart defaults do the rest. Start from the **Managed Cloud (EKS / AKS)** preset.

**Self-signed kubelets** -- kind, k3s, and kubeadm clusters accept kubelet TLS without verification — the posture every local-cluster runbook assumes. Start from the **Self-Signed Kubelets (kind / k3s / kubeadm / on-prem)** preset.

**HA with verified TLS** -- two replicas, a PodDisruptionBudget, and verified kubelet certificates for production fleets. Start from the **HA with Verified TLS (production hardening)** preset.

## Works With

- [**Kubernetes HorizontalPodAutoscaler**](/infra-catalog/kubernetes-horizontal-pod-autoscaler) -- consumes the resource-metrics API this component registers, with no explicit reference needed
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- `kubectl top pod` and `kubectl top node` work for its workloads the moment the first scrape lands
- [**KEDA**](/infra-catalog/kubernetes-keda) -- complements metrics-server: KEDA covers event-driven and external metrics, metrics-server covers instantaneous CPU/memory
- [**Cert Manager Issuer**](/infra-catalog/kubernetes-issuer) -- signs the serving certificate when the cert-manager TLS arm is chosen
