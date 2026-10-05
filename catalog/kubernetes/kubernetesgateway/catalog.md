# Kubernetes Gateway

Creates a namespaced Kubernetes Gateway API `Gateway` -- an instance of traffic-handling infrastructure that binds a set of **listeners** (logical endpoints with a port, protocol, and optional TLS) to addresses, programmed by the controller behind a `GatewayClass`. The Gateway is the role-oriented successor to Ingress: platform teams own the `GatewayClass`, infrastructure teams own the `Gateway`, and application teams attach `Routes` (HTTP/gRPC/TLS/TCP) to its listeners. This kind mirrors the upstream Gateway API v1 `Gateway` spec with full fidelity while adding proto validation, typed SDKs, and InfraChart composability.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A namespaced Gateway** named after `metadata.name` in `spec.namespace`, attached to the `spec.gatewayClassName` GatewayClass, with the listeners (and any requested addresses, infrastructure attributes, and gateway-wide TLS) you configure.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

The controller behind the GatewayClass reconciles the Gateway asynchronously: it assigns addresses and reports per-listener conditions in the Gateway's status, which you observe with `kubectl` (these controller-managed values are intentionally not stored as outputs).

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Gateway API CRDs installed** -- deploy the `KubernetesGatewayApiCrds` component first. Gateway is a Gateway API type and will not register without the CRDs.
- **A controller-backed GatewayClass exists** -- `spec.gatewayClassName` must resolve to a `KubernetesGatewayClass` whose controller (Istio, Envoy Gateway, NGINX, ...) is installed and running.
- **The target namespace exists** -- `spec.namespace` should resolve to a real `KubernetesNamespace`.
- **Referenced TLS Secrets exist** -- for HTTPS/TLS listeners in `Terminate` mode, the `certificateRefs` Secrets must exist in the Gateway's namespace (or be authorized cross-namespace by a `KubernetesReferenceGrant`). They are commonly produced by a cert-manager `KubernetesCertificate`.

## Deploy

### Console

Open the deployment store, find **Kubernetes Gateway**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and seven spec steps: **Namespace** (immutable) and **Gateway Class** (immutable) first, then the required **Listeners**, followed by the optional **Addresses**, **Infrastructure**, **Allowed Listeners**, and **Gateway TLS** steps. Start from the **HTTPS Gateway with TLS Termination** or **Multi-Protocol Gateway** preset in the [Presets](#presets) tab for a directly deployable configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGateway
metadata:
  name: my-https-gateway
  org: acme-corp
  env: prod
spec:
  namespace:
    value: istio-ingress
  gatewayClassName:
    value: istio
  listeners:
    - name: https
      hostname: app.acme-corp.com
      port: 443
      protocol: HTTPS
      tls:
        mode: Terminate
        certificateRefs:
          - name:
              value: app-tls
      allowedRoutes:
        namespaces:
          from: Same
        kinds:
          - kind: HTTPRoute
```

```shell
planton apply -f gateway.yaml
```

This creates a Gateway in `istio-ingress` with one HTTPS listener on port 443 that terminates TLS using the `app-tls` Secret and accepts `HTTPRoute`s from its own namespace. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the class and the TLS certificate to resources managed by other Infra Components:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: istio-ingress-namespace
      fieldPath: spec.name
  gatewayClassName:
    valueFrom:
      kind: KubernetesGatewayClass
      name: istio-class
      fieldPath: status.outputs.gateway_class_name
  listeners:
    - name: https
      hostname: app.acme-corp.com
      port: 443
      protocol: HTTPS
      tls:
        mode: Terminate
        certificateRefs:
          - name:
              valueFrom:
                kind: KubernetesCertificate
                name: app-cert
                fieldPath: status.outputs.secret_name
```

The InfraPipeline deploys the namespace, the GatewayClass, and the certificate first, then creates the Gateway — the listener terminates with the issued Secret the moment cert-manager materializes it.

## Key Configuration

These are the most important decisions when configuring a Gateway. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Namespace** -- The `namespace` field is where the Gateway runs. It is **immutable**: Routes and TLS Secrets in this namespace attach without a `ReferenceGrant`; those elsewhere require one. Reference an existing `KubernetesNamespace` or type the name directly.

**Gateway Class** -- The `gatewayClassName` field selects the `GatewayClass`, and therefore the controller that programs the Gateway. It is **immutable** (a Gateway cannot switch controllers in place) and is a foreign key to a `KubernetesGatewayClass` output, so the class is deployed before the Gateway.

**Listeners** -- One or more `listeners` (1-64). Each is a `name` + `port` + `protocol` (HTTP, HTTPS, TLS, TCP, UDP, or a domain-prefixed custom protocol), with an optional virtual `hostname`. HTTPS listeners may only TERMINATE (`mode` unset or `Terminate`, with `certificateRefs` or `options` required); a TLS-protocol listener must declare its mode explicitly (`Terminate` or `Passthrough`); HTTP/TCP/UDP listeners must carry no `tls` block at all -- the spec enforces each of these at authoring time. Each `certificateRefs` entry's `name` is a value-or-reference object (`value:` for a literal Secret name, `valueFrom:` against a KubernetesCertificate's `status.outputs.secret_name`), never a bare string. Any listener may restrict attachment via **allowedRoutes** (namespaces `from` All/Selector/Same + a label selector, and the allowed Route `kinds`).

