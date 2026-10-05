# Istio Peer Authentication

Defines an Istio PeerAuthentication: a namespaced policy that controls whether incoming traffic to your workloads must arrive over a mutual-TLS (mTLS) tunnel. Use it to require encrypted, authenticated service-to-service traffic across a namespace, tighten a single workload, or carve out a specific port that must stay plaintext.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A PeerAuthentication policy** -- a namespaced Istio policy that tells the mesh how to treat incoming connections to the workloads it selects: require mTLS (Strict), accept either plaintext or mTLS (Permissive), refuse mTLS (Disable), or inherit the surrounding default (Unset). With no workload selector it sets the default for every workload in its namespace (mesh-wide when created in the mesh root namespace); with a selector it applies only to matching pods, overriding any looser namespace default; individual workload ports can override the workload-level mode.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Istio installed** -- the Istio CRDs (Istio Base CRDs) must be present and the Istio control plane (istiod) running. The policy is only enforced where istiod is active and the selected workloads are part of the mesh.
- **Target namespace exists** -- the policy is created in a specific namespace; reference an existing one or create it first.
- **Callers on the mesh before Strict** -- enabling Strict rejects plaintext, so confirm every client of the selected workloads connects over mTLS first. Permissive is the safe migration mode.

## Deploy

### Console

Open the deployment store, find **Istio Peer Authentication**, and click **Deploy**. The creation wizard walks you through the namespace, the optional workload selector, the mTLS mode, and any per-port overrides, with guidance at each step. Start from the **Require Strict mTLS Across a Namespace** preset in the [Presets](#presets) tab for the hardening baseline.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPeerAuthentication
metadata:
  name: default
  org: acme-corp
  env: prod
spec:
  namespace:
    value: prod-apps
  mtls:
    mode: STRICT
```

```shell
planton apply -f peer-authentication.yaml
```

This requires mutual TLS for every workload in the `prod-apps` namespace. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to place the policy in a namespace managed alongside it:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: prod-apps
      fieldPath: spec.name
  mtls:
    mode: STRICT
```

The InfraPipeline creates the namespace first, then the policy inside it. To order the policy after a workload it protects, express that dependency through `metadata.relationships` -- the selector itself is a runtime label match and creates no ordering edge.

## Key Configuration

These are the most important decisions when configuring a PeerAuthentication policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Strict is a cutover, not a starting point** -- the moment Strict applies, every plaintext caller of the selected workloads is refused. Migrate through **Permissive** (accepts both), confirm every client connects over mTLS, then flip to Strict. Rolling back is the same one-field change.

**Selector scope decides the blast radius** -- no selector means the whole namespace (and the whole mesh when created in the Istio root namespace); a selector narrows the policy to matching pods and overrides any looser namespace default for them. The selector is a plain label match, not a resource reference: istiod evaluates it at runtime, so it creates no ordering dependency -- use `metadata.relationships` when deploy order matters in an InfraChart.

**Per-port overrides require a selector** -- `portLevelMtls` keys on the workload container's port numbers (never the Kubernetes Service port) and is honored by istiod only when a workload selector is present; the spec enforces the pairing at apply time. The classic use: keep a metrics or health-check port plaintext while everything else is Strict.

**Omitting `mtls` and setting UNSET are both inheritance -- but differently visible** -- leaving the `mtls` block out inherits the parent (namespace, then mesh) policy silently; `mode: UNSET` declares the inheritance explicitly in the manifest. An `mtls` block, once present, must carry a real mode.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

This component's `status.outputs` only echo the resource's identity back -- `peer_authentication_name` (equals `metadata.name`) and `namespace` (the resolved `spec.namespace`). PeerAuthentication has no controller-reconciled status worth consuming: istiod enforces the policy in the data plane, so there is nothing downstream to wire via ValueFromRef. Order dependent resources through `metadata.relationships` instead.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Namespace-wide Strict mTLS** -- require mutual TLS for every workload in a namespace. The canonical hardening baseline. Start from the **Require Strict mTLS Across a Namespace** preset.

**Strict workload with a plaintext port** -- require mTLS for one selected workload while exempting a single port (e.g. a metrics scrape or health check that a non-mesh client probes). Start from the **Strict mTLS for One Workload, with a Plaintext Port** preset.

## Works With

- [**Istio Base CRDs**](/infra-catalog/kubernetes-istio-base-crds) -- installs the CRDs this policy's API depends on
- [**Istio**](/infra-catalog/kubernetes-istio) -- the control plane (istiod) that enforces the policy in the mesh
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace the policy governs
- [**Istio Authorization Policy**](/infra-catalog/kubernetes-authorization-policy) -- the natural pairing: PeerAuthentication decides HOW traffic arrives (mTLS), AuthorizationPolicy decides WHO may call
- [**Istio Request Authentication**](/infra-catalog/kubernetes-request-authentication) -- end-user JWT validation layered on top of the mTLS transport
