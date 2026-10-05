# GCP Managed SSL Certificate

Deploys a Google-managed classic Compute Engine SSL certificate — Google issues it, renews it, and rotates the key material automatically; you own only the domain list. The hands-off TLS choice for global external Application Load Balancers. FQDN-only (wildcards need Certificate Manager or an imported certificate), global-only, and DNS-gated: the certificate stays PROVISIONING until every listed domain's public DNS points at the load balancer. Every field is immutable — a domain change is a create-before-destroy replacement.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Google-managed Compute Engine SSL Certificate** -- global scope, covering every FQDN in `domains` (1-100 entries)
- **Compute Engine API enablement** -- `compute.googleapis.com` enabled in the target project (never disabled on destroy)
- **Automatic issuance and renewal** -- Google validates each domain by observing DNS point at the load balancer, then issues and renews with no rotation calendar

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the certificate will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The module enables the Compute Engine API itself.
- **Control of each domain's DNS** -- issuance completes only once every listed FQDN resolves to the load balancer's IP.

## Deploy

### Console

Open the deployment store, find **GCP Managed SSL Certificate**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Single-Domain Load Balancer Certificate** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedSslCertificate
metadata:
  name: prod-lb-tls-2026
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  certificateName: "prod-lb-tls-2026"
  domains:
    - app.example.com
```

```shell
planton apply -f managed-ssl-certificate.yaml
```

This creates the certificate in PROVISIONING state. Attach it to a target HTTPS proxy, point `app.example.com` at the load balancer's IP, and Google issues within minutes to hours. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the certificate to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph, deploys the project first, then provisions the certificate — and downstream target proxies reference its `self_link` output.

## Key Configuration

These are the most important decisions when configuring a managed SSL certificate. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Domains** -- Every FQDN the certificate must cover; apex and www are separate entries. No wildcards — `*.example.com` needs a GcpCertManagerCert with a DNS authorization, or an imported GcpSslCertificate. Immutable: changing the list replaces the certificate and restarts DNS-gated provisioning, so keep the old certificate attached until the new one is ACTIVE.

**Versioned name** -- `certificateName` defaults to the resource name; bake a version into it (`prod-lb-tls-2026`) so the domain-change replacement can run create-before-destroy. Google-managed and self-managed certificates share one name namespace per scope.

**The provisioning order** -- Reserve the IP (GcpGlobalAddress) → create this certificate → attach both to the proxy → point DNS at the IP → issuance completes. Creating the certificate before DNS cutover is expected; it simply waits.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | Self-link URI of the certificate | GcpTargetHttpsProxy `sslCertificates` list via ValueFromRef |
| `certificate_name` | Name as it exists in GCP | Audit, replacement planning |
| `certificate_id` | Server-assigned numeric ID | GCP console links, API references |
| `expire_time` | RFC3339 expiry — EMPTY until Google issues | Provisioning-state check; Google renews before it automatically |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single domain** -- One FQDN behind a global external load balancer — the standard serving certificate. Start from the **Single-Domain Load Balancer Certificate** preset.

**Apex + www** -- Both spellings of the site on one certificate. Start from the **Multi-Domain Certificate (Apex + WWW)** preset.

**Versioned rotation** -- An explicit `certificateName` decoupled from the resource name demonstrates the domain-change replacement workflow. Start from the **Explicit GCP Certificate Name** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the certificate is created
- [**GCP Target HTTPS Proxy**](/infra-catalog/gcp-target-https-proxy) -- consumes the certificate's `self_link` in its certificate list
- [**GCP SSL Certificate**](/infra-catalog/gcp-ssl-certificate) -- the bring-your-own alternative for wildcards, private CAs, and internal load balancers
- [**GCP DNS Record**](/infra-catalog/gcp-dns-record) -- points each domain at the load balancer to unlock issuance
