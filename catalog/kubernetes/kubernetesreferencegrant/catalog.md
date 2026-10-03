# Kubernetes ReferenceGrant

Creates a namespaced Kubernetes Gateway API `ReferenceGrant` -- a runtime authorization that permits resources in *other* namespaces to reference specified kinds of resources in *this* grant's namespace. In the Gateway API, every cross-namespace reference (a Gateway's TLS `certificateRefs`, a Route's `backendRefs`, and similar) is denied by default; a ReferenceGrant placed in the *referenced* ("to") namespace is what explicitly authorizes it. This component mirrors the upstream Gateway API v1 `ReferenceGrant` spec with full fidelity while adding proto validation, typed SDKs, and InfraChart composability.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A namespaced ReferenceGrant** named after `metadata.name` in `spec.namespace`, declaring the trusted sources (`from`) that may reference the permitted targets (`to`) in that namespace.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

ReferenceGrant has no controller-managed status upstream (the Gateway API project deliberately omitted it), so there is no `Accepted`/`Programmed` condition to wait on -- the grant takes effect as soon as it exists.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Gateway API CRDs installed** -- deploy the **Kubernetes Gateway API CRDs** component first. ReferenceGrant is a Gateway API type and will not register without the CRDs.
- **The target namespace exists** -- `spec.namespace` should resolve to a real `KubernetesNamespace`. Reference an existing one or type the namespace name directly.

## Deploy

### Console

Open the deployment store, find **Kubernetes ReferenceGrant**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and three spec steps: **Namespace** (the immutable "to" namespace), **Sources** (the trusted `from` entries), and **Targets** (the referenceable `to` entries). Start from the **Allow a Gateway to Reference TLS Secrets in Another Namespace** or **Allow Routes to Reference Backend Services in Another Namespace** preset in the [Presets](#presets) tab for a directly deployable configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesReferenceGrant
metadata:
  name: allow-gateway-secret-ref
  org: acme-corp
  env: prod
spec:
  namespace:
    value: cert-manager
  from:
    - group: gateway.networking.k8s.io
      kind: Gateway
      namespace: istio-ingress
  to:
    - group: ""
      kind: Secret
```

```shell
planton apply -f reference-grant.yaml
```

This creates a ReferenceGrant in `cert-manager` that authorizes Gateways in `istio-ingress` to reference Secrets there. An Infra Job tracks the provisioning in real time.

### InfraChart

When the "to" namespace is itself Planton-managed, wire the grant to it so the InfraPipeline orders the namespace first:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: cert-manager-namespace
      fieldPath: spec.name
  from:
    - group: gateway.networking.k8s.io
      kind: Gateway
      namespace: istio-ingress
  to:
    - group: ""
      kind: Secret
```

The InfraPipeline resolves the namespace reference and creates the grant inside it.

## Key Configuration

These are the most important decisions when configuring a ReferenceGrant. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The grant lives in the TARGET namespace, not the source** -- the non-obvious part of ReferenceGrant. `namespace` is the namespace being referenced *into* (the "to" side): a `Gateway` in `istio-ingress` needing a TLS `Secret` in `cert-manager` means the grant is created in `cert-manager`, with `from` naming the Gateway's namespace and kind and `to` naming `kind: Secret`. Deleting the grant revokes the access. The namespace is also **immutable**: the trust boundary is defined relative to it, so changing it means creating a different grant.

**Sources (`from`) are kind-level trust, combined with OR** -- each `{ group, kind, namespace }` tuple (1-16 entries) trusts every resource of that kind in that source namespace, not one specific object. `kind` and `namespace` are required; `group` is `gateway.networking.k8s.io` for Gateways and Routes, or empty `""` for core kinds. Grant per kind and per namespace deliberately -- a broad `from` list is a broad trust statement.

**Targets (`to`) default to ALL resources of the kind** -- each `{ group, kind, name? }` tuple (1-16) with `name` empty authorizes referencing every resource of that kind in the grant's namespace. Set `name` to narrow the grant to a single Secret or Service when the target namespace holds resources of the same kind that should stay unreferenceable.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

The `from`/`to` entries are kind-level trust assertions (not pointers to specific resource instances), so they create no deploy-ordering dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `reference_grant_name` | Name of the created ReferenceGrant (equals `metadata.name`) | Lets Gateways/Routes making the cross-namespace reference order themselves after the grant in an InfraChart |
| `namespace` | The resolved "to" namespace the grant was created in | Confirming which namespace's resources the grant authorizes inbound references to |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Allow a Gateway to reference TLS Secrets** -- A Gateway terminates TLS using a certificate Secret in another namespace (typically the cert-manager namespace). Create the grant in the Secret's namespace, with `from` the Gateway's namespace/kind and `to` `kind: Secret`. Start from the **Allow a Gateway to Reference TLS Secrets in Another Namespace** preset.

**Allow Routes to reference backend Services** -- HTTP/gRPC routes in an application namespace forward traffic to backend Services in another namespace. Create the grant in the backend namespace, with one `from` entry per trusted route kind and `to` `kind: Service`. Start from the **Allow Routes to Reference Backend Services in Another Namespace** preset.

## Works With

- [**Kubernetes Gateway API CRDs**](/infra-catalog/kubernetes-gateway-api-crds) -- installs the Gateway API CRDs (prerequisite, install first).
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the "to" namespace (`spec.namespace`) and the source namespaces named in `from`.
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- a common `from` source when referencing cross-namespace TLS Secrets.
- [**Kubernetes HTTPRoute**](/infra-catalog/kubernetes-http-route), [**Kubernetes GRPCRoute**](/infra-catalog/kubernetes-grpc-route), [**Kubernetes TLSRoute**](/infra-catalog/kubernetes-tls-route), and [**Kubernetes TCPRoute**](/infra-catalog/kubernetes-tcp-route) -- common `from` sources when referencing cross-namespace backend Services.
