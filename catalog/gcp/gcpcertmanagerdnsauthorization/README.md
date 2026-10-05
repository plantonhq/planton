# GcpCertManagerDnsAuthorization

## Overview

`GcpCertManagerDnsAuthorization` manages one Certificate Manager DNS
authorization — the proof-of-domain-control a Google-managed certificate
needs before it can be issued for a domain that is not yet serving
traffic. One authorization covers a single domain AND its wildcard:
authorizing `example.com` issues certificates for both `example.com` and
`*.example.com`.

## Purpose

DNS authorization decouples certificate issuance from traffic serving.
The authorization exports a CNAME validation record; once that record
resolves publicly, any certificate referencing the authorization can reach
ACTIVE — before the load balancer exists, before DNS cuts over, with zero
downtime. It is also the only validation mode that supports wildcard
domains.

The full composition is three explicit nodes:

1. **GcpCertManagerDnsAuthorization** — the authorization (this kind)
2. **GcpDnsRecord** — serves the exported validation record in the zone
3. **GcpCertManagerCert** — references the authorization by ID

## Key Features

- Covers the domain and its wildcard with one authorization
- Exports the validation record (`dns_record_name` / `dns_record_type` /
  `dns_record_data`) for direct composition into a `GcpDnsRecord`
- `FIXED_RECORD` and `PER_PROJECT_RECORD` types (per-project records let
  multiple projects share one validation record)
- Regional authorizations for regional certificates
- User labels merged beneath platform attribution labels, identically on
  both engines
- `deletionPolicy` (DELETE / PREVENT / ABANDON) controls what a destroy
  does — PREVENT protects the validation chain a certificate migration
  depends on

## Example Usage

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerDnsAuthorization
metadata:
  name: example-com-auth
spec:
  domain: example.com
```

Serve its validation record from the zone:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDnsRecord
metadata:
  name: example-com-cert-validation
spec:
  managedZone:
    valueFrom:
      kind: GcpDnsZone
      name: example-com
      fieldPath: status.outputs.zone_name
  type: CNAME
  name: "_acme-challenge.example.com."
  values:
    - <dns_record_data from this authorization's outputs>
```

## Best Practices

1. **Create the authorization and its record FIRST**, then the
   certificate — issuance validates immediately instead of polling.
2. **One authorization per distinct domain**; reuse it across every
   certificate that covers that domain or its wildcard.
3. **Keep the authorization alive** across certificate rotations — it has
   its own lifecycle precisely so certificates can come and go.

## Related Kinds

- **GcpCertManagerCert** — references this authorization by ID
- **GcpDnsRecord** / **GcpDnsZone** — serve the validation record

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
