# DigitalOcean DNS Record

A single DNS record in a DigitalOcean-hosted zone, described once in a Planton manifest: every record type the DigitalOcean API accepts (A, AAAA, CNAME, MX, TXT, SRV, NS, CAA, SOA), the per-type fields each requires, and a zone reference so records compose with their zone in infra charts.

## What this component models

The spec maps one-to-one onto DigitalOcean's `digitalocean_record`:

| Spec field | What it controls |
|---|---|
| `domain` | The zone the record lives in — a literal domain or a reference to a `DigitalOceanDnsZone`; changing it recreates the record |
| `name` | The host name relative to the zone (`@` for the apex, `www`, `api.v1`) |
| `type` | The record type; changing it recreates the record |
| `value` | The record's target — an IP, hostname, or text, as a literal or a reference to another resource's output |
| `ttlSeconds` | Cache TTL; omit to take DigitalOcean's default (1800) |
| `priority` | Required for MX and SRV (lower wins) |
| `weight` / `port` | Required for SRV |
| `flags` / `tag` | Required for CAA (`flags: 0` non-critical, `128` critical; `tag`: `issue`, `issuewild`, or `iodef`) |

The provider's type-conditional requirements are enforced at validation time — an MX without a priority or an SRV missing its port never reaches a provisioner.

## Quick start

Point a subdomain at a server:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDnsRecord
metadata:
  name: app-a-record
spec:
  domain:
    valueFrom:
      kind: DigitalOceanDnsZone
      name: my-zone
      fieldPath: status.outputs.zone_name
  name: app
  type: A
  value:
    value: 203.0.113.10
```

A service locator with the full SRV surface:

```yaml
spec:
  domain:
    value: example.com
  name: _sip._tcp
  type: SRV
  value:
    value: sip.example.com.
  priority: 10
  weight: 60
  port: 5060
```

## Behavior worth knowing

- **Hostname values read back with a trailing dot** — CNAME/MX/NS/SRV/CAA targets are stored fully qualified (`mail.example.com.`, `letsencrypt.org.`); author the dot, or a zone-relative name (`mail`). A bare `letsencrypt.org` is re-applied on every run.
- **Explicit zeros are dropped** — the provider omits a `priority`/`weight`/`port`/`flags` of exactly 0 from the create request and the API's default applies; use positive values when exactness matters (CAA `flags: 0` is safe — the API default IS 0).
- **Concurrent writes to one zone can deadlock** — DigitalOcean fails one of several simultaneous record writes to the same domain with a `422 ... Deadlock found` error; sequence many standalone records, or use the zone kind's inline `records`, which are written one at a time.
- **TTLs harmonize server-side** — DigitalOcean forces one TTL across records sharing a fully-qualified name (RFC 2181), so the live TTL can drift when a sibling record changes it.
- **Records import with a two-part ID** — `{domain},{record_id}`; both are stack outputs of this component.

## Outputs

| Output | Meaning |
|---|---|
| `record_id` | The record's numeric ID (string form) — with `domain`, how the API and imports address it |
| `hostname` | The fully-qualified name (the provider's computed fqdn) |
| `record_type` | The type that was created |
| `domain` | The zone the record was created in |
| `ttl_seconds` | The applied TTL (the API default when the spec left it unset) |

## See also

- `GUIDE.md` — operational judgment calls (record-vs-zone ownership, TTL strategy, SRV/CAA anatomy)
- `presets/` — apex A record and www CNAME starting points
- `v1alpha1/reference.md` — the generated field-by-field contract

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
