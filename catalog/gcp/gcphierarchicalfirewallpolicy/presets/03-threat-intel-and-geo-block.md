# Threat Intel And Geo Block

The matching surface legacy VPC firewall rules never had, at the top of the
organization: deny traffic from Google's known-malicious and Tor
exit-node lists and from listed countries on ingress, deny traffic to
known-malicious destinations on egress, then delegate.

## What it configures

- Priority `100` — `deny` ingress from `iplist-known-malicious-ips` and
  `iplist-tor-exit-nodes`, Google-curated threat-intelligence lists that
  Google keeps current; logged.
- Priority `200` — `deny` ingress from the ISO 3166-1 country codes listed
  in `srcRegionCodes`; logged.
- Priority `300` — `deny` egress to `iplist-known-malicious-ips`, so a
  compromised VM cannot call home; logged.
- Priorities `2000` and `2100` — `goto_next` for ingress and egress, the
  delegation that leaves every other decision to the levels below.
- `deletionPolicy: PREVENT`.

## Adjust before deploying

- **`organizationId`** (both places) — your numeric organization ID.
- **`srcRegionCodes`** — the countries your organization never serves; the
  two listed are placeholders.
- **Lists** — Google also curates `iplist-vpn-providers`,
  `iplist-anon-proxies`, `iplist-crypto-miners`, and the public-cloud lists.

## When to choose something else

For administrative-port denies, start from **Org Deny SSH From Internet**;
for a folder-scoped allow-list, from **Folder Allow Internal Goto Next**.
