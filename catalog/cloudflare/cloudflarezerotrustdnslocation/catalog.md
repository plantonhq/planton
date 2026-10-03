# Cloudflare Zero Trust DNS Location

Deploys a Cloudflare Gateway DNS location: a named entry point — an office, site, or network — whose DNS traffic Gateway filters. Cloudflare assigns the location its resolver endpoints (a unique DoH subdomain and destination IPs), and Gateway policies match on the location to apply per-site filtering. A plain CRUD object with one sharp edge: updates are full replaces at the API, so omitted fields can actively reset rather than stay unmanaged.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Gateway DNS Location** — a `cloudflare_zero_trust_dns_location` carrying the name, the four-endpoint resolver tree (DoH, DoT, IPv4, IPv6), source-network allowlists, TTL capping, and the account-default flag

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with a Cloudflare API token carrying Account → Zero Trust → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** — the account must have completed Cloudflare Zero Trust onboarding (the team-name step).
- **The site's egress CIDRs** (only for the plain-IPv4 endpoint) — the source networks to allowlist in `networks`; Cloudflare caps IPv4 CIDRs at /24, nothing broader.
- **A dedicated destination-IP mapping** (only for `dnsDestinationIpsId`) — a BYOIP or dedicated-resolver-IP mapping the account actually holds. Leave the field unset otherwise.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust DNS Location**, and click **Deploy**. The creation wizard walks you through the owning account, the location name, the resolver endpoint tree with its source-network allowlists, TTL capping, and the account-default flag. Start from the **Office network location** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustDnsLocation
metadata:
  name: hq-office
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: hq-office
  networks:
    - network: 203.0.113.0/24
```

```shell
planton apply -f dns-location.yaml
```

This creates a named location whose plain-IPv4 endpoint accepts queries from the office's egress CIDR; Cloudflare assigns the DoH subdomain and destination IPs. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a DNS location. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Omitting `maxTtl` on update resets it** — unlike the settings singletons elsewhere in Zero Trust, a location update sends the whole object: an update that omits `maxTtl` resets the TTL behavior to inherit. If you manage a TTL override, keep it declared forever. Validation pairs `ttlSecs` (60–36000) with mode `override` so a manifest cannot get the pairing wrong.

**`clientDefault` is an account-level lever** — setting it true makes this location the attribution target for traffic from unregistered sources, account-wide, one location at a time. Flipping it on a new location silently changes how every unattributed query is filtered. Treat it like changing a default route: deliberately, in a change window, never on a scratch location.

**The endpoint tree is all-or-nothing** — setting `endpoints` declares all four types at once (DoH, DoT, IPv4, IPv6), each enabled or not; Cloudflare's API takes the whole tree in one write. Unset keeps Cloudflare's endpoint defaults. The IPv4 endpoint is the one type whose source networks live at the spec's top-level `networks` field rather than inline.

**Token-gated DoH is the roaming shape** — the DoH subdomain is world-reachable by design; `requireToken` on the doh endpoint rejects resolvers that merely discovered the URL. For office networks the source-network allowlists carry the gating; for roaming devices the token does.

**Never pin the shared destination pool** — `dnsDestinationIpsId` left unset lets Cloudflare auto-assign the shared IPv4 destination pair. Copying the shared pool's UUID into a manifest pins what auto-assign would have done anyway and blocks a future move to dedicated IPs. Set the field only for a dedicated mapping you actually hold.

**Destroy is a real delete** — the location and its endpoints disappear immediately. Anything resolving against the location's destination IP or DoH subdomain loses DNS; repoint resolvers first.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The optional destination-IP mapping (`dnsDestinationIpsId`) is a literal Cloudflare mapping UUID because that surface carries no catalog kind.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `location_id` | The Cloudflare-assigned UUID of the location | Gateway policies matching on the location |
| `doh_subdomain` | The unique DoH subdomain (`https://<sub>.cloudflare-gateway.com/dns-query`) | Resolver URLs on roaming clients and browsers |
| `ip` | The IPv4 destination for the plain-DNS endpoint | Office resolver and DHCP configuration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Office network location** — the site's egress CIDR allowlisted on the plain-IPv4 endpoint, the other endpoint types declared off. Point the office resolver or DHCP at the assigned destination IP (the `ip` output) and match Gateway policies on this location. Start from the **Office network location** preset.

**Roaming DoH location** — DNS-over-HTTPS only, token-gated (source networks cannot gate devices that move), with a short TTL cap so policy changes propagate quickly. Clients embed the `doh_subdomain` output in their resolver URL. Start from the **Roaming DoH location** preset.

**Per-site policy targeting** — one location per office or site, each with its own egress CIDRs, so Gateway policies can filter a branch differently from headquarters and reporting attributes queries to the right site.

## Works With

- [**Cloudflare Zero Trust Gateway Policy**](/infra-catalog/cloudflare-zero-trust-gateway-policy) — the filtering rules that match on this location's queries.
- [**Cloudflare Zero Trust Gateway Settings**](/infra-catalog/cloudflare-zero-trust-gateway-settings) — the account-wide Gateway posture the location's filtering runs under.
