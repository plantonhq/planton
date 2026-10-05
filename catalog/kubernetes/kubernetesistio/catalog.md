# Istio

Deploys the Istio service mesh control plane on a Kubernetes cluster: istiod, the Istio CRDs, and -- in ambient mode -- the istio-cni node agent and ztunnel node proxies. Control-plane-only by design: no ingress gateway ships with it; gateways are provisioned on demand from Gateway API Gateways that use the `istio` GatewayClass. One version pin drives every chart, and per-release Helm value escape hatches cover the long tail.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Istio CRDs** -- applied by the module itself via server-side apply, outside any Helm release. Module-owned CRDs are co-ownable with the CRDs-only Istio Base CRDs component, so a cluster running just the CRDs upgrades to the full mesh with a plain redeploy.
- **base Helm Release** -- the validation-webhook plumbing and cluster-wide resources istiod requires.
- **istiod Helm Release** -- the control plane, with your capacity, autoscaling, mesh configuration, proxy defaults, and scheduling choices rendered as chart values.
- **cni Helm Release** -- the istio-cni node agent. Always installed in ambient mode (it is how traffic reaches ztunnel); opt-in in sidecar mode, where it replaces the injected privileged init-container.
- **ztunnel Helm Release** -- the ambient L4 node proxy. Installed in ambient mode only.
- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; hosts everything above.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- One KubernetesIstio per cluster: the CRDs and the validation-webhook plumbing are cluster singletons.
- In sidecar mode, namespaces that should participate in the mesh need the `istio-injection: enabled` label (or workloads opt in individually when the injection policy is `disabled`).
- Auto-provisioned gateways default to LoadBalancer Services -- on clusters without a cloud load-balancer controller (bare metal, kind, k3s), set the gateway service type to NodePort or ClusterIP.

## Deploy

### Console

Open the deployment store, find **Istio**, and click **Deploy**. The creation wizard walks you through placement, the release pin, the sidecar-vs-ambient decision, control-plane capacity, mesh identity, proxy defaults, the node agents, gateway defaults, and the Helm escape hatches -- with guidance at each step. Start from the **Standard Istio Mesh (Sidecar)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesIstio
metadata:
  name: mesh
  org: acme-corp
  env: prod
spec:
  namespace:
    value: istio-system
  createNamespace: true
  version: 1.30.3
  dataplaneMode: sidecar
  meshConfig:
    trustDomain: prod.acme.internal
  istiod:
    priorityClassName: system-cluster-critical
```

```shell
planton apply -f istio.yaml
```

This installs the control plane into `istio-system` with a production trust domain and eviction protection for istiod; everything else runs on the chart defaults. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the control plane to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: istio-system-namespace
      fieldPath: spec.name
  createNamespace: false
  dataplaneMode: sidecar
```

The InfraPipeline deploys the namespace first, then installs the mesh into it.

## Key Configuration

These are the most important decisions when configuring an Istio mesh. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One version drives every chart** -- Istio versions its charts in lockstep, and `version` pins them all. Upgrades step ONE minor at a time -- skipping minors in place is unsupported upstream, so a cluster two minors behind is two sequential redeploys away, not one.

**The data plane mode is an install-time decision** -- `sidecar` (a proxy per pod, the classic and widely documented mode) or `ambient` (sidecar-less: per-node ztunnel for L4/mTLS, optional waypoints for L7). Switching later is a workload-by-workload migration, not a redeploy -- choose deliberately.

**Set the trust domain BEFORE production** -- it is the identity root of every workload certificate (`spiffe://<trustDomain>/...`), and changing it re-identifies every workload in the mesh: authorization policies matching the old principals stop matching. A stable, organization-unique value from day one costs nothing; changing it live costs a policy migration.

**Control-plane capacity is fixed replicas OR the chart HPA** -- the two are mutually exclusive. Pair whichever you choose with a PodDisruptionBudget and `istiod.priorityClassName: system-cluster-critical` -- an evicted control plane stops programming proxies.

**Proxy resources multiply** -- the per-proxy requests/limits in the proxy defaults apply to EVERY injected pod, so a 100m CPU request across 500 pods is 50 cores of reserved capacity. Size them from real sidecar usage, not habit.

**`REGISTRY_ONLY` is the egress lockdown** -- the outbound traffic policy that blocks all external traffic except services declared as Service Entries. The secure posture, and a breaking one: turn it on before workloads depend on undeclared egress, or inventory egress first.

**Gateways default to LoadBalancer** -- auto-provisioned Gateway API gateways create LoadBalancer Services; on clusters without a cloud LB controller (bare metal, kind, k3s) set the gateway Service type to NodePort or ClusterIP or the gateway never gets an address.

**Helm values merge last** -- per-release YAML documents over the typed fields, for chart surface beyond them -- the escape hatch, never the primary interface.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components and operators can reference:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Where the control plane runs | Placing mesh-adjacent tooling beside istiod |
| `istiod_service_name` | The discovery Service proxies connect to | Debugging and custom bootstrap configuration |
| `revision` | The control-plane revision (empty = default) | Labeling workloads `istio.io/rev` under revisioned operation |
| `gateway_class_name` | The GatewayClass this control plane serves | Gateway API Gateways reference it to be provisioned by this mesh |
| `trust_domain` | The identity root (`spiffe://<trust_domain>/...`) | Writing authorization policies that match workload identities |
| `dataplane_mode` | `sidecar` or `ambient` | Runbooks and tooling that branch on the mesh's mode |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Classic sidecar mesh** -- the widely documented default: label namespaces for injection and get a proxy per pod. Start from the **Standard Istio Mesh (Sidecar)** preset.

**Ambient mesh** -- sidecar-less L4/mTLS everywhere with waypoints added only where L7 policy is needed -- lower overhead, no workload restarts on proxy upgrades. Start from the **Ambient Mesh (Sidecar-less)** preset.

**Production sidecar mesh** -- the classic mode hardened: control-plane HA, eviction protection, and a production trust domain. Start from the **Production Sidecar Mesh** preset.

**Egress lockdown** -- `REGISTRY_ONLY` outbound policy with every external service declared as a Service Entry.

## Works With

- [**Istio Destination Rule**](/infra-catalog/kubernetes-destination-rule) and [**Istio Service Entry**](/infra-catalog/kubernetes-service-entry) -- traffic management against this control plane.
- [**Istio Peer Authentication**](/infra-catalog/kubernetes-peer-authentication), [**Istio Request Authentication**](/infra-catalog/kubernetes-request-authentication), and [**Istio Authorization Policy**](/infra-catalog/kubernetes-authorization-policy) -- mesh security policy.
- [**Istio Telemetry**](/infra-catalog/kubernetes-telemetry) -- observability configuration; [**Istio Envoy Filter**](/infra-catalog/kubernetes-envoy-filter) -- the extensibility escape hatch.
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- Gateways using the exported `gateway_class_name` are provisioned by this control plane, with [**Kubernetes HTTPRoute**](/infra-catalog/kubernetes-http-route) and its siblings attaching to them.
- [**Istio Base CRDs**](/infra-catalog/kubernetes-istio-base-crds) -- the CRDs-only alternative for clusters that need the API types without the mesh runtime; upgrading to this full mesh later is a plain redeploy.
