# WWW CNAME Record

This preset aliases `www` to the zone's apex hostname — the standard "www.example.com follows example.com" record. The zone is referenced as a `DigitalOceanDnsZone` resource so the record composes in Infra Charts.

## When to Use

- Making `www` (or any subdomain) follow another hostname
- Pointing a subdomain at a platform-provided hostname (an App Platform default hostname, a load balancer's DNS name, a CDN endpoint)

## Key Configuration Choices

- **CNAME semantics** — the target is a hostname, never an IP. A CNAME cannot exist at the zone apex; use an A record (preset `01-a-record`) there.
- **Trailing dot on the target** (`example.com.`) — DigitalOcean stores CNAME/MX/NS/SRV targets fully qualified and reads them back with a trailing dot; authoring the dot avoids a permanent diff.
- **One-hour TTL** (`ttlSeconds: 3600`) — DigitalOcean's API default is 1800 when omitted.

## Placeholders to Replace

- `metadata.name` — your record's name.
- `domain.valueFrom.name` — the name of your `DigitalOceanDnsZone` resource (or replace the block with `value: example.com`).
- `value.value` (`example.com.` is a documentation example) — the target hostname, fully qualified.

## Related Presets

- **01-a-record** — point the zone apex (or any name) at an IPv4 address instead.
