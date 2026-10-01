# CloudflareDnsZone guide

Operational judgment for configuring the zone well. The README covers what each field is; this covers how the pieces interact.

## Pending vs. active zones

A newly created zone is `pending` until your registrar delegates the domain to the nameservers in `status.outputs.nameservers`. Most things work on a pending zone -- records (all types, live-verified), rulesets, most zone-scoped resources -- but nothing resolves publicly and proxied records serve no traffic. One live-measured exception: **DNSSEC cannot be enabled on a pending zone** -- Cloudflare rejects it with 400 code 1017 "Invalid zone plan for action" until the registrar delegates (the same call succeeds on an active free-plan zone, so the wall is delegation state, not plan tier). Check `status.outputs.status` before treating the zone as live, and expect 1-24h of registrar propagation after updating nameservers.

## Inline records vs. standalone CloudflareDnsRecord

Both surfaces carry identical depth (all 21 types, typed structured data, tags, settings), so the choice is purely lifecycle:

- **Inline** when the records live and die with the zone: bootstrap A/AAAA records, MX + SPF/DKIM TXTs, CAA issuance policy, verification TXTs. One manifest, one apply.
- **Standalone** when a record changes on its own cadence (a CI-managed deploy record, per-service SRVs owned by different teams) or is created by another chart wiring `zone_id` via ValueFromRef.

Renaming an inline record's `name` or `type` replaces the record (the resource key includes both) -- brief NXDOMAIN between destroy and create. For records where that matters, standalone resources give finer control.

## Simple content vs. typed data blocks

A record's value comes from exactly one place, and it must match `type`:

- `content` for A, AAAA, CNAME, MX, NS, PTR, TXT, OPENPGPKEY -- the presentation string, as a literal (`value:`) or read from another resource's output (`valueFrom:`).
- A typed block named after the record type for SRV, CAA, CERT, DNSKEY, DS, HTTPS, LOC, NAPTR, SMIMEA, SSHFP, SVCB, TLSA, URI (e.g. `srv: {priority, weight, port, target}`).

Declare `priority` in exactly one place: top-level for MX, inside the typed block for SRV, URI, HTTPS, and SVCB. For SRV and URI the modules mirror the typed-block priority into the provider's top-level field themselves — Cloudflare reflects it there on read, so the mirror is what keeps re-plans clean (live-measured).

## TTL and the proxy

`ttl: 1` (or leaving it unset) means "automatic", which is what proxied records should use -- Cloudflare answers proxied queries with its own edge IPs and manages the TTL itself. Explicit TTLs (30-86400s) matter only for gray-cloud records. `settings.flatten_cname` is likewise a gray-cloud lever: proxied CNAMEs are always flattened.

## DNSSEC is a two-step handshake

`dnssec.enabled: true` makes Cloudflare sign the zone, but the chain of trust completes only when you enter the DS material (`dnssec_ds` / `dnssec_digest` / `dnssec_key_tag` outputs) at your registrar. Do it in that order; entering DS records for an unsigned zone breaks resolution. And do it only AFTER the zone is active: enabling DNSSEC on a still-pending zone fails outright (see "Pending vs. active zones"). `multi_signer` and `presigned` are for multi-provider and Cloudflare-as-secondary setups -- leave both off for the normal single-provider case.

## Entitlement walls you meet before plan tiers

Several small levers fail with a 400 on accounts/zones that lack the matching feature, even when the field looks innocuous (all live-measured on a free-plan account):

- Record `tags`: tag quota is 0 without the feature -- the whole record create fails with code 9300.
- Record `settings.ipv4_only`/`ipv6_only`: code 9227 "not available to this zone".
- `dns_settings` non-default values: SOA tuning, `ns_ttl`, and `flatten_all_cnames` each answer code 1003 "not available to this account or zone".

On top of that, the Terraform provider (v5.23.0) echoes server defaults into any `dns_settings` field you left unset, and then plans to remove them forever -- so declare EVERY field the server echoes (ns_ttl, the SOA block, nameservers) or omit `dns_settings` entirely. Partial `dns_settings` blocks are a permanent-diff trap even on entitled accounts.

## Importing an existing zone

The zone imports by its bare zone id, and every zone-singleton satellite (DNSSEC, hold, subscription) shares that identity. Inline records import as `{zone_id}/{record_id}`; a prior deploy's `record_ids` output carries each record's id keyed by its name-type-index key, so recipes derive them without dashboard archaeology. `dns_settings` has no importer at provider v5.23.0 -- after adopting a zone, one apply re-asserts the settings from configuration (an idempotent PUT).

## The hold is a takeover guard, not a lock

`hold.enabled: true` stops the zone's hostname from being ADDED as a zone in any other Cloudflare account -- protection while a domain migrates between accounts or when subdomains delegate to other teams (`include_subdomains: true` extends the guard to every subdomain, including SSL-for-SaaS custom hostnames). It does not freeze the zone's own configuration. A future-dated `hold_after` temporarily releases the hold until that instant -- a planned migration window that re-arms itself.

## Subscription is a billing action

`subscription.rate_plan` is real money on anything above `free`, effective at apply, and the deploying token needs Billing Write scope (a plain Zone-Edit token gets a 403 on just this satellite). Plan choice also gates other levers in this spec: `vanity_name_servers` needs Business+, `foundation_dns` is a paid add-on, and `multi_provider`/`secondary_overrides` are plan-gated. If your organization manages plans centrally, leave `subscription` unset and let the account default stand.

## Zone types pair with specific settings

- `full` (default): Cloudflare hosts DNS entirely. The normal case.
- `partial`: CNAME setup for partner-hosted zones; most zone-wide DNS settings don't apply.
- `secondary`: Cloudflare mirrors an external primary -- pair with `dns_settings.secondary_overrides` to allow proxied overrides, and note the zone-transfer configuration itself is Enterprise territory.
- `internal`: internal-only resolution -- pair with `dns_settings.internal_dns.reference_zone_id` (a ValueFromRef to another CloudflareDnsZone) for fallback resolution.
