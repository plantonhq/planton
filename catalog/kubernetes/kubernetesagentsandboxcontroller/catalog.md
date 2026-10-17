# Agent Sandbox Controller

Installs agent-sandbox, the Kubernetes SIG Apps controller that gives every AI agent its own isolated, stateful sandbox: one pod with a stable identity, optional persistent volumes, and a lifecycle the controller manages (create, suspend with volumes kept, expire). Turn on the extensions for warm pools and claims, which is how upstream's Python and Go SDKs hand an agent a ready sandbox in seconds, and add the router so callers reach a sandbox by name.

This kind installs the **engine**, once per cluster. Sandboxes themselves are declared with the **Agent Sandbox**, **Agent Sandbox Template**, **Agent Sandbox Warm Pool** and **Agent Sandbox Claim** kinds, or created by your product through upstream's SDKs.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions, from upstream's own v1.0.7 release:

- **The `agent-sandbox-system` namespace** (with `createNamespace: true`) — the name is fixed by upstream: its manifests, and the network policy every SandboxTemplate gets, name it. Exactly one installation per cluster. Leave `createNamespace` off when a **Kubernetes Namespace** resource owns it.
- **The resource definitions** — `Sandbox`, and with extensions on also `SandboxTemplate`, `SandboxWarmPool` and `SandboxClaim`. They are applied as kept, ownership-checked resources before anything else, so destroying the controller never deletes a sandbox, its pod or its volume, a reinstall adopts them again, and an older release than the one that wrote them is refused as a downgrade.
- **The controller** — the `agent-sandbox-controller` Deployment, its ServiceAccount, RBAC and metrics Service, with every typed setting applied onto upstream's own arguments. The deploy waits until the controller is rolled out and ready.
- **The router** (when `router` is set) — the `sandbox-router` Deployment, its Service (`sandbox-router-svc`), PodDisruptionBudget, RBAC and NetworkPolicy. It is the HTTP entry point that forwards a request to the sandbox named in its `X-Sandbox-ID` header, and the only access path to sandboxes running under gVisor or Kata.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **No other agent-sandbox installation** — on GKE, check that the managed **Agent Sandbox add-on** is off: it installs the same controller in the same namespace. Run the add-on or this kind on a cluster, never both. The deploy refuses to install over definitions or a controller it did not create, and names the owner.
- **A sandbox runtime (optional)** — sandboxes run on the cluster's default runtime unless a template names a RuntimeClass such as gVisor. On GKE, enable GKE Sandbox on a node pool first.

## Deploy

### Console

Open the deployment store, find **Agent Sandbox Controller**, and click **Deploy**. The creation wizard walks you through the extensions, the controller's settings, the router and its authorization, TLS, and image sourcing. Start from the **Core preset**, or the **SDK ready preset** if your product uses upstream's SDKs.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSandboxController
metadata:
  name: agent-sandbox
  org: acme-corp
  env: dev
spec:
  namespace:
    value: agent-sandbox-system
  createNamespace: true
  extensions: {}
  router:
    authorization:
      tokenReview:
        requireToken: true
```

```shell
planton apply -f agent-sandbox.yaml
```

This installs the controller with warm pools and claims, and a router that admits a caller only when it presents a Kubernetes ServiceAccount token the cluster accepts. An Infra Job tracks the provisioning in real time.

### InfraChart

Serve the router over TLS. A **Kubernetes Namespace** named `agent-sandbox-system` owns the namespace, a **Kubernetes Certificate** is issued into it, and the controller references both:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: agent-sandbox-system
      fieldPath: spec.name
  router:
    authorization:
      tokenReview:
        requireToken: true
    tls:
      certificateSecret:
        valueFrom:
          kind: KubernetesCertificate
          name: sandbox-router-tls
          fieldPath: status.outputs.secret_name
      httpEnabled: false
```

The InfraPipeline creates the namespace, issues the certificate into it, then installs the controller and points the router at the Secret. Destroying the controller later leaves the namespace and the certificate in place.

## Key Configuration

These are the decisions that matter most. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Who owns the namespace** — the name is always `agent-sandbox-system`; what you choose is its owner. `createNamespace: true` lets this resource create it and delete it on destroy, the right shape when nothing else lives there. Reference a **Kubernetes Namespace** instead the moment anything this kind reads must be there first: the router's TLS Secret, scoped-token keys, a client CA, image pull secrets. The deploy refuses a namespace that is missing while another resource owns it, and one someone else created while this resource would create it.

**No version field, on purpose** — the kind installs agent-sandbox **v1.0.7**. The catalog's Agent Sandbox, Agent Sandbox Template, Agent Sandbox Warm Pool and Agent Sandbox Claim kinds mirror these definitions field for field, so the controller version and the typed resources move together, in a deliberate catalog update.

