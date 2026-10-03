# Istio Destination Rule

Defines an Istio DestinationRule: a namespaced resource that tunes *how* traffic is sent to a service once routing has chosen it. It controls load balancing, connection-pool sizing, circuit breaking (outlier detection), and the TLS the sidecar originates upstream, and it can carve a service into named subsets (versions) that route rules target. This is how you add resilience, session affinity, and secure egress to a service without changing the application.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **DestinationRule** -- a namespaced `networking.istio.io/v1` policy applying the declared traffic policy to a service host, plus any named subsets and per-port overrides. istiod picks it up and programs every sidecar (or only the sidecars matched by `workloadSelector`) accordingly.

The rule is pure configuration: no pods, no Services. A rule written for a host that is not in the service registry (no Kubernetes Service and no ServiceEntry) is silently ignored -- deploying it "succeeds" while doing nothing.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Istio installed** -- the Istio CRDs (deploy **Istio Base CRDs**) must be present and the Istio control plane (istiod) running. The rule is only honored where istiod is active.
- **Target namespace exists** -- the rule is created in a specific namespace; reference an existing one or create it first.
- **The host in the registry** -- a Kubernetes Service for in-mesh destinations, or a ServiceEntry for external hosts (only needed for egress rules).

## Deploy

### Console

Open the deployment store, find **Istio Destination Rule**, and click **Deploy**. The creation wizard walks you through the namespace, the destination host, the traffic policy, and any subsets. Start from the **Circuit Breaking & Outlier Detection** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesDestinationRule
metadata:
  name: reviews-circuit-breaker
  org: acme-corp
  env: prod
spec:
  namespace:
    value: prod-apps
  host:
    value: reviews.prod-apps.svc.cluster.local
  trafficPolicy:
    loadBalancer:
      simple: LEAST_REQUEST
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http2MaxRequests: 1000
        maxRequestsPerConnection: 10
    outlierDetection:
      consecutive5xxErrors: 5
      interval: 10s
      baseEjectionTime: 30s
      maxEjectionPercent: 50
```

```shell
planton apply -f destination-rule.yaml
```

This caps how many connections and requests callers push at `reviews`, and ejects any host that returns five consecutive 5xx errors -- classic circuit breaking that keeps a struggling backend from taking down its callers. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the host to the Service whose traffic this rule shapes:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: prod-apps-namespace
      fieldPath: spec.name
  host:
    valueFrom:
      kind: KubernetesService
      name: reviews
      fieldPath: status.outputs.kube_endpoint
```

The InfraPipeline deploys the namespace and Service first, then provisions the DestinationRule against the resolved host -- the rule can never drift from the Service's actual name.

## Key Configuration

These are the most important decisions when configuring a DestinationRule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Host: reference or literal** -- referencing a Planton-managed Kubernetes Service through `valueFrom` orders this rule after the Service and pins the rule to its real in-cluster FQDN. A literal `value:` covers everything else: ServiceEntry hosts, external FQDNs, wildcards. Prefer fully-qualified hosts either way -- istiod resolves short names relative to this rule's namespace, and a rule that resolves against the wrong namespace applies to nothing.

**Per-port overrides fully replace, never merge** -- a `portLevelSettings` entry replaces the entire destination-level policy for that port; there is no field-level merge. Set everything a port needs inside its override, or the destination-level settings you thought applied silently stop applying on that port.

**Outlier detection is the circuit breaker** -- `consecutive5xxErrors`, the scan `interval`, `baseEjectionTime`, and `maxEjectionPercent` decide when a failing endpoint is ejected and for how long. Cap `maxEjectionPercent` below 100 unless you are prepared for the rule to eject every endpoint of a uniformly failing service and fail all traffic outright.

**Connection-pool limits protect the caller, not the callee** -- `tcp.maxConnections` and the HTTP request caps bound what the *client-side* sidecars will push. Undersized pools surface as 503s under load with a healthy backend -- size them from observed peak concurrency, not defaults.

**Subsets only bite when routed** -- a subset's traffic policy takes effect once a route rule actually sends traffic to that subset by name. Defining subsets without the accompanying routing is a silent no-op; plan the two together for canary and blue/green rollouts.

**Upstream TLS mode is an egress decision** -- `tls.mode` (DISABLE / SIMPLE / MUTUAL / ISTIO_MUTUAL) sets what the sidecar originates toward the destination. MUTUAL with a `credentialName` presents client certificates from a Kubernetes secret; the secret reference is runtime-only (no dependency edge), so order this rule after the secret with `metadata.relationships`.

**Scope with workloadSelector deliberately** -- `workloadSelector` is a plain label match applied at runtime by istiod, not a foreign key: it creates no dependency edge and no deploy-time validation. Without it the rule applies to every sidecar in the namespace that talks to the host.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesService** | `host` | `status.outputs.kube_endpoint` |

### What This Kind Provides

A DestinationRule is a policy resource consumed by istiod -- it has no controller-reconciled status worth exporting. `status.outputs` carries only the resource identity (`destination_rule_name`, `namespace`), both echoes of what the manifest declared; downstream resources have nothing to consume from it.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Circuit breaking and outlier detection** -- cap connection-pool size and eject hosts that keep returning errors, so a struggling backend cannot take down its callers and the mesh routes around unhealthy endpoints. Start from the **Circuit Breaking & Outlier Detection** preset.

**mTLS origination to an egress host** -- have the sidecar originate mutual TLS to an external service, presenting client certificates loaded from a Kubernetes secret, so in-mesh workloads can reach a partner API or managed database that requires client-cert auth. Pair with a ServiceEntry that brings the host into the registry. Start from the **mTLS Origination to an Egress Host** preset.

**Session affinity for stateful backends** -- consistent-hash load balancing on a cookie or header pins each client to one endpoint, trading balanced load for cache locality and sticky sessions.

## Works With

- [**Istio Base CRDs**](/infra-catalog/kubernetes-istio-base-crds) -- the prerequisite CRDs the DestinationRule kind is defined by
- [**Istio**](/infra-catalog/kubernetes-istio) -- the control plane (istiod) that reads the rule and programs the sidecars
- [**Kubernetes Service**](/infra-catalog/kubernetes-service) -- the in-mesh destination whose traffic the rule shapes; the `host` foreign key targets it
- [**Istio Service Entry**](/infra-catalog/kubernetes-service-entry) -- brings an external host into the registry so this rule can configure egress to it
- [**Istio Peer Authentication**](/infra-catalog/kubernetes-peer-authentication) -- the mTLS acceptance side; a shared workload selector scopes both to the same sidecars
