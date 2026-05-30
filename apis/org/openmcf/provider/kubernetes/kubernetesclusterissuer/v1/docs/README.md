# KubernetesClusterIssuer: Research and Design

## Introduction

A ClusterIssuer is a cert-manager custom resource that represents a Certificate Authority (CA) capable of issuing TLS certificates across all namespaces in a Kubernetes cluster. Unlike namespace-scoped Issuers, ClusterIssuers operate at the cluster level, making them the standard choice for platform teams managing TLS for multiple applications.

This component decouples ClusterIssuer lifecycle management from the cert-manager controller installation (KubernetesCertManager), enabling independent management of certificate authority configuration and the cert-manager controller itself.

## Why Decouple ClusterIssuers from cert-manager?

Previously, OpenMCF bundled ClusterIssuer creation directly into the KubernetesCertManager component. While convenient for initial setup, this created several operational problems:

1. **Lifecycle coupling** -- Adding a new DNS domain required redeploying the entire cert-manager controller. A configuration-only change triggered an infrastructure-level operation.

2. **Blast radius** -- A misconfigured DNS provider in a ClusterIssuer could cause the entire cert-manager stack to roll back, affecting all existing issuers and in-flight certificate renewals.

3. **Multi-team friction** -- When different teams own different DNS domains, they all needed to coordinate changes to a single KubernetesCertManager resource.

4. **Single Responsibility Principle** -- cert-manager installation (Helm chart, CRDs, controller SA, DNS resolver config) and ACME issuer configuration (email, server, DNS solver) are fundamentally different concerns.

The decoupled model:
- **KubernetesCertManager** -- Installs the controller, CRDs, and optionally configures workload identity for cloud DNS authentication
- **KubernetesClusterIssuer** -- Creates one ClusterIssuer per DNS domain with its ACME and DNS solver configuration

## ACME DNS-01 Challenge Explained

ACME (Automatic Certificate Management Environment) is the protocol used by Let's Encrypt and other CAs to automate certificate issuance. The DNS-01 challenge type proves domain ownership by creating a specific TXT record in the domain's DNS zone.

**Flow:**
1. cert-manager requests a certificate from the ACME server
2. ACME server provides a challenge token
3. cert-manager creates a `_acme-challenge.{domain}` TXT record with the token
4. ACME server verifies the TXT record exists
5. Certificate is issued and stored as a Kubernetes Secret

**Why DNS-01 over HTTP-01:**
- Supports wildcard certificates (`*.example.com`)
- Works for clusters behind firewalls or private networks
- Does not require an ingress controller to be running
- Works before any HTTP infrastructure is in place

## DNS Provider Authentication Patterns

### Cloudflare (API Token Secret)

Cloudflare uses API tokens for DNS record management. The token is stored as a Kubernetes Secret in the cert-manager namespace, and the ClusterIssuer references this secret.

**Security model:** Per-issuer credentials. Each KubernetesClusterIssuer with Cloudflare creates its own Secret. Tokens should be scoped to specific zones (`Zone:Zone:Read` + `Zone:DNS:Edit`).

### GCP Cloud DNS (Workload Identity)

GCP uses Workload Identity to bind the cert-manager Kubernetes ServiceAccount to a GCP Service Account. The ClusterIssuer only needs the GCP project ID -- authentication happens transparently through the SA binding.

**Security model:** Controller-level identity. The cert-manager SA gets the `iam.gke.io/gcp-service-account` annotation via KubernetesCertManager's `workload_identity.gke` config. The GCP SA needs the `dns.admin` role.

### AWS Route53 (IRSA)

AWS uses IAM Roles for Service Accounts (IRSA) to let the cert-manager pod assume an IAM role. The ClusterIssuer specifies the AWS region -- IRSA handles authentication.

**Security model:** Controller-level identity. The cert-manager SA gets the `eks.amazonaws.com/role-arn` annotation via KubernetesCertManager's `workload_identity.eks` config. The IAM role needs Route53 record modification permissions.

### Azure DNS (Managed Identity)

Azure uses Workload Identity (formerly AAD Pod Identity) to bind the cert-manager SA to a Managed Identity. The ClusterIssuer specifies subscription and resource group.

