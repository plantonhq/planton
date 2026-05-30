# KubernetesIssuer: Research and Design

## Introduction

An Issuer is a namespace-scoped cert-manager custom resource that represents a Certificate Authority capable of signing TLS certificates within a single namespace. Unlike ClusterIssuers (which operate cluster-wide with ACME DNS-01 challenges), Issuers are designed for namespace-local certificate management using simpler signing modes.

This component focuses on CA and SelfSigned issuer types -- the two modes most commonly used for internal PKI and development workflows. ACME issuance is deliberately excluded (handled by KubernetesClusterIssuer) following the 80/20 rule.

## Why Namespace-Scoped Issuers?

ClusterIssuers handle the common case of externally-trusted TLS certificates (Let's Encrypt via ACME). However, several scenarios require namespace-scoped issuers:

1. **Internal PKI** -- Organizations running their own CA need to sign certificates with a CA keypair, not through ACME. The CA Secret lives in a specific namespace, and the Issuer must be co-located.

2. **CA Chain Bootstrapping** -- A SelfSigned Issuer creates a root CA Certificate, which generates a CA Secret. A CA Issuer then references that Secret to sign leaf certificates. This three-resource chain (SelfSigned Issuer → CA Certificate → CA Issuer) is the standard cert-manager pattern for custom CAs.

3. **Namespace Isolation** -- Teams owning different namespaces can manage their own Issuers independently without cluster-admin permissions.

4. **Development/Testing** -- SelfSigned Issuers provide instant certificate issuance with zero external dependencies.

## CA Chain Bootstrap Pattern

The most common use of KubernetesIssuer is the CA chain bootstrap:

```
SelfSigned Issuer (KubernetesIssuer with self_signed)
       ↓ signs
Root CA Certificate (KubernetesCertificate with is_ca=true)
       ↓ generates Secret with tls.crt + tls.key
CA Issuer (KubernetesIssuer with ca, referencing the Secret)
       ↓ signs
Leaf Certificates (KubernetesCertificate referencing the CA Issuer)
```

This pattern is documented in the [cert-manager CA documentation](https://cert-manager.io/docs/configuration/ca/) and is the recommended approach for internal PKI.

## Design Decisions

### CA and SelfSigned Only

The proto spec only supports `ca` and `self_signed` issuer types. Other cert-manager Issuer types (Vault, Venafi, external) are excluded because:

- **ACME** -- Handled by KubernetesClusterIssuer (cluster-scoped, DNS-01 challenges)
- **Vault/Venafi/External** -- Niche use cases that can be added as separate components when needed

### Issuer Name = metadata.name

Unlike KubernetesClusterIssuer (where the k8s resource name equals `dns_domain`), the Issuer name comes from `metadata.name`. This is because:

- Issuers don't follow the ingress-domain convention (they're not used by ingress components)
- Certificate resources reference Issuers by name within the same namespace
- The foreign-key pattern uses `metadata.name` as the stable identifier

### No Namespace Creation

The Issuer targets an existing namespace. Namespace creation is a separate concern (handled by the namespace's owner or a KubernetesNamespace component). This follows the principle that Issuer lifecycle is independent from namespace lifecycle.

### Foreign Key on ca_secret_name

The `ca_secret_name` field has a foreign-key annotation pointing to `KubernetesCertificate.status.outputs.secret_name`. This enables the UI to wire a CA Issuer to a CA Certificate automatically -- the user picks a KubernetesCertificate (with `is_ca=true`), and the Secret name is resolved from the Certificate's outputs.

## Production Considerations

### CA Secret Must Pre-exist

For CA mode, the referenced Secret must exist before the Issuer is created. In GitOps workflows, ensure the KubernetesCertificate (which generates the CA Secret) is deployed before the KubernetesIssuer that references it.

### Self-Signed Certificates Are Not Trusted

Self-signed certificates are not trusted by browsers or external clients. They are only suitable for:
- Bootstrapping a CA chain (as described above)
- Development and testing environments
- Internal service-to-service mTLS where the CA bundle is explicitly distributed

### Namespace Co-location

The CA Secret, Issuer, and Certificate resources must all reside in the same namespace. cert-manager enforces this constraint -- an Issuer cannot reference a Secret in a different namespace.

## Composing in Infra Charts

`KubernetesIssuer` is a namespaced issuer, typically used for the CA-bootstrap
workflow (see project decision DD-009):

1. **Data dependencies use `valueFrom`.** `namespace` (-> `KubernetesNamespace`)
   and `ca_secret_name`
   (-> `KubernetesCertificate.status.outputs.secret_name`) are `StringValueOrRef`
   fields, so the platform builds those DAG edges automatically -- the CA
   Certificate is provisioned before this Issuer that signs from it.
2. **Downstream wiring.** A `KubernetesCertificate` references this issuer through
   its `issuerRef` `StringValueOrRef` (`status.outputs.issuer_name`).

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: "{{ values.env }}-ns"
      fieldPath: spec.name
  ca_secret_name:
    valueFrom:
      kind: KubernetesCertificate
      name: "{{ values.env }}-ca-cert"
      fieldPath: status.outputs.secret_name
```

Bootstrap chain: `SelfSigned Issuer -> CA Certificate -> CA Issuer (this) -> leaf Certificates`.
