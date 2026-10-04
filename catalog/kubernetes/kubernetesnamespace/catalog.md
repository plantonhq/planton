# Kubernetes Namespace

Deploys a production-ready Kubernetes namespace with optional resource quotas, LimitRanges, network policies, pod security standards, and service mesh sidecar injection. Manages namespace-level governance as a single Infra Component, turning multi-object namespace setup into a one-step operation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- the core namespace object with user-specified labels, annotations, and pod security standard labels
- **ResourceQuota** -- created only when a resource profile is configured (built-in T-shirt size or custom quotas); enforces aggregate CPU, memory, and object count limits
- **LimitRange** -- created only when custom default limits are specified; injects default CPU and memory requests/limits into containers that omit them
- **NetworkPolicy (ingress isolation)** -- created only when `networkConfig.isolateIngress` is `true`; default-denies all ingress and allows traffic only from namespaces listed in `allowedIngressNamespaces`
- **NetworkPolicy (egress restriction)** -- created only when `networkConfig.restrictEgress` is `true`; blocks all egress except to kube-system (DNS) and the Kubernetes API, plus any CIDRs in `allowedEgressCidrs`
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- A running Kubernetes cluster (v1.25+ recommended for Pod Security Standards enforcement).
- A CNI that supports NetworkPolicy (Calico, Cilium, or similar) if you plan to enable ingress isolation or egress restriction.
- Istio, Linkerd, or Consul installed if you plan to enable service mesh sidecar injection.

## Deploy

### Console

Open the deployment store, find **Kubernetes Namespace**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Namespace** preset for development or the **Production Namespace with Custom Quotas** preset for production workloads in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesNamespace
metadata:
  name: backend-services
  org: acme-corp
  env: prod
spec:
  name: backend-services
  resourceProfile:
    preset: medium
  podSecurityStandard: baseline
```

```shell
planton apply -f namespace.yaml
```

This creates a namespace with a medium resource quota profile (4 CPU / 8Gi requests, 8 CPU / 16Gi limits, 50 pods max) and baseline pod security enforcement. Network isolation and service mesh injection are not enabled. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a Kubernetes Namespace. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Resource profile** -- Choose a built-in T-shirt size (`small`, `medium`, `large`, `xlarge`) for quick resource guardrails, or specify `custom` quotas for precise control over CPU, memory, object counts, and default container limits. Custom quotas also accept `additionalHardLimits` -- an open map for any other ResourceQuota key Kubernetes supports (extended resources like `requests.nvidia.com/gpu`, storage classes, hugepages), merged with the typed fields into one ResourceQuota. Omitting the profile creates a namespace with no resource quotas.

**Pod security standard** -- Set to `baseline` for most workloads (prevents known privilege escalations), `restricted` for security-sensitive production (requires non-root, drops all capabilities), or `privileged` only for system-level workloads like monitoring agents and CNI plugins.

**Network isolation** -- Enable `networkConfig.isolateIngress` to default-deny all inbound traffic, then whitelist specific namespaces via `allowedIngressNamespaces`. Enable `networkConfig.restrictEgress` to block outbound traffic except DNS and the Kubernetes API, then open specific CIDRs or domains.

**Service mesh injection** -- Set `serviceMeshConfig.enabled` to `true` and choose a `meshType` (Istio, Linkerd, or Consul) to add the appropriate sidecar injection label. For Istio, the optional `revisionTag` field targets a specific control plane revision for safe canary upgrades.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | The name of the created namespace | The `spec.namespace` reference nearly every namespaced kind wires against |

The remaining outputs (`namespace_id`, the `*_applied` flags, `pod_security_standard`, `service_mesh_type`, `labels_json`, `annotations_json`) echo the applied configuration back for verification and audit — they carry nothing downstream resources compose on.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Development namespace with basic guardrails** -- Small resource profile, baseline pod security, no network isolation. Fast to provision and suitable for most development and staging workloads. Start from the **Standard Namespace** preset.

**Production namespace with governance** -- Custom resource quotas, default container limits, ingress isolation allowing only kube-system and istio-system, restricted pod security. Designed for multi-tenant production clusters. Start from the **Production Namespace with Custom Quotas** preset.

**Service mesh namespace** -- Medium resource profile with Istio sidecar injection enabled. All pods deployed in this namespace automatically receive an Istio proxy for mTLS, traffic management, and observability. Start from the **Istio-Enabled Namespace** preset.

## Works With

This component references no other components -- it IS the placement unit the rest of the Kubernetes catalog builds on. Nearly every namespaced kind (Deployment, Secret, ConfigMap, ServiceAccount, Ingress, and the rest) carries a `spec.namespace` field that references this component's `spec.name`, so Infra Charts create the namespace first and place everything else inside it in dependency order.
