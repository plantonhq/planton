# GCP Cert Manager DNS Authorization

Creates one Certificate Manager DNS authorization — the proof of domain control a Google-managed certificate needs before it can be issued for a domain that is not yet serving traffic. One authorization covers a single domain AND its wildcard: authorizing `example.com` issues certificates for both `example.com` and `*.example.com`. It is the only validation mode that supports wildcards, and it decouples certificate issuance from traffic serving — the key to zero-downtime TLS migration.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Certificate Manager API enablement** (`certificatemanager.googleapis.com`) on the target project (never disabled on destroy)
- **Certificate Manager DNS Authorization** -- a `google_certificate_manager_dns_authorization` for the domain, exporting the DNS validation record (a CNAME) that must exist in the domain's zone

The validation record itself is NOT created here — serve it with a [GcpDnsRecord](/infra-catalog/gcp-dns-record) wired to this kind's outputs, and validation completes automatically.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.

### GCP Project

- **An authoritative DNS zone** for the domain — typically a [GcpDnsZone](/infra-catalog/gcp-dns-zone) — where the exported validation CNAME will be served.

## Deploy

### Console

Open the deployment store, find **GCP Cert Manager DNS Authorization**, and click **Deploy**. The wizard walks two decisions: the authorization's envelope (project, name, location), then the domain being authorized and its record scope. Start from the **Standard Domain Authorization** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerDnsAuthorization
metadata:
  name: orders-example-auth
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  domain: example.com
```

```shell
planton apply -f authorization.yaml
```

This creates a global authorization for `example.com` (covering its wildcard too) and exports the validation CNAME to serve in the domain's zone. An Infra Job tracks the provisioning in real time.

### InfraChart

The validation chain in one chart — this authorization, the GcpDnsRecord serving its CNAME, and the certificate that consumes it:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  domain: example.com
```

The InfraPipeline provisions the project reference first; downstream, a GcpDnsRecord wires its name/type/values to this kind's `dns_record_name` / `dns_record_type` / `dns_record_data` outputs so validation completes without a manual DNS step.

## Key Configuration

These are the most important decisions when configuring a DNS authorization. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Domain** -- Immutable, bare domain only (no `*.` prefix, no trailing dot). The authorization covers the domain AND its wildcard — never create a separate authorization for `*.example.com`. Subdomain wildcards need their own authorization (`*.example.com` does not cover `a.b.example.com`).

**Validation record scope** (`type`) -- Immutable. `FIXED_RECORD` is the classic DNS-01 style record, one per (domain, authorization); `PER_PROJECT_RECORD` scopes the record to (domain, project) so multiple Certificate Manager resources across projects share one record. Empty lets GCP pick the location-appropriate default.

**Location** -- Empty means `global`. Regional authorizations pair with regional certificates only — keep it aligned with the certificate that will consume this authorization.

**Authorization name** -- 1-64 chars starting with a letter; empty defaults to `metadata.name`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `authorization_id` | Fully-qualified resource ID (`projects/*/locations/*/dnsAuthorizations/*`) | A GcpCertManagerCert's `dnsAuthorizations` list — the family's composition key |
| `authorization_name` | The authorization's GCP name | Auditing |
| `domain` | The authorized domain | Cross-checking coverage |
| `dns_record_name` | The validation record's fully-qualified name | A GcpDnsRecord's record name |
| `dns_record_type` | The validation record's type (CNAME) | A GcpDnsRecord's record type |
| `dns_record_data` | The value the CNAME must point at | A GcpDnsRecord's record values |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard domain authorization** -- one global authorization per domain, its CNAME served by a GcpDnsRecord in the domain's zone. Create it once and reuse it for every certificate covering that domain or its wildcard — validation stays warm across renewals and new certificates. Start from the **Standard Domain Authorization** preset.

**Shared per-project record** -- `type: PER_PROJECT_RECORD` scopes the validation record to (domain, project), so multiple Certificate Manager resources across projects share one record instead of each demanding its own. The shape for organizations issuing certificates for the same domain from several projects. Start from the **Shared Per-Project Authorization** preset.

**Pre-migration validation** -- create the authorization and its DNS record while the domain still serves from the old infrastructure; the certificate issues and reaches ACTIVE before any traffic moves — the zero-downtime TLS migration this kind exists for.

## Works With

- [**GCP Cert Manager Cert**](/infra-catalog/gcp-cert-manager-cert) -- the Google-managed certificate that references this authorization by ID
- [**GCP DNS Record**](/infra-catalog/gcp-dns-record) -- serves the exported validation CNAME; wire its name/type/values to this kind's outputs
- [**GCP DNS Zone**](/infra-catalog/gcp-dns-zone) -- the authoritative zone the record lives in
