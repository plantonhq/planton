# CloudflareCertificatePack

Order an **advanced edge certificate** for a Cloudflare zone — a publicly-trusted
TLS certificate, provisioned and auto-renewed by Cloudflare, covering the hostnames
you list (beyond the free Universal SSL certificate). Use it when you need a
specific certificate authority, multiple/longer-lived certificates, or coverage for
hostnames Universal SSL does not include.

## When to use

- You need a chosen CA (Google, Let's Encrypt, or SSL.com) at the edge.
- You need a custom set of covered hostnames or a specific validity period.

## Quick start

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCertificatePack
metadata:
  name: edge-cert
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: my-zone
      fieldPath: status.outputs.zone_id
  certificateAuthority: google
  validationMethod: txt
  validityDays: 90
  hosts:
    - example.com
    - "*.example.com"
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `zoneId` | yes | Zone ID (literal or `CloudflareDnsZone` reference) |
| `certificateAuthority` | yes | `google`, `lets_encrypt`, or `ssl_com` |
| `type` | no | `advanced` (default) |
| `validationMethod` | yes | `txt`, `http`, or `email` |
| `validityDays` | yes | 14, 30, 90, or 365 |
| `hosts` | yes | Covered hostnames (must include the zone apex, ≤50) |
| `cloudflareBranding` | no | Add Cloudflare branding to the order |

## Outputs

| Output | Description |
|---|---|
| `certificate_pack_id` | The certificate pack identifier |
| `zone_id` | The zone the pack was ordered in (its API identity is `zone_id` + `certificate_pack_id`) |

There is no `status` output: pack issuance is asynchronous (`initializing` → `pending_validation` → `active`), and a point-in-time phase is never a stable output. There is no `primary_certificate` output for the same reason — the server fills it in asynchronously as issuance progresses (measured live: absent at order, `"0"` seconds later, then the real certificate id). Read both from the Cloudflare API or dashboard.

## Notes

- Most attributes are immutable: changing them re-orders (replaces) the pack.
- For a zone using Cloudflare's nameservers, `txt` validation completes
  automatically — no manual DNS record is required.
- Advanced certificate packs require Advanced Certificate Manager on the zone.

## Related kinds

- `CloudflareDnsZone` — the zone the certificate is ordered for.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