**Security model:** Controller-level identity. The cert-manager SA gets the `azure.workload.identity/client-id` annotation via KubernetesCertManager's `workload_identity.aks` config. The Managed Identity needs `DNS Zone Contributor` role.

## Design Decisions

### One ClusterIssuer per Component Instance

Each KubernetesClusterIssuer creates exactly one ClusterIssuer. This is deliberate:

- **Independent lifecycle** -- Each domain's issuer can be added, updated, or removed independently
- **Clear ownership** -- One manifest owns one issuer, making GitOps workflows straightforward
- **Blast radius isolation** -- A misconfigured issuer for one domain doesn't affect others

### ClusterIssuer Name = DNS Domain

The ClusterIssuer Kubernetes resource is named after the `dns_domain` field (e.g., `example.com`). This preserves the convention used by all 15+ OpenMCF ingress components, which derive the issuer name from the ingress hostname:

```
hostname: argocd.example.com → issuer: example.com
```

This convention is implemented via `extractDomainFromHostname()` (or equivalent `strings.Split/Join` logic) in each component's Pulumi locals or Terraform locals.

### No Namespace Fields

Unlike most OpenMCF Kubernetes components, KubernetesClusterIssuer does not have `namespace` or `create_namespace` fields. This is because ClusterIssuers are cluster-scoped Kubernetes resources -- they don't live in a namespace. The `cert_manager_namespace` field exists solely to know where to create Secrets (Cloudflare API tokens, ACME private keys).

### Workload Identity Belongs to cert-manager

Cloud provider authentication (GKE Workload Identity, EKS IRSA, AKS Managed Identity) is configured on the cert-manager controller's ServiceAccount, not per-issuer. This avoids cross-resource SA patching and keeps the security boundary at the controller level. KubernetesCertManager owns this configuration via its `workload_identity` field.

## Production Best Practices

### Use Let's Encrypt Production for Real Certificates

The default ACME server is Let's Encrypt production. Use the staging server (`https://acme-staging-v02.api.letsencrypt.org/directory`) only for testing to avoid rate limits during development.

### Scope Cloudflare Tokens

Create Cloudflare API tokens with minimal permissions: `Zone:Zone:Read` and `Zone:DNS:Edit`, scoped to the specific zone. Avoid using global API keys.

### Monitor Certificate Expiry

cert-manager handles automatic renewal (default: 30 days before expiry), but monitor the `cert-manager_certificate_expiration_timestamp_seconds` Prometheus metric as a safety net.

### Plan for Rate Limits

Let's Encrypt enforces rate limits (50 certificates per registered domain per week for production). For large-scale deployments, consider using a single wildcard certificate per domain.

## Conclusion

KubernetesClusterIssuer provides a clean, single-responsibility abstraction for managing cert-manager ClusterIssuers. By decoupling from the cert-manager installation, it enables independent lifecycle management, better multi-team workflows, and reduced blast radius -- while preserving the domain-named issuer convention that all OpenMCF ingress components rely on.

## Composing in Infra Charts

`KubernetesClusterIssuer` sits between cert-manager and the certificates it issues
(see project decision DD-009):

1. **Data dependencies use `valueFrom`.** `cert_manager_namespace` is a
   `StringValueOrRef` (`default_kind = KubernetesCertManager`,
   `default_kind_field_path = status.outputs.namespace`), so referencing the
   cert-manager installation builds the DAG edge automatically -- the issuer
   deploys after cert-manager is ready.
2. **Downstream wiring.** A `KubernetesCertificate` references this issuer through
   its `issuerRef` `StringValueOrRef`
   (`status.outputs.cluster_issuer_name`), so the certificate deploys after the
   issuer.

```yaml
spec:
  cert_manager_namespace:
    valueFrom:
      kind: KubernetesCertManager
      name: "{{ values.env }}-cert-manager"
      fieldPath: status.outputs.namespace
```

Full ingress stack:
`CertManager -> ClusterIssuer -> Certificate -> (Secret) -> Gateway -> HTTPRoute / GRPCRoute`.