**Addresses** -- Optional requested addresses (`type` IPAddress/Hostname/NamedAddress + `value`, up to 16). When omitted, the controller assigns addresses; set them to pin a reserved static IP or hostname.

**Infrastructure** -- Optional `labels` and `annotations` (up to 8 each) propagated to the resources the controller creates -- the standard way to tune cloud load-balancer behavior -- plus an optional per-Gateway `parametersRef`.

**Allowed Listeners** -- Optional ListenerSet attachment policy (`from` All/Selector/Same/None). Defaults to None: a Gateway must opt in before any Listener Set can merge additional listeners into it — leave it at None unless you deliberately delegate listener management.

**Gateway TLS** -- Optional gateway-wide TLS: **frontend** client-certificate validation (mTLS), with a default and optional per-port overrides, and a **backend** client certificate the Gateway presents to upstreams. Per-listener HTTPS termination is configured on each listener instead.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesGatewayClass** | `gatewayClassName` | `status.outputs.gateway_class_name` |
| **KubernetesSecret** | `listeners[].tls.certificateRefs[].name`, `tls.backend.clientCertificateRef.name` | `status.outputs.secret_name` |
| **KubernetesConfigMap** | `tls.frontend` CA bundle `caCertificateRefs[].name` | `status.outputs.configmap_name` |

Certificate Secrets are typically produced by a **Cert Manager Certificate** — reference its `status.outputs.secret_name` instead of the Secret directly. Literal names cover material created outside Planton; cross-namespace references require a ReferenceGrant.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `gateway_name` | Name of the created Gateway (equals `metadata.name`) | Routes reference this in their `parentRefs`; orders the Gateway before its Routes in an InfraChart |
| `namespace` | The resolved namespace the Gateway was created in | Same-namespace / ReferenceGrant rules for attaching Routes and Secrets |
| `gateway_class_name` | The resolved GatewayClass the Gateway belongs to | Confirming which controller programs the Gateway |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**HTTPS with TLS termination** -- A single HTTPS listener on 443 that terminates TLS using a Secret from cert-manager, accepting same-namespace `HTTPRoute`s. Start from the **HTTPS Gateway with TLS Termination** preset.

**Multi-protocol Gateway** -- An HTTP listener (often redirecting to HTTPS) alongside an HTTPS listener, sharing the Gateway's address. Start from the **Multi-Protocol Gateway** preset.

**TLS passthrough** -- A TLS listener in `Passthrough` mode that forwards the encrypted stream to a TLS Route without decrypting at the Gateway; `certificateRefs` are ignored in this mode.

## Works With

- [**Kubernetes Gateway API CRDs**](/infra-catalog/kubernetes-gateway-api-crds) -- installs the Gateway API CRDs (prerequisite, install first)
- [**Kubernetes GatewayClass**](/infra-catalog/kubernetes-gateway-class) -- the class (`gatewayClassName`) that selects the controller (install first)
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the namespace the Gateway runs in
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) -- commonly produces the TLS Secrets referenced by HTTPS listeners
- [**Kubernetes ReferenceGrant**](/infra-catalog/kubernetes-reference-grant) -- authorizes cross-namespace TLS Secret / CA references from this Gateway
- [**Kubernetes HTTPRoute**](/infra-catalog/kubernetes-http-route) -- the most common Route kind attaching to HTTP/HTTPS listeners
- [**Kubernetes GRPCRoute**](/infra-catalog/kubernetes-grpc-route) -- gRPC traffic attaching to HTTPS listeners
- [**Kubernetes TLSRoute**](/infra-catalog/kubernetes-tls-route) -- the Route kind behind Passthrough TLS listeners
- [**Kubernetes TCPRoute**](/infra-catalog/kubernetes-tcp-route) -- raw TCP forwarding from TCP listeners
- [**Kubernetes ListenerSet**](/infra-catalog/kubernetes-listener-set) -- merges additional listeners into this Gateway when `allowedListeners` opts in
