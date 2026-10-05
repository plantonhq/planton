# GCP Certificate Map

Creates a Certificate Manager certificate map — the hostname-to-certificate routing table an external HTTPS load balancer consults at TLS handshake time: each entry binds a hostname (or the PRIMARY fallback) to up to fifteen certificates, and the map attaches to a GcpTargetHttpsProxy via its `certificate_map` argument. Maps are how a proxy serves MANY certificates — beyond the ~15-certificate direct-attach limit, with per-hostname (SNI) selection at scale.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Certificate map** -- a `certificatemanager.CertificateMapResource` (global — no location by API design)
- **Map entries** -- one `certificatemanager.CertificateMapEntry` per spec entry (hostname or PRIMARY matcher, 1–15 certificates each)
- **Certificate Manager API enablement** -- `certificatemanager.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** to host the map (directly or via a GcpProject reference).
- **Certificates**: GcpCertManagerCert resources (or existing certificate resource names) to bind — entries require 1–15 each.
- **IAM**: the deploying identity needs `roles/certificatemanager.editor` or broader.

## Deploy

### Console

Open the deployment store, find **GCP Certificate Map**, and click **Deploy**. Start from the **Hostname Routing** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertificateMap
metadata:
  name: prod-tls
  org: acme-corp
  env: prod
spec:
  entries:
    - entryName: www
      hostname: www.example.com
      certificates:
        - value: projects/acme-prod-12345/locations/global/certificates/www-cert
    - entryName: fallback
      matcher: PRIMARY
      certificates:
        - value: projects/acme-prod-12345/locations/global/certificates/wildcard-cert
```

```shell
planton apply -f certificate-map.yaml
```

This creates a map that serves `www.example.com` with its dedicated certificate and every other SNI with the wildcard fallback. An Infra Job tracks the provisioning in real time.

### InfraChart

The multi-domain TLS edge in one chart: GcpCertManagerCert resources per domain, this map binding hostnames to them, and a GcpTargetHttpsProxy consuming the `map_uri` output — add a domain by adding a certificate and an entry:

```yaml
spec:
  entries:
    - entryName: www
      hostname: www.example.com
      certificates:
        - valueFrom:
            kind: GcpCertManagerCert
            name: www-cert
            fieldPath: status.outputs.certificate_id
```

The InfraPipeline resolves the dependency graph, provisions the certificates first, then binds them into the map.

## Key Configuration

These are the most important decisions when configuring a certificate map. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**entries** -- the routing table. `hostname` matches the client's SNI (FQDN or wildcard `*.example.com`); `matcher: PRIMARY` is the fallback used when no hostname entry matches — ship one or unmatched handshakes fail.

**certificates** -- 1–15 per entry (the API cap, enforced at manifest time). Each references a GcpCertManagerCert's `certificate_id` output. This is the MUTABLE surface: rotate by editing the list in place — attach the replacement before detaching the old one.

**Entry immutability** -- hostname, matcher, and entry name all REPLACE the entry (a brief window where that hostname has no binding). Plan hostname changes as add-new-then-remove-old.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpCertManagerCert** | `entries[].certificates[]` | `status.outputs.certificate_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `map_id` | Full resource name | gcloud commands, debugging |
| `map_uri` | `//certificatemanager.googleapis.com/...` | A GcpTargetHttpsProxy's `certificate_map` argument |
| `map_name` | Short name | Display |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Hostname routing** -- per-domain certificates selected by SNI, with a PRIMARY fallback. Start from the **Hostname Routing** preset.

**Single wildcard** -- one wildcard certificate serving every subdomain via PRIMARY. Start from the **Wildcard Fallback** preset.

## Works With

- [**GCP Target HTTPS Proxy**](/infra-catalog/gcp-target-https-proxy) -- consumes `map_uri` as its `certificate_map`
- [**GCP Cert Manager Cert**](/infra-catalog/gcp-cert-manager-cert) -- the certificates entries bind
- [**GCP Cert Manager DNS Authorization**](/infra-catalog/gcp-cert-manager-dns-authorization) -- domain-ownership proof for managed certificates
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project
