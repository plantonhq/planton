# Kubernetes BackendTLSPolicy

Defines a Gateway API BackendTLSPolicy that encrypts the hop between your gateway and the Services behind it, and verifies that the backend really is who it claims to be. A route decides WHERE traffic goes; this policy decides HOW the gateway-to-backend leg is secured — the leg most setups leave in plaintext because the client-facing side already has TLS. The spec is 100% faithful to the upstream Gateway API v1.6.1 BackendTLSPolicy (standard channel, served as `gateway.networking.k8s.io/v1`).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A BackendTLSPolicy** — a namespaced Gateway API policy attached directly to one or more Services in its own namespace. When a gateway forwards a request to a targeted Service it opens a TLS connection instead of a plaintext one, sends the configured hostname as the SNI so the backend can pick a certificate, verifies that certificate against your trust anchor (your own CA bundle, or the implementation's system store), and optionally checks the certificate's Subject Alternative Names — for backends whose identity is a SPIFFE URI rather than a DNS name.
- **Kubernetes Labels** — resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Kubernetes Cluster

- **Gateway API CRDs installed** — BackendTLSPolicy ships with the Gateway API standard channel; install the CRDs (**Kubernetes Gateway API CRDs**) before the policy.
- **A gateway controller that IMPLEMENTS BackendTLSPolicy** — support is still uneven across implementations. A policy attached behind a controller that does not implement it is accepted by the API server and then silently ignored: the gateway keeps sending plaintext, with nothing in the events to say so. Check the policy's `Accepted` and `ResolvedRefs` conditions after the first deploy.
- **Everything in one namespace** — upstream forbids cross-namespace targets here, so the policy, the Services it secures, and any CA-bundle ConfigMap all live together.

## Deploy

### Console

Open the deployment store, find **Kubernetes BackendTLSPolicy**, and click **Deploy**. The creation wizard walks you through the namespace, the Service targets (optionally narrowed to one named port), the trust-anchor choice, the SNI hostname, and optional Subject Alternative Names. Start from the **Internal CA via ConfigMap** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesBackendTlsPolicy
metadata:
  name: checkout-backend-tls
  org: acme-corp
  env: prod
spec:
  namespace:
    value: prod-apps
  targetRefs:
    - group: ""
      kind: Service
      name:
        value: checkout-api
  validation:
    caCertificateRefs:
      - group: ""
        kind: ConfigMap
        name:
          value: internal-ca-bundle
    hostname: checkout-api.prod-apps.svc.cluster.local
```

```shell
planton apply -f backend-tls-policy.yaml
```

This makes the gateway originate TLS to the `checkout-api` Service, presenting `checkout-api.prod-apps.svc.cluster.local` as the SNI and verifying the backend's certificate against the PEM bundle in the `internal-ca-bundle` ConfigMap (key `ca.crt`). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the targets to resources managed by other Infra Components — the policy then deploys after the Service it secures and the ConfigMap that carries the trust anchor:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: prod-apps-namespace
      fieldPath: spec.name
  targetRefs:
    - group: ""
      kind: Service
      name:
        valueFrom:
          kind: KubernetesService
          name: checkout-api
          fieldPath: status.outputs.service_name
  validation:
    caCertificateRefs:
      - group: ""
        kind: ConfigMap
        name:
          valueFrom:
            kind: KubernetesConfigMap
            name: internal-ca-bundle
            fieldPath: status.outputs.configmap_name
    hostname: checkout-api.prod-apps.svc.cluster.local
```

The InfraPipeline resolves the dependency graph and deploys the Service and ConfigMap before the policy that references them.

## Key Configuration

These are the most important decisions when configuring a backend TLS policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The trust anchor is exactly one of two arms — never both, never neither.** `caCertificateRefs` brings your own CA: the Core shape is ONE ConfigMap in this namespace with the PEM chain under the key `ca.crt` — exactly what a cert-manager CA chain materializes, so a root Certificate's ConfigMap (or a trust-manager Bundle target) composes directly. `wellKnownCACertificates: System` trusts the implementation's system store — the arm for backends serving publicly-issued certificates. The CRD's own validation rejects both-set and neither-set.

**Hostname does double duty.** It is sent as the SNI AND, unless `subjectAltNames` is listed, it is the identity the backend certificate must match. Once you list SANs the hostname only selects the certificate; if it should still be accepted as an identity, add it as a `Hostname` SAN too. Getting this wrong produces a handshake failure that names neither the field nor the policy.

**SANs are the SPIFFE seam.** Each `subjectAltNames` entry is `type: Hostname` (with `hostname`) or `type: URI` (with `uri` — SPIFFE IDs like `spiffe://cluster.example.com/ns/prod-apps/sa/checkout` are the common case). The type/value pairing is enforced at validation time, mirroring the CRD's own rules.

**Target one Service; narrow with a port name.** Core support targets a Service; `sectionName` narrows a reference to one named Service port (omit it to cover every port). A `sectionName` that does not exist on the target makes the policy fail to attach — surfaced through the `ResolvedRefs` condition, not an apply error. Upstream notes implementations SHOULD support a single targetRef: multiple entries are accepted by the API, but one is the safest portable shape.

**Attachment is asynchronous controller state.** Whether the policy actually attached is reconciled after apply — read the `Accepted` and `ResolvedRefs` conditions with kubectl rather than expecting them in the outputs.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesService** | `targetRefs[].name` | `status.outputs.service_name` |
| **KubernetesConfigMap** | `validation.caCertificateRefs[].name` | `status.outputs.configmap_name` |

### What This Kind Provides

`status.outputs` carries only `policy_name` and `namespace`, which echo the manifest back — there is nothing downstream to consume from this policy. Whether it actually attached is controller state (the `Accepted` and `ResolvedRefs` conditions), reconciled asynchronously and read with kubectl.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Internal CA** — the backend serves a certificate issued by your own CA; the policy verifies it against a PEM bundle in a same-namespace ConfigMap. Start from the **Internal CA via ConfigMap** preset.

**Publicly-issued backend certificate** — the backend serves a certificate from a public CA; trust the implementation's system store instead of shipping a bundle. Start from the **Public CA via System Trust Store** preset.

**SPIFFE-identified backend** — the backend's certificate identity is a SPIFFE URI rather than a DNS name; the hostname selects the certificate and a URI SAN authenticates it. Start from the **SPIFFE mTLS Backend** preset.

## Works With

- [**Kubernetes Service**](/infra-catalog/kubernetes-service) — the policy's targets, referenced by name and optionally narrowed to one port.
- [**Kubernetes ConfigMap**](/infra-catalog/kubernetes-config-map) — carries the CA bundle when you bring your own trust anchor.
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) — cert-manager issues the backend certificates this policy verifies.
- [**Kubernetes Gateway API CRDs**](/infra-catalog/kubernetes-gateway-api-crds) — installs the CRD this policy is served by.
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) — the gateway whose controller enforces the policy.
