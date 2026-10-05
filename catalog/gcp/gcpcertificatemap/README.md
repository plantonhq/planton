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

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertificateMap
metadata:
  name: prod-tls
spec:
  entries:
    - entryName: www
      hostname: www.example.com
      certificates:
        - value: projects/my-project/locations/global/certificates/www-cert
    - entryName: fallback
      matcher: PRIMARY
      certificates:
        - value: projects/my-project/locations/global/certificates/wildcard-cert
```

```shell
planton apply -f certificate-map.yaml
```

Attach the `map_uri` output to a GcpTargetHttpsProxy's `certificate_map` argument.

## Outputs

| Output | Description |
|--------|-------------|
| `map_id` | Full resource name |
| `map_uri` | The `//certificatemanager.googleapis.com/...` form a GcpTargetHttpsProxy consumes |
| `map_name` | The short map name |

## Works With

- **GcpTargetHttpsProxy** -- consumes `map_uri` as its `certificate_map`
- **GcpCertManagerCert** -- the certificates entries bind (via `certificate_id`)
- **GcpCertManagerDnsAuthorization** -- proves domain ownership for managed certificates
- **GcpProject** -- provides the GCP project the map lives in

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
