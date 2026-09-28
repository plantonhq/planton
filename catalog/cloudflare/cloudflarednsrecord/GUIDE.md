# CloudflareDnsRecord guide

Operational judgment for configuring a record well. The README covers what each field is; this covers how the pieces interact, including behaviors verified against the live Cloudflare API.

## Names: declare relative, read back fully-qualified

Declare `name` relative to the zone (`www`, `_sip._tcp`, `@` for the apex). Cloudflare stores and answers reads with the full FQDN (`www.example.com`), so the dashboard and API will always show the long form. The `record_name` output deliberately echoes the DECLARED name, not the API's echo, so it stays stable across refreshes.

## Priority lives in exactly one place

MX records take `priority` at the spec top level. SRV and URI records carry priority inside their typed block (`srv.priority`, `uri.priority`) -- never set the top-level field for them. Cloudflare itself mirrors the typed priority into the record's top-level API field, and the IaC modules send that mirror so the deployed record never drifts (measured live: an SRV record deployed without the mirror re-plans forever).

## Records work on pending zones

A record can be created in a zone that is still `pending` (not yet delegated at the registrar) -- verified live. That makes pre-staging practical: build a zone's full record set before cutting nameservers over. Nothing resolves publicly until delegation completes, and proxied records serve no traffic on a pending zone.

## Simple content vs. typed blocks

A record's value comes from exactly one place, and it must match `type`: `content` for A, AAAA, CNAME, MX, NS, PTR, TXT, OPENPGPKEY; a typed block named after the record type for the 13 structured types (SRV, CAA, CERT, DNSKEY, DS, HTTPS, LOC, NAPTR, SMIMEA, SSHFP, SVCB, TLSA, URI). The schema groups the typed blocks in a oneof named `data`, but a manifest never writes a `data:` key -- the active case sits at the spec top level (e.g. `srv: {priority, weight, port, target}`).

## Content that comes from another resource

`content` is a literal (`value:`) or a reference (`valueFrom:`) to another resource's output, resolved before the record is written and ordering the record after that resource. Two compositions use it:

- **Delegating a subdomain.** Publish a child zone's name servers as NS records, one record per name server: `status.outputs.nameservers.0`, `.1`, and so on, from an AwsRoute53Zone, GcpDnsZone or CloudflareDnsZone. The delegation then follows the zone through a recreate, where copied values would silently keep naming the old name servers. A list output is always addressed one element at a time; a reference to the whole list is refused.
- **Proving control of a domain.** Publish a validation target as a CNAME: an Auth0CustomDomain's `status.outputs.dns_record_value`, or a GcpCertManagerDnsAuthorization's `status.outputs.dns_record_data`. Keep such records DNS-only (`proxied: false`): a verifier that resolves the name must see the target itself, not Cloudflare's proxy.

## TTL and the proxy

`ttl: 1` (or leaving it unset) means "automatic", which is what proxied records should use. Explicit TTLs (30-86400s) matter only for gray-cloud records. `proxied` applies to A/AAAA/CNAME only; the module silently ignores it elsewhere.

## Importing an existing record

A record's API identity is compound: `{zone_id}/{record_id}`. The zone id is on the zone's dashboard Overview page (or `GET /zones?name=<domain>`); the record id is in the record's dashboard URL (or `GET /zones/<zone_id>/dns_records?name=<fqdn>`). The import round-trip is proven for both simple and structured records.
