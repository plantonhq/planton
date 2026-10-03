# Kubernetes Backend TLS Policy

Provision a Kubernetes Gateway API `BackendTLSPolicy` -- the standard way to
tell a Gateway implementation to originate TLS to the backends BEHIND the
gateway and verify the certificate they present. Routes decide WHERE traffic
goes; this policy decides HOW the gateway-to-backend hop is secured. Use it
for end-to-end encryption: TLS terminates at the gateway and is
re-originated -- verified -- to the backend.

BackendTLSPolicy is a standard-channel resource served as
`gateway.networking.k8s.io/v1` (Gateway API v1.6.1; the `v1alpha3` version is
deprecated upstream and no longer served). It is a Direct policy attachment:
it targets Services (Core support) in ITS OWN namespace, optionally narrowed
to a single named Service port through `sectionName`.

## What Gets Created

- A namespaced `gateway.networking.k8s.io/v1` `BackendTLSPolicy` custom
  resource.
- One to 16 target references to same-namespace Services (upstream note:
  implementations SHOULD support a single targetRef -- one is the safest
  portable shape).
- A validation block carrying the trust anchor (a CA-bundle ConfigMap XOR
  the system trust store), the SNI/authentication hostname, and optional
  Subject Alternative Names.

## Prerequisites

- Gateway API standard-channel CRDs installed (`KubernetesGatewayApiCrds`).
- A Gateway implementation that honors BackendTLSPolicy (support varies --
  see the policy's per-ancestor `Accepted` condition after a controller
  reconciles it).
- The backend Service the policy targets, in the same namespace.
- For the bring-your-own-CA arm: a same-namespace ConfigMap carrying the
  PEM CA bundle in a key named `ca.crt` (exactly what a cert-manager CA
  chain materializes).

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesBackendTlsPolicy
metadata:
  name: my-backend-tls
spec:
  namespace:
    value: app-ns
  targetRefs:
    - group: "" # Services live in the core API group — empty but present
      kind: Service
      name:
        value: my-backend
  validation:
    wellKnownCACertificates: System
    hostname: api.example.com
```

```bash
planton apply -f backendtlspolicy.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace to create the policy in. The policy can only target Services in its own namespace. |
| `targetRefs` | list | 1-16 references to same-namespace backend Services. `group` must be present (empty string for the core group); `kind` is `Service` for Core support; `name` is a reference (defaults to `KubernetesService`). |
| `validation` | object | The trust anchor, hostname, and optional SANs. |
| `validation.hostname` | string | Sent as the SNI for the backend connection and -- unless `subjectAltNames` is set -- the identity the backend certificate must prove. |

Exactly one trust-anchor arm is required inside `validation` (the spec
mirrors the CRD's own CEL rules):

| Arm | Description |
|-----|-------------|
| `caCertificateRefs` | 1-8 same-namespace objects carrying the PEM CA bundle. Core support: ONE ConfigMap with the bundle in a key named `ca.crt`. Each `name` is a reference (defaults to `KubernetesConfigMap`). |
| `wellKnownCACertificates` | `System` (the one upstream-defined value) trusts the implementation's system certificate store -- for backends serving publicly-issued certificates. Implementation-specific support. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `targetRefs[].sectionName` | string | A Service PORT NAME -- narrows the policy to connections to that port; omit to cover every port. A name that does not exist on the target makes the policy fail to attach (ResolvedRefs condition). |
| `validation.subjectAltNames` | list | Up to 5 SANs the backend certificate must contain at least one of -- `type: Hostname` with `hostname`, or `type: URI` with `uri` (SPIFFE IDs are the common URI case). With SANs set, `hostname` only selects the certificate (SNI). Extended support. |
| `options` | map | Up to 16 implementation-specific TLS options (domain-prefixed keys recommended), e.g. a vendor's minimum-TLS-version knob. |

## Examples

### Internal CA via ConfigMap

```yaml
spec:
  namespace:
    value: app-ns
  targetRefs:
    - group: ""
      kind: Service
      name:
        value: my-backend
  validation:
    caCertificateRefs:
      - group: "" # ConfigMaps live in the core API group — empty but present
        kind: ConfigMap
        name:
          value: internal-ca-bundle
    hostname: backend.internal.example.com
```

### SPIFFE mTLS backend (port-scoped, URI SAN)

```yaml
spec:
  namespace:
    value: app-ns
  targetRefs:
    - group: ""
      kind: Service
      name:
        value: my-backend
      sectionName: https # only the named port gets the TLS treatment
  validation:
    caCertificateRefs:
      - group: ""
        kind: ConfigMap
        name:
          value: mesh-trust-bundle
    hostname: backend.internal.example.com # SNI only — SANs authenticate below
    subjectAltNames:
      - type: URI
        uri: spiffe://cluster.example.com/ns/payments/sa/backend
```

## Composing in Infra Charts

`KubernetesBackendTlsPolicy` sits next to the backend it secures: it targets
a Service and (in the internal-CA arm) references the ConfigMap carrying the
CA bundle. Every one of those neighbor references is a foreign key --
`namespace` (defaults to `KubernetesNamespace`), `targetRefs[].name`
(defaults to `KubernetesService`), and `caCertificateRefs[].name` (defaults
to `KubernetesConfigMap`) -- so wiring them with `valueFrom` creates real
dependency edges and the platform orders the deployment automatically:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesBackendTlsPolicy
metadata:
  name: "{{ values.env }}-backend-tls"
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: "{{ values.env }}-ns"
      fieldPath: spec.name
  targetRefs:
    - group: ""
      kind: Service
      name:
        valueFrom:
          kind: KubernetesService
          name: "{{ values.backend_service }}"
          fieldPath: status.outputs.service_name
  validation:
    caCertificateRefs:
      - group: ""
        kind: ConfigMap
        name:
          valueFrom:
            kind: KubernetesConfigMap
            name: "{{ values.ca_bundle_configmap }}"
            fieldPath: status.outputs.configmap_name
    hostname: backend.internal.example.com
```

When a target or CA referent is not Planton-managed, pass the literal name
with `value:` instead.

The policy is namespace-local by upstream rule: cross-namespace targetRefs
and CA references are invalid, so create the policy in the namespace of the
backend Services it secures.

## Outputs

| Output | Description |
|--------|-------------|
| `policy_name` | Name of the created BackendTLSPolicy (equals `metadata.name`). |
| `namespace` | Namespace the BackendTLSPolicy was created in. |

## Related Kinds

- [Kubernetes Gateway](../kubernetesgateway)
- [Kubernetes HTTP Route](../kuberneteshttproute)
- [Kubernetes Gateway API CRDs](../kubernetesgatewayapicrds)
- [Kubernetes Service](../kubernetesservice)
- [Kubernetes Config Map](../kubernetesconfigmap)
- [Kubernetes Namespace](../kubernetesnamespace)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
