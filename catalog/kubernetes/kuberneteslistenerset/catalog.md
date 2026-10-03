# Kubernetes ListenerSet

Creates a namespaced Kubernetes Gateway API `ListenerSet` -- a set of additional listeners **merged into an existing Gateway**. ListenerSets are the per-tenant/per-team delegation model for shared gateways: a platform team runs the Gateway centrally, and application teams attach their own listeners (ports, hostnames, TLS certificates) from their own namespaces -- without editing the Gateway itself. The listener entries carry the same shape as a Gateway's own listeners (port, protocol, optional hostname and TLS, route-attachment policy). This component mirrors the upstream Gateway API `ListenerSet` (standard channel from v1.5, `gateway.networking.k8s.io/v1`) spec with full fidelity while adding proto validation, typed SDKs, and InfraChart composability.

> **Attachment is opt-in.** Gateways allow **no** ListenerSet attachment by default: the parent Gateway must explicitly allow it through its `allowedListeners` configuration, naming which namespaces may attach.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A namespaced ListenerSet** named after `metadata.name` in `spec.namespace`, attached to the Gateway in `spec.parentRef`, merging the listeners declared in `spec.listeners` into it.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

The Gateway controller merges listeners from the Gateway and all attached ListenerSets -- the Gateway's own listeners take precedence, then ListenerSets by creation time, then alphabetical namespace/name order -- and reports per-listener `Accepted` / `Programmed` conditions in the ListenerSet's status, observed with `kubectl`.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Gateway API CRDs at v1.5.0+ installed** -- deploy the `KubernetesGatewayApiCrds` component first (ListenerSet joined the standard channel in v1.5; the catalog targets v1.6.1).
- **A parent Gateway that allows attachment** -- `spec.parentRef` should resolve to a `KubernetesGateway` whose `allowedListeners` policy admits this ListenerSet's namespace.
- **The target namespace exists** -- `spec.namespace` should resolve to a real `KubernetesNamespace`.
- **TLS certificate Secrets exist** -- terminating listeners reference `kubernetes.io/tls` Secrets, resolved in the ListenerSet's OWN namespace without a `ReferenceGrant` (not the parent Gateway's).

## Deploy

### Console

Open the deployment store, find **Kubernetes ListenerSet**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and three spec steps: **Namespace** (immutable -- the delegation boundary), then **Parent Gateway** (the Gateway these listeners merge into), then **Listeners** (the endpoints to add, with the same editor a Gateway's own listeners use). Start from the **Team HTTPS Listeners** or **TLS Passthrough Listener** preset in the [Presets](#presets) tab for a directly deployable configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesListenerSet
metadata:
  name: team-a-listeners
  org: acme-corp
  env: prod
spec:
  namespace:
    value: team-a
  parentRef:
    name:
      value: shared-gateway
    namespace: platform-ingress
  listeners:
    - name: team-a-https
      port: 443
      protocol: HTTPS
      hostname: team-a.example.com
      tls:
        certificateRefs:
          - name:
              value: team-a-tls
```

```shell
planton apply -f listener-set.yaml
```

This merges an HTTPS listener for `team-a.example.com` into `shared-gateway` (in `platform-ingress`), terminating with the `team-a-tls` Secret from the ListenerSet's own `team-a` namespace. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the ListenerSet to its namespace, parent Gateway, and certificate:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: team-a-namespace
      fieldPath: spec.name
  parentRef:
    name:
      valueFrom:
        kind: KubernetesGateway
        name: shared-gateway
        fieldPath: status.outputs.gateway_name
    namespace: platform-ingress
  listeners:
    - name: team-a-https
      port: 443
      protocol: HTTPS
      hostname: team-a.example.com
      tls:
        certificateRefs:
          - name:
              valueFrom:
                kind: KubernetesCertificate
                name: team-a-cert
                fieldPath: status.outputs.secret_name
```

The InfraPipeline deploys the namespace, Gateway, and certificate first, then attaches the ListenerSet.

## Key Configuration

These are the most important decisions when configuring a ListenerSet. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Namespace** -- The `namespace` field is where the ListenerSet lives -- typically the team's own namespace. It is **immutable** and carries the delegation model's weight: the parent Gateway must allow attachment FROM this namespace, and TLS certificate Secrets referenced by the listeners resolve here without a `ReferenceGrant`. Reference an existing `KubernetesNamespace` or type the name directly.

**Parent Gateway** -- The singular `parentRef` names the Gateway these listeners merge into. Unlike a route's parentRef it has no `sectionName` or `port` -- a ListenerSet always attaches to the Gateway as a whole. The `name` is a foreign key to `KubernetesGateway`: reference a Planton-managed Gateway (the ListenerSet then deploys after it), or pass a literal name for one created outside Planton.

**Listeners** -- 1-64 `listeners`, each a port + protocol endpoint with an optional hostname (HTTP/HTTPS/TLS only), TLS termination for HTTPS/TLS protocols, and a route-attachment policy. Names need only be unique within this ListenerSet; across the merged result, each listener must have a unique port + protocol + hostname combination (the Gateway's own listeners win conflicts). Certificate references are foreign keys to `KubernetesSecret` -- typically a cert-manager `KubernetesCertificate`'s issued Secret via its `secret_name` output.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesGateway** | `parentRef.name` | `status.outputs.gateway_name` |
| **KubernetesSecret** | `listeners[].tls.certificateRefs[].name` | `status.outputs.secret_name` |

Literal names cover targets created outside Planton; certificate references typically wire against a KubernetesCertificate's `status.outputs.secret_name` instead, terminating with cert-manager-issued material the moment it is issued.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `listener_set_name` | Name of the created ListenerSet (equals `metadata.name`) | Routes name it as a parentRef (with `sectionName` for one listener) and order themselves after it |
| `namespace` | The resolved namespace the ListenerSet was created in | Same-namespace / ReferenceGrant rules for its references |
| `gateway_name` | The resolved parent Gateway name | Composition and auditing of the delegation chain |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team HTTPS listeners** -- A team merges its own HTTPS endpoint (its hostname, its certificate Secret) into the platform's shared Gateway. Start from the **Team HTTPS Listeners** preset.

**TLS passthrough listener** -- Add a TLS listener in Passthrough mode so an in-cluster backend terminates TLS itself (paired with a `KubernetesTlsRoute`). Start from the **TLS Passthrough Listener** preset.

## Works With

- [**Kubernetes Gateway API CRDs**](/infra-catalog/kubernetes-gateway-api-crds) -- installs the Gateway API CRDs (v1.5.0+ standard channel carries ListenerSet); deploy first
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- the parent Gateway (`spec.parentRef`) whose `allowedListeners` policy must admit this namespace; install first
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the namespace (`spec.namespace`) the ListenerSet runs in
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) -- issues the TLS Secrets terminating listeners reference (via its `secret_name` output)
- [**Kubernetes HTTPRoute**](/infra-catalog/kubernetes-http-route) -- attaches to this ListenerSet by naming it as a parentRef (optionally one listener via `sectionName`)
- [**Kubernetes TLSRoute**](/infra-catalog/kubernetes-tls-route) -- pairs with a Passthrough TLS listener so the backend terminates TLS itself
- [**Kubernetes TCPRoute**](/infra-catalog/kubernetes-tcp-route) -- attaches raw TCP forwarding to a merged listener
- [**Kubernetes UDPRoute**](/infra-catalog/kubernetes-udp-route) -- attaches UDP forwarding to a merged listener
- [**Kubernetes GRPCRoute**](/infra-catalog/kubernetes-grpc-route) -- attaches gRPC method routing to a merged HTTPS listener
