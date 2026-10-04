# Apex A Record

This preset points a zone's apex (the bare domain, `@`) at an IPv4 address — the standard "make example.com resolve to my server" record. The zone is referenced as a `DigitalOceanDnsZone` resource so the record composes in Infra Charts; replace the reference with a literal (`value: example.com`) for a zone managed outside Planton.

## When to Use

- Pointing a domain at a web server, load balancer, or reserved IP
- The first record almost every new zone needs

## Key Configuration Choices

- **`name: "@"`** — the zone apex. Use a subdomain name (`www`, `api`) to address a host under the zone instead.
- **Zone by reference** (`domain.valueFrom`) — resolves to the zone's `zone_name` output at deploy time, so the record deploys after its zone in one chart.
- **One-hour TTL** (`ttlSeconds: 3600`) — a sensible production default. Lower it to 300 before planned IP changes; DigitalOcean's API default is 1800 when omitted.

## Placeholders to Replace

- `metadata.name` — your record's name.
- `domain.valueFrom.name` — the name of your `DigitalOceanDnsZone` resource (or replace the block with `value: example.com`).
- `value.value` (`203.0.113.10` is a documentation example) — your server's public IPv4 address; it can also reference another resource's output (e.g. a Droplet's `ipv4_address`).

## Related Presets

- **02-cname-record** — alias a subdomain to another hostname instead of an IP.
