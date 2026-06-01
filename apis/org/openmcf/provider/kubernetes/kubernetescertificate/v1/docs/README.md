# KubernetesCertificate: Research and Design

## Introduction

A cert-manager Certificate is a custom resource that declaratively requests a signed X.509 certificate from an Issuer or ClusterIssuer. The controller watches for Certificate resources, communicates with the configured issuer, and stores the resulting signed certificate and private key in a Kubernetes Secret.

This component wraps the Certificate CR lifecycle into a single OpenMCF resource, supporting both public TLS (via ACME ClusterIssuers) and internal PKI (via SelfSigned/CA Issuers).

## Why a Standalone Certificate Component?

Previously, OpenMCF components that needed TLS certificates (KubernetesLocust, KubernetesArgocd, KubernetesDeployment, etc.) each created their own Certificate inline. This worked well for ingress-coupled certificates but left a gap for two important patterns:

1. **CA Bootstrap** -- Creating a self-signed root CA certificate requires a dedicated Certificate with `isCA: true`, decoupled from any ingress or application component

2. **Shared Certificates** -- Some certificates (e.g., internal mTLS root CAs) are consumed by multiple components but owned by none of them

KubernetesCertificate fills this gap by providing a first-class resource for certificates that exist independently of any particular application component.

## Certificate Lifecycle

### ACME (Public TLS)

```
Certificate CR created
   → cert-manager creates a CertificateRequest
   → CertificateRequest references the ClusterIssuer
   → ClusterIssuer's ACME solver creates DNS-01 challenge
   → ACME server verifies domain ownership
   → Signed certificate stored in Secret
   → cert-manager renews automatically before expiry
```

### Self-Signed (CA Bootstrap)

```
Certificate CR created (isCA: true)
   → cert-manager creates a CertificateRequest
   → CertificateRequest references a SelfSigned Issuer
   → cert-manager generates a self-signed certificate
   → CA certificate + private key stored in Secret
   → The Secret becomes input for a CA Issuer
```

## CA Bootstrap Workflow In Detail

The CA bootstrap workflow is the primary reason KubernetesCertificate supports `is_ca` and namespace-scoped Issuer references. This workflow creates a complete internal PKI in four steps:

### Step 1: Create a SelfSigned Issuer

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesIssuer
metadata:
  name: selfsigned-bootstrap
spec:
  namespace:
    value: cert-manager
  selfSigned: {}
```

This creates a namespace-scoped Issuer that can sign certificates using a generated private key. It is the only issuer type that needs no external CA -- it "bootstraps from nothing."

### Step 2: Create a Root CA Certificate

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertificate
metadata:
  name: my-root-ca
spec:
  namespace:
    value: cert-manager
  dnsNames:
    - my-root-ca
  secretName: my-root-ca-secret
  isCa: true
  issuerRef:
    issuer:
      name:
        value: selfsigned-bootstrap
  durationConfig:
    duration: "87600h"
    renewBefore: "2160h"
  privateKey:
    algorithm: rsa
    size: 4096
    encoding: pkcs1
    rotationPolicy: never
```

Key decisions:
- **`isCa: true`** tells cert-manager to add the CA basic constraint and cert-sign key usage
- **Long duration** (87600h = 10 years) because root CAs should be long-lived
- **`rotationPolicy: never`** because rotating a root CA key invalidates all issued certificates
- **`dnsNames: ["my-root-ca"]`** is a conventional SAN; the actual identity is in the CA certificate's subject

### Step 3: Create a CA Issuer

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesIssuer
metadata:
  name: my-internal-ca
spec:
  namespace:
    value: cert-manager
  ca:
    secretName:
      value: my-root-ca-secret
```

This creates an Issuer backed by the root CA certificate from Step 2. It reads the CA certificate and private key from the Secret and uses them to sign new certificates.

### Step 4: Issue Leaf Certificates

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertificate
metadata:
  name: my-service-cert
spec:
  namespace:
    value: my-app
  dnsNames:
    - my-service.my-app.svc.cluster.local
  secretName: my-service-tls
  issuerRef:
    issuer:
      name:
        value: my-internal-ca
```

## Issuer Reference Design

The proto spec uses a `oneof` for the issuer reference:

```protobuf
oneof issuer_type {
  ClusterIssuerRef cluster_issuer = 1;
  IssuerRef issuer = 2;
}
```

This maps directly to cert-manager's `issuerRef.kind`:
- `cluster_issuer` → `kind: ClusterIssuer` (cluster-scoped, ACME)
- `issuer` → `kind: Issuer` (namespace-scoped, SelfSigned/CA)

