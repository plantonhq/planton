# GCP Cert Manager Cert

Creates one Certificate Manager certificate — the modern certificate resource external Application Load Balancers consume through a target HTTPS proxy's certificate list or a certificate map. Exactly one of two arms: **Google-managed** (GCP provisions and RENEWS the certificate automatically — expiry stops being an incident class) or **self-managed** (you upload a PEM chain and key; rotation is an in-place spec update).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Certificate Manager API enablement** (`certificatemanager.googleapis.com`) on the target project (never disabled on destroy)
- **Certificate Manager Certificate** -- a `google_certificate_manager_certificate` in the chosen project and location, either MANAGED (with the listed domains and validation channel) or SELF_MANAGED (with the uploaded PEM material)

Domain validation resources are NOT bundled: DNS authorizations are first-class [GcpCertManagerDnsAuthorization](/infra-catalog/gcp-cert-manager-dns-authorization) resources you compose, which is what makes issuing a certificate BEFORE traffic serves (zero-downtime migration) possible.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Org Secret** (self-managed certificates) -- store the private key as an org secret and reference it as `$secret/<slug>`; the runner resolves it just-in-time and the platform rejects plaintext.

### GCP Project

- **For DNS-authorization validation** -- one [GcpCertManagerDnsAuthorization](/infra-catalog/gcp-cert-manager-dns-authorization) per distinct domain, its exported CNAME served by a [GcpDnsRecord](/infra-catalog/gcp-dns-record) in the domain's zone. Required for wildcard domains and for issuing before the load balancer serves.
- **For load-balancer validation** -- nothing up front: GCP validates through the serving load balancer once traffic reaches it (the certificate stays PROVISIONING until then; wildcards are not supported).

## Deploy

### Console

Open the deployment store, find **GCP Cert Manager Cert**, and click **Deploy**. The wizard leads with the managed-vs-self-managed fork, then the certificate's envelope (project, name, location, serving scope), then the arm's own step — domains & validation for managed, the PEM pair for uploads. Start from the **Managed Certificate with DNS Authorization** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerCert
metadata:
  name: orders-tls
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  managed:
    domains:
      - orders.example.com
    dnsAuthorizations:
      - valueFrom:
          kind: GcpCertManagerDnsAuthorization
          name: orders-example-auth
          fieldPath: status.outputs.authorization_id
```

```shell
planton apply -f cert.yaml
```

This creates a Google-managed certificate for `orders.example.com` validated through the referenced DNS authorization — it can reach ACTIVE before any load balancer serves the domain. An Infra Job tracks the provisioning in real time.

### InfraChart

The zero-downtime TLS shape in one chart: a DNS authorization per domain, a GcpDnsRecord serving each validation CNAME, and this certificate referencing the authorizations:

```yaml
spec:
  managed:
    domains:
      - orders.example.com
    dnsAuthorizations:
      - valueFrom:
          kind: GcpCertManagerDnsAuthorization
          name: orders-example-auth
          fieldPath: status.outputs.authorization_id
```

The InfraPipeline provisions the authorizations and their DNS records first, then requests the certificate — issuance completes without touching live traffic.

## Key Configuration

These are the most important decisions when configuring a Certificate Manager certificate. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Certificate type** -- Exactly one of `managed` or `selfManaged`. Managed certificates renew forever; self-managed certificates serve exactly what you uploaded until you rotate them.

**Domains** (managed) -- Immutable, at least one; bare (`example.com`) or wildcard (`*.example.com`) names. A wildcard covers ONE level and not the apex — list the apex separately. Wildcards require DNS authorizations or an issuance config.

**Validation channel** (managed) -- `dnsAuthorizations` and `issuanceConfig` are mutually exclusive; omitting BOTH selects load-balancer authorization. DNS authorizations issue before traffic serves and are the only channel that validates wildcards; an issuance config signs from your own CA (private PKI) instead of a public one.

**PEM material** (self-managed) -- `pemCertificate` is the public chain, leaf first then intermediates. `pemPrivateKey` is secret material — always a `$secret/<slug>` reference. Both are MUTABLE: updating the pair in place rotates the served certificate with no downtime.

**Location and scope** -- Immutable. Empty location means `global` — what classic external HTTPS load balancers consume; a region serves regional load balancers only. `scope` decides where Google serves from: `DEFAULT` (core data centers), `EDGE_CACHE` (Media CDN), `ALL_REGIONS` (cross-region internal ALBs — global certificates only), `CLIENT_AUTH` (presented to the backend for mTLS).

**Certificate name** -- Immutable, unique per location, 1-64 chars starting with a letter. Empty defaults to `metadata.name`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpCertManagerDnsAuthorization** | `managed.dnsAuthorizations[]` | `status.outputs.authorization_id` |
| **GcpCertManagerIssuanceConfig** | `managed.issuanceConfig` | `status.outputs.issuance_config_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `certificate_id` | Fully-qualified resource ID (`projects/*/locations/*/certificates/*`) | Certificate maps, auditing |
| `certificate_name` | The certificate's GCP name | A GcpTargetHttpsProxy's `certificate_manager_certificates` list |
| `san_dnsnames` | The SANs in the issued certificate | Verifying coverage |
| `location` | The Certificate Manager location | Cross-checking proxy placement |
| `managed_state` | `PROVISIONING` / `ACTIVE` / `FAILED` (managed only) | Gating DNS cutover on ACTIVE |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Managed with DNS authorization** -- the zero-downtime migration shape: prove domain control through DNS, watch `managed_state` reach ACTIVE, then cut traffic over — the certificate is serving-ready before the load balancer ever sees a request. Start from the **Managed Certificate with DNS Authorization** preset.

**Wildcard certificate** -- one certificate for `*.example.com` plus the apex listed separately (a wildcard never covers the apex or nested levels). DNS authorization is mandatory here — load-balancer validation cannot prove control of a wildcard. Start from the **Wildcard Certificate** preset.

**Self-managed upload** -- a PEM chain and `$secret/`-referenced private key for certificates issued outside GCP (corporate CAs, EV certificates). Renewal is on you: update the pair in place before expiry and the rotation serves with no downtime. Start from the **Self-Managed (Uploaded) Certificate** preset.

**Load-balancer validation** -- no DNS authorization and no issuance config: GCP validates through the serving load balancer once traffic reaches it. The simplest channel for a domain already pointing at the LB — but the certificate sits in PROVISIONING until then, so it cannot front a zero-downtime migration and never validates wildcards.

## Works With

- [**GCP Cert Manager DNS Authorization**](/infra-catalog/gcp-cert-manager-dns-authorization) -- proves domain control before issuance; required for public wildcards
- [**GCP Cert Manager Issuance Config**](/infra-catalog/gcp-cert-manager-issuance-config) -- has your private CA pool issue and renew the certificate; its `issuance_config_id` output feeds `managed.issuanceConfig`
- [**GCP DNS Record**](/infra-catalog/gcp-dns-record) -- serves each authorization's validation CNAME in the zone
- [**GCP DNS Zone**](/infra-catalog/gcp-dns-zone) -- the zone those records live in
