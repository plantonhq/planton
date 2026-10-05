# Kubernetes TLS Route

Provision a Kubernetes Gateway API `TLSRoute` -- namespaced TLS passthrough rules
that attach to a Gateway and forward connections, by SNI hostname, to backend
Services. The backend (not the Gateway) terminates TLS, so the encrypted stream
is forwarded end to end.

TLSRoute is a standard-channel resource served as
`gateway.networking.k8s.io/v1` (standard since Gateway API v1.5).

## What Gets Created

- A namespaced `gateway.networking.k8s.io/v1` `TLSRoute` custom resource.
- Exactly one rule (the upstream maximum for a TLSRoute) that forwards to one or
  more weighted backend refs.

## Prerequisites

- Gateway API CRDs installed on the cluster (`KubernetesGatewayApiCrds`).
- A `Gateway` to attach to via `parentRefs` (`KubernetesGateway`) with a listener
  of protocol `TLS` (typically `tls.mode: Passthrough`).
- The target namespace (`KubernetesNamespace`).
- The backend Services the route forwards to.

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTlsRoute
metadata:
  name: secure-route
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name:
        value: my-gateway
      sectionName: tls
  hostnames:
    - secure.example.com
  rules:
    - backendRefs:
        - name:
            value: secure-backend
          port: 8443
```

```bash
planton apply -f tlsroute.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace to create the route in. |
| `hostnames` | list | One to 1024 SNI hostnames that select this route (no IPs). |
| `rules` | list | Exactly one routing rule. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `parentRefs` | list | Gateways (and optional listener `sectionName`) the route attaches to. Each `name` is a reference (defaults to `KubernetesGateway`). |
| `rules[].name` | string | Optional rule name (unique within the route). |
| `rules[].backendRefs` | list | One to 16 weighted backends; each `name` is a reference (defaults to `KubernetesService`). |

## Examples

### TLS passthrough by SNI

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name:
        value: my-gateway
      sectionName: tls
  hostnames:
    - secure.example.com
  rules:
    - backendRefs:
        - name:
            value: secure-backend
          port: 8443
```

### Weighted backends (canary)

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name:
        value: my-gateway
      sectionName: tls
  hostnames:
    - secure.example.com
  rules:
    - backendRefs:
        - name:
            value: secure-stable
          port: 8443
          weight: 90
        - name:
            value: secure-canary
          port: 8443
          weight: 10
```

## Composing in Infra Charts

`KubernetesTlsRoute` is a leaf in the ingress DAG: it attaches to a `Gateway` and
forwards to backend Services. Every neighbor reference is a foreign key --
`namespace`, `parentRefs[].name` (defaults to `KubernetesGateway`), and
`backendRefs[].name` (defaults to `KubernetesService`) -- so wiring them with
`valueFrom` creates real dependency edges and the platform orders the deployment
automatically:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTlsRoute
metadata:
  name: "{{ values.env }}-secure-route"
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: "{{ values.env }}-ns"
      fieldPath: spec.name
  parentRefs:
    - name:
        valueFrom:
          kind: KubernetesGateway
          name: "{{ values.env }}-gateway"
          fieldPath: status.outputs.gateway_name
      sectionName: tls
  hostnames:
    - "secure.{{ values.domain }}"
  rules:
    - backendRefs:
        - name:
            valueFrom:
              kind: KubernetesService
              name: "{{ values.service_name }}"
              fieldPath: status.outputs.service_name
          port: 8443
```

When a parent or backend is not Planton-managed, pass the literal name with
`value:` instead.

Full ingress stack DAG:

```
KubernetesCertManager -> KubernetesClusterIssuer -> KubernetesCertificate
   -> (Secret) -> KubernetesGateway -> KubernetesTlsRoute / KubernetesHttpRoute
```

(For pure TLS passthrough the backend holds its own certificate, so the
cert-manager prefix is optional; it applies when the Gateway terminates TLS.)

## Outputs

| Output | Description |
|--------|-------------|
| `routeName` | Name of the created TLSRoute (equals metadata.name). |
| `namespace` | Namespace the TLSRoute was created in. |

## Related Kinds

- [Kubernetes Gateway](../kubernetesgateway)
- [Kubernetes Gateway Class](../kubernetesgatewayclass)
- [Kubernetes HTTP Route](../kuberneteshttproute)
- [Kubernetes TCP Route](../kubernetestcproute)
- [Kubernetes Gateway API CRDs](../kubernetesgatewayapicrds)
- [Kubernetes Namespace](../kubernetesnamespace)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
