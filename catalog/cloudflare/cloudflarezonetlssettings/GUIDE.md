# CloudflareZoneTlsSettings guide

Operational judgment for a zone's edge TLS posture. The reference page covers what each field is; this covers the one field that can take a site down, which surfaces actually go away on destroy, and the provider quirks the module works around.

## Universal SSL is the one switch that can cause an outage

`universalSslEnabled: false` stops Universal SSL certificate issuance for the zone. Any proxied hostname not covered by another certificate -- dedicated, custom, or Total TLS -- becomes unreachable over HTTPS. Browsers show certificate errors; there is no partial degradation. That is why the field is an explicit optional bool rather than a default: the module only touches Universal SSL when you deliberately set the field, and leaving it unset manages nothing. Disable it only as a step in a planned migration, and confirm every proxied hostname has other coverage first. Remember it has no delete at Cloudflare -- if you disable it and then destroy the resource, the zone stays disabled until something re-enables it.

## Know your destroy class before you destroy

The six surfaces split 4-and-2. `universalSslEnabled`, `totalTls`, `autoOriginTlsKex`, and `caHostnameAssociations` have no delete at Cloudflare: destroy drops them from state and the zone keeps the last-applied values. `hostnameSettings` and `originTlsComplianceModes` have real deletes: destroy removes the overrides and clears the compliance requirement, and hostnames fall back to zone-wide settings. The practical rule: for the four no-delete surfaces, write the value you want to leave behind before destroying, because destroy itself changes nothing at Cloudflare. For the two real-delete surfaces, destroy is a real revert -- do not destroy a resource carrying a compliance requirement you still need.

## One hostname row fans out into per-setting API objects

A `hostnameSettings` row is a convenience grouping. At Cloudflare, each set attribute is its own API object keyed by (setting, hostname): a row setting `minTlsVersion` and `http2` for `api.example.com` becomes two objects, not one. The module keys its resources the same way, with the hostname as the for_each key, so editing one row never churns another row's resources -- adding a hostname or changing one row's override is a surgical change, not a re-index of the list. The flip side: the hostname is part of each object's API identity, so renaming a hostname replaces its override objects rather than updating them.

## Provider import defect on hostname TLS settings (v5.23.0)

The pinned provider's `cloudflare_hostname_tls_setting` resource does not restore the hostname on import -- an imported override comes back without the attribute that identifies it, so every post-import plan destroys and recreates the override. The import catalog therefore marks this resource NOT IMPORTABLE, and the right way to adopt existing per-hostname overrides is to skip import entirely: declare the overrides in your manifest and apply. The per-(setting, hostname) write is an idempotent upsert, so an apply that matches the live values adopts them without churn, and one that differs converges them. For everything else in this kind, fresh applies are unaffected; this only bites import workflows.

## Total TLS validity is not yours to set

Total TLS certificates live for 90 days, fixed by Cloudflare. The provider exposes the validity period as a computed-only attribute, so the spec deliberately does not model it -- there is nothing to decide. Choose the `certificateAuthority` if compliance cares.

## Two surfaces need Advanced Certificate Manager; one needs an active zone

Measured live (2026-08-27): enabling Total TLS AND writing any per-hostname override (`hostnameSettings`) both fail with 401 code 1450 -- "This feature is available with the Advanced Certificate Manager" -- unless the zone carries the ACM subscription (a per-zone add-on billed monthly, independent of the zone's plan tier; a Pro zone without ACM fails the same way a Free zone does). If an apply hits 1450, the fix is to buy ACM for that zone or drop the field, not to retry.

Separately, `autoOriginTlsKex` exists only on ACTIVE zones: writing it on a pending (undelegated) zone fails with 400 code 1000 and the misleading message "Invalid zone identifier". The zone id is fine -- activate the zone first.

## Compliance modes are an open vocabulary on purpose

`originTlsComplianceModes` accepts any string. Cloudflare documents `fips` and `pqh` today and may add values; the spec deliberately does not wall the list with validation, so a new Cloudflare mode works the day it ships without a schema change. The cost is that a typo also passes through -- Cloudflare's API, not the manifest validation, is what rejects `fisp`. Check the apply output, not just the manifest lint.

## The provider is pinned

The IaC module pins the Cloudflare Terraform provider at v5.23.0. The behaviors recorded here -- the destroy classes, the per-(setting, hostname) object model, the import defect, the computed-only validity -- are verified against that version. A provider bump is the moment to re-check them, especially the import defect.

## Pairs well with

- **CloudflareDnsZone** -- reference the zone via `zoneId` with `valueFrom` so the zone deploys first and the TLS settings follow in the graph.
- **CloudflareCertificatePack** -- the advanced certificates that make disabling Universal SSL safe; order coverage before flipping the switch.
- **CloudflareZoneSettings** -- zone-wide TLS knobs (minimum TLS version for the whole zone, Always Use HTTPS) live there; this kind is the per-hostname and issuance layer.
- **CloudflareCustomHostname** -- TLS for SaaS vanity hostnames outside this zone; per-hostname overrides here apply only to hostnames in this zone.
