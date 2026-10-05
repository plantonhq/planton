# KubernetesListenerSet

> A Kubernetes Gateway API `ListenerSet`: a namespaced set of additional listeners merged into an existing Gateway, letting teams own their listeners and TLS certificates without touching the central Gateway.

## Overview

`KubernetesListenerSet` is a first-class catalog kind that provisions an
upstream Gateway API `ListenerSet` resource at 100% fidelity with the standard
channel of Gateway API v1.6.1 (ListenerSet is standard-channel since v1.5). It
is the per-team delegation model for shared gateways: a platform team runs one
central `Gateway`, and each tenant team attaches its own listeners -- ports,
hostnames, TLS certificates, route-attachment policy -- through a ListenerSet in
its own namespace.

The parent Gateway must explicitly opt in through its `allowed_listeners`
configuration; Gateways allow **no** ListenerSet attachment by default. Routes
can then attach to the ListenerSet directly (parentRef `kind: ListenerSet`,
optionally with `sectionName` targeting one listener).

Unlike a raw `KubernetesManifest`, this kind gives you proto validation,
foreign-key wiring (to `KubernetesNamespace`, `KubernetesGateway`, and the
`KubernetesSecret` objects its TLS references point at), typed Pulumi and
Terraform modules, and InfraChart composability.

## Prerequisites

- The Gateway API CRDs are installed on the target cluster
  (`KubernetesGatewayApiCrds`; ListenerSet requires v1.5.0 or newer).
- A parent `Gateway` (`KubernetesGateway`) whose `allowed_listeners` permits
  attachment from this ListenerSet's namespace (e.g. `namespaces.from: Same`).
- The target namespace exists (`KubernetesNamespace`).
- For HTTPS listeners, a `kubernetes.io/tls` Secret with the certificate/key
  (commonly produced by a cert-manager `KubernetesCertificate`) in the
  ListenerSet's own namespace.

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesListenerSet
metadata:
  name: tenant-listeners
spec:
  namespace:
    value: tenant-ns
  parentRef:
    name:
      value: shared-gateway
    namespace: gateway-ns
  listeners:
    - name: tenant-https
      hostname: tenant.example.com
      port: 443
      protocol: HTTPS
      tls:
        mode: Terminate
        certificateRefs:
          - name:
              value: tenant-tls
```

```bash
planton apply -f listenerset.yaml
```

## How It Works

1. The ListenerSet is created in `spec.namespace` and names its parent Gateway
   via `spec.parentRef` (a whole-Gateway reference -- no section or port). The
   Gateway must allow attachment from this namespace via `allowed_listeners`.
2. The Gateway merges listeners from itself and all attached ListenerSets. Its
   own listeners take precedence, then ListenerSets by creation time, then
   alphabetical namespace/name order.
3. Each listener binds a port and protocol; HTTPS/TLS listeners carry TLS
   configuration. Listener names must be unique within the ListenerSet only --
   not across the Gateway and its other ListenerSets.
4. Routes attach either to the parent Gateway or to the ListenerSet itself
   (parentRef `kind: ListenerSet`, optional `sectionName` for one listener),
   subject to each listener's `allowedRoutes` policy -- which for ListenerSets
   defaults to the ListenerSet's own namespace.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `namespace` | `StringValueOrRef` -> `KubernetesNamespace` | yes | Namespace to create the ListenerSet in. |
| `parent_ref` | `KubernetesGatewayApiParentGatewayReference` | yes | The Gateway the listeners attach to; `name` is an FK to `KubernetesGateway`. |
| `listeners` | `[]KubernetesListenerSetListener` | yes (1-64) | Listeners merged into the parent Gateway (name, port, protocol, tls, allowed_routes). |

### Listener fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Listener name, unique within this ListenerSet (Route `parentRef.sectionName` target). |
| `hostname` | string | no | Virtual host to match (HTTP/HTTPS/TLS); wildcard prefix allowed. |
| `port` | int32 (1-65535) | yes | Network port. |
| `protocol` | string | yes | `HTTP`, `HTTPS`, `TLS`, `TCP`, `UDP`, or a domain-prefixed custom protocol. |
| `tls.mode` | string | no | `Terminate` (default) or `Passthrough`. |
| `tls.certificate_refs` | `[]SecretObjectReference` | conditionally | TLS Secrets to terminate with (required for Terminate). Each `name` is an FK to `KubernetesSecret` -- typically wired with `valueFrom` against a `KubernetesCertificate`'s `status.outputs.secret_name`. |
| `allowed_routes` | `KubernetesGatewayApiAllowedRoutes` | no | Which Route kinds/namespaces may attach (defaults to the ListenerSet's own namespace). |

## Composing in Infra Charts

The per-team delegation pattern: the platform chart owns the Gateway (with
`allowed_listeners` opened up), and each team chart owns a ListenerSet plus its
certificate and routes. Every neighbor reference is a foreign key, so
`valueFrom` wiring creates real dependency edges and the platform orders the
deployment automatically:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesListenerSet
metadata:
  name: "{{ values.team }}-listeners"
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: "{{ values.team }}-ns"
      fieldPath: spec.name
  parentRef:
    name:
      valueFrom:
        kind: KubernetesGateway
        name: shared-gateway
        fieldPath: status.outputs.gateway_name
    namespace: gateway-ns
  listeners:
    - name: team-https
      hostname: "{{ values.team }}.example.com"
      port: 443
      protocol: HTTPS
      tls:
        mode: Terminate
        certificateRefs:
          - name:
              valueFrom:
                kind: KubernetesCertificate
                name: "{{ values.team }}-cert"
                fieldPath: status.outputs.secret_name
```

When the Gateway or Secret is not Planton-managed, pass the literal name with
`value:` instead.

## Outputs

| Output | Description |
|--------|-------------|
| `listener_set_name` | Name of the created ListenerSet (equals `metadata.name`); the target of Route `parentRefs` with `kind: ListenerSet`. |
| `namespace` | Namespace the ListenerSet was created in. |
| `gateway_name` | Name of the parent Gateway the listeners attach to. |

## Related Kinds

- [`KubernetesGateway`](../kubernetesgateway/) -- the parent Gateway; must opt in via `allowed_listeners`.
- [`KubernetesGatewayApiCrds`](../kubernetesgatewayapicrds/) -- installs the Gateway API CRDs (prerequisite).
- [`KubernetesCertificate`](../kubernetescertificate/) -- provisions the TLS Secret HTTPS listeners reference.
- [`KubernetesHttpRoute`](../kuberneteshttproute/) -- routes that attach to the merged listeners.
- [`KubernetesNamespace`](../kubernetesnamespace/) -- the namespace the ListenerSet is created in.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