**Extensions** — absent means core only: `Sandbox` alone, which is upstream's default and all a product needs when it creates sandboxes directly. Set `extensions: {}` for templates, warm pools and claims. `allowedLabelDomains` decides which labels a claim may put on its sandbox's pod; the list replaces upstream's `sandbox.users.io`, so include it to keep it.

**Router authorization is chosen, never inherited** — upstream's router lets every request through by default, so this kind makes you pick one mode: `allowAll` (the cluster network and the router's NetworkPolicy are the only fence), `tokenReview` (a caller presents a Kubernetes ServiceAccount token, checked with TokenReview, optionally against `audiences`), or `scopedToken` (per-sandbox signed tokens, verified with keys from a Secret). A browser path adds `cookie`, which needs `proxy.pathRoutingPrefix`.

**Router TLS** — `tls.certificateSecret` names a TLS Secret in `agent-sandbox-system`, typically from a **Kubernetes Certificate**. The router then serves HTTPS on 8443; set `httpEnabled: false` to close the plain listener, and `mtlsMode` with `clientCa` to require client certificates.

**Reaching your sandboxes** — the router's NetworkPolicy is on by default and opens only port 8888 to sandboxes. List every port your sandboxes serve in `networkPolicy.sandboxPorts`, and narrow callers with `clientNamespaceLabels`. The policy only takes effect on a cluster whose network enforces NetworkPolicy.

**Running more than a demo** — `controller.replicas` adds leader-elected standbys for failover, not throughput; throughput comes from the concurrency settings under `extensions.claim`, `extensions.warmPool` and `controller.kubeApi`. The router defaults to two replicas with a disruption budget of one.

**Mirrors and air gaps** — `imageRegistry` moves both images (controller and router, published on registry.k8s.io) to your registry, keeping each path and tag. Sandbox images are your own and never move.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesCertificate** | `router.tls.certificateSecret` | `status.outputs.secret_name` |
| **KubernetesCertificate** | `controller.metrics.tls.certificateSecret` | `status.outputs.secret_name` |
| **KubernetesSecret** | `router.authorization.scopedToken.verificationKeys.secret` | `status.outputs.secret_name` |
| **KubernetesSecret** | `router.tls.clientCa.secret` | `status.outputs.secret_name` |
| **KubernetesSecret** | `router.telemetry.headers.secret` | `status.outputs.secret_name` |
| **KubernetesSecret** | `controller.tracing.headers.secret` | `status.outputs.secret_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | `agent-sandbox-system` | Composition, debugging |
| `metrics_service_name` | The controller's metrics Service | A Prometheus ServiceMonitor |
| `metrics_port` | The controller's metrics port | A Prometheus ServiceMonitor |
| `router_service_name` | `sandbox-router-svc` (empty without a router) | A Gateway route to the router |
| `router_endpoint` | The router's in-cluster URL (empty without a router) | Your product's SDK configuration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Core** — the controller alone, serving `Sandbox`: for a product that creates sandboxes directly. Start from the **Core preset**.

**SDK ready** — extensions on and an open router: what upstream's Python and Go SDKs need, for development and trusted networks. Start from the **SDK ready preset**.

**Production** — a warm standby controller, declared resources, three router pods behind TokenReview, and callers fenced to labelled namespaces. Start from the **Production preset**.

**Router TLS** — the router serving HTTPS only, from a cert-manager certificate, on TLS 1.3. Start from the **Router TLS preset**.

## Works With

- [**Agent Sandbox**](/infra-catalog/kubernetes-agent-sandbox) — one long-lived sandbox this controller runs.
- [**Agent Sandbox Template**](/infra-catalog/kubernetes-agent-sandbox-template) — the hardened blueprint, and network fence, of the sandboxes a product claims (extensions on).
- [**Agent Sandbox Warm Pool**](/infra-catalog/kubernetes-agent-sandbox-warm-pool) — ready sandboxes from a template, which this controller keeps filled.
- [**Agent Sandbox Claim**](/infra-catalog/kubernetes-agent-sandbox-claim) — a sandbox taken from a pool, which this controller binds.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — own `agent-sandbox-system` with its quotas and policies, so the Secrets this kind reads can be placed there first.
- [**Kubernetes Certificate**](/infra-catalog/kubernetes-certificate) — the TLS certificate the router (or the controller's metrics endpoint) serves.
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) — the scoped-token verification keys, a client CA bundle, and OTLP collector headers.
- [**Kubernetes Service Monitor**](/infra-catalog/kubernetes-service-monitor) — scrape the controller's metrics Service.
- [**GKE Cluster**](/infra-catalog/gcp-gke-cluster) — on GKE, use this kind or the cluster's Agent Sandbox add-on, never both.