Both branches use a `StringValueOrRef` name field, enabling foreign key references to existing KubernetesClusterIssuer or KubernetesIssuer resources.

## Private Key Configuration

### Proto Enum to cert-manager CRD Value Mapping

The proto enums use lowercase names (following protobuf naming conventions), while the cert-manager CRD expects specific casing:

| Proto Enum Value | cert-manager CRD Value | Notes |
|---|---|---|
| `rsa` | `RSA` | Most common; default if unspecified |
| `ecdsa` | `ECDSA` | Smaller keys, faster TLS handshakes |
| `ed25519` | `Ed25519` | Newest; not supported by all clients |
| `pkcs1` | `PKCS1` | Traditional RSA key encoding |
| `pkcs8` | `PKCS8` | Universal encoding, works with all algorithms |
| `always` | `Always` | Regenerate key on every renewal |
| `never` | `Never` | Keep same key across renewals (required for CA certs) |

The Pulumi module contains explicit mapping functions (`mapAlgorithm`, `mapEncoding`, `mapRotationPolicy`) that bridge this naming gap. In Terraform, the middleware populates defaults using the cert-manager CRD format directly.

### Why PrivateKey is a Pointer

In the Pulumi locals, `PrivateKey` is `*certmanagerv1.CertificateSpecPrivateKeyArgs` (a pointer). When nil, no `privateKey` block is emitted in the Certificate spec, which lets cert-manager apply its own CRD-level defaults (RSA 2048, PKCS1, Always). This is the desired behavior for most certificates -- users only specify private key configuration when they need non-default values (e.g., ECDSA for performance, or `Never` rotation for CA certs).

## Production Best Practices

### Use Appropriate Key Algorithms

- **RSA 2048** -- Default; broadest compatibility with clients and load balancers
- **RSA 4096** -- Higher security for CA certificates; avoid for leaf certs (slower TLS)
- **ECDSA P-256** -- Faster TLS handshakes, smaller certificates; good for high-traffic services
- **Ed25519** -- Most modern but limited client support; verify compatibility first

### CA Certificate Rotation

CA certificates should use `rotationPolicy: never`. Rotating a CA's private key means all previously issued certificates can no longer be verified against the CA's public key, requiring re-issuance of every leaf certificate.

### Duration Guidelines

| Certificate Type | Recommended Duration | Recommended RenewBefore |
|---|---|---|
| Public TLS (ACME) | 2160h (90 days) | 360h (15 days) |
| Internal leaf cert | 8760h (1 year) | 720h (30 days) |
| Root CA | 87600h (10 years) | 2160h (90 days) |
| Intermediate CA | 43800h (5 years) | 2160h (90 days) |

## Conclusion

KubernetesCertificate provides a clean, single-resource abstraction for cert-manager Certificate lifecycle management. Its primary value is enabling the CA bootstrap workflow (SelfSigned → CA cert → CA Issuer → leaf certs) while also serving as a standalone resource for certificates that exist independently of ingress or application components.

## Composing in Infra Charts

`KubernetesCertificate` is the bridge between the cert-manager family and the
Gateway family: it issues a TLS Secret that a `KubernetesGateway` listener
terminates with (see project decision DD-009):

1. **Data dependencies use `valueFrom`.** `namespace` (-> `KubernetesNamespace`)
   and `issuerRef` (-> `KubernetesClusterIssuer.status.outputs.cluster_issuer_name`
   or `KubernetesIssuer.status.outputs.issuer_name`) are `StringValueOrRef` fields,
   so the platform builds those DAG edges automatically -- the issuer deploys
   before this certificate.
2. **Downstream wiring (cross-family).** This component's
   `status.outputs.secret_name` is the Secret a `KubernetesGateway` references in
   `listeners[].tls.certificate_refs`. Because `certificate_refs` is a plain
   reference (not a foreign key), the Gateway expresses that dependency via
   `metadata.relationships` (`type: uses`) pointing at this Certificate.

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: "{{ values.env }}-ns"
      fieldPath: spec.name
  issuerRef:
    name:
      valueFrom:
        kind: KubernetesClusterIssuer
        name: "{{ values.env }}-issuer"
        fieldPath: status.outputs.cluster_issuer_name
```

Full ingress stack:
`CertManager -> ClusterIssuer -> Certificate (this) -> (Secret) -> Gateway -> HTTPRoute / GRPCRoute`.
