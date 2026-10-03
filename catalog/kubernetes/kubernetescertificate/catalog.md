# Cert Manager Certificate

Requests one signed X.509 certificate from a cert-manager issuer and keeps it renewed for as long as the resource exists. The signed certificate, its private key, and the CA chain land in a Kubernetes TLS Secret that consumers — Ingress TLS blocks, Gateway listeners, workload volume mounts, CA issuers — reference by name. The issuer decides WHO signs; this resource decides WHAT is requested: the names, lifetime, key parameters, usages, and output formats. Covers the complete cert-manager.io/v1 Certificate surface.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Certificate** -- the cert-manager custom resource in the specified namespace, wired to the selected issuer
- **TLS Secret** -- created and kept renewed by cert-manager (keys: `tls.crt`, `tls.key`, `ca.crt`), optionally extended with JKS/PKCS#12 keystores, DER, or combined-PEM entries
- When `isCa: true` -- the Secret carries a CA certificate suitable for signing other certificates (the internal-PKI bootstrap)

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **Cert Manager must be installed first** -- the Certificate is a cert-manager custom resource.
- **At least one issuer available** -- a Cert Manager Cluster Issuer (serves any namespace) or a Cert Manager Issuer in the SAME namespace as this certificate; or a third-party issuer kind (e.g. AWS Private CA) via the external reference.
- Wildcard names (`*.apps.example.com`) require a DNS-01-capable ACME issuer or a private CA.

## Deploy

### Console

Open the deployment store, find **Cert Manager Certificate**, and click **Deploy**. The creation wizard walks you through the namespace, the issuer, the requested names, the output Secret, lifetime and renewal, private-key parameters, and advanced X.509 settings. Start from the **Standard TLS Certificate via ClusterIssuer** preset for standard public TLS in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesCertificate
metadata:
  name: api-example-com
  org: acme-corp
  env: prod
spec:
  namespace:
    value: team-payments
  secretName: api-example-com-tls
  issuerRef:
    clusterIssuer:
      name:
        value: letsencrypt-production
  dnsNames:
    - api.example.com
```

```shell
planton apply -f certificate.yaml
```

This requests a certificate for `api.example.com` from the platform ClusterIssuer and keeps it renewed into the `api-example-com-tls` Secret — everything else (lifetime, key parameters, usages) deliberately left to the issuer's defaults. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the namespace and issuer to resources managed by other Infra Components:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: team-payments-namespace
      fieldPath: spec.name
  secretName: api-example-com-tls
  issuerRef:
    clusterIssuer:
      name:
        valueFrom:
          kind: KubernetesClusterIssuer
          name: letsencrypt-production
          fieldPath: status.outputs.cluster_issuer_name
  dnsNames:
    - api.example.com
```

The InfraPipeline deploys the namespace and the issuer first, then requests the certificate against them.

## Key Configuration

These are the most important decisions when configuring a Certificate. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Three issuer reference arms** -- `issuerRef.clusterIssuer` (cluster-scoped, serves any namespace — the platform default), `issuerRef.issuer` (namespace-scoped, must live in the certificate's namespace), or `issuerRef.external` (group + kind + name of a third-party issuer kind, e.g. `awspca.cert-manager.io` / `AWSPCAClusterIssuer`). Exactly one.

**At least one name** -- DNS names are the common case; IP SANs serve bare-IP endpoints (internal issuers only — public ACME CAs refuse them), URI SANs carry SPIFFE IDs for mTLS workload identity, email SANs serve S/MIME. A common name alone also satisfies the requirement, but modern TLS validation ignores CN in favor of SANs.

**The two renewal dials are exclusive** -- `renewBefore` (an absolute window, e.g. `360h`) and `renewBeforePercentage` (renew when N% of the lifetime remains — scales with issuer-decided lifetimes). Set at most one; the default (a third of the lifetime) is production-ready.

**The issuer may override the lifetime** -- `duration` is a request; ACME CAs set their own policy (Let's Encrypt always issues 90 days). Leaving lifetime fields empty is an honest, recommended posture.

**Key parameters when a consumer dictates them** -- ECDSA gives smaller, faster keys; RSA maximizes legacy compatibility; Ed25519 is rejected by many CAs (including Let's Encrypt). Key size must match the algorithm family (RSA: 2048–8192; ECDSA: 256/384/521). PKCS#8 encoding is required by some Java consumers. Rotation policy `always` (the upstream default) is the safe choice.

**Keystores for JVM/Windows consumers** -- JKS and PKCS#12 entries can be added to the Secret alongside the PEM data, each protected by a password (declared sensitively, materialized platform-side). The PKCS#12 `modern2023` profile uses current crypto but requires consumers newer than ~2013.

**The CA bootstrap** -- `isCa: true` against a self-signed issuer produces a root CA Secret; a CA-backend issuer then signs leaf certificates with it. Pair with `nameConstraints` to restrict what names the delegated CA may sign — defense-in-depth for internal PKI.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesClusterIssuer** | `issuerRef.clusterIssuer.name` | `status.outputs.cluster_issuer_name` |
| **KubernetesIssuer** | `issuerRef.issuer.name` | `status.outputs.issuer_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace where the Certificate was created | Verifying consumer co-location (the Secret is namespace-local) |
| `certificate_name` | Name of the Certificate resource | Debugging issuance (`kubectl describe certificate`) |
| `secret_name` | The TLS Secret name | Gateway `certificateRefs`, Ingress `tls.secretName`, or `ca_secret_name` on a CA-backend Issuer |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public TLS** -- One DNS name from the platform ClusterIssuer into one Secret. Start from the **Standard TLS Certificate via ClusterIssuer** preset.

**Wildcard for multi-service domains** -- `*.apps.example.com` via a DNS-01-capable ClusterIssuer. Start from the **Wildcard TLS Certificate** preset.

**Root CA bootstrap** -- `isCa: true` against a self-signed Issuer, producing the CA Secret for internal PKI. Start from the **Self-Signed Root CA Certificate (CA Bootstrap)** preset.

## Works With

- [**Cert Manager**](/infra-catalog/kubernetes-cert-manager) -- must be installed first; provides the controller and CRDs.
- [**Cert Manager Cluster Issuer**](/infra-catalog/kubernetes-cluster-issuer) -- the cluster-scoped signing authority; [**Cert Manager Issuer**](/infra-catalog/kubernetes-issuer) is the namespace-scoped alternative, and a CA-backend Issuer also consumes this resource's Secret output for the CA bootstrap.
- [**Ingress NGINX**](/infra-catalog/kubernetes-ingress-nginx) -- terminates HTTPS with the output Secret; [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) listeners reference it through `certificateRefs`.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- reference it so infra charts create the namespace and this certificate in dependency order.
