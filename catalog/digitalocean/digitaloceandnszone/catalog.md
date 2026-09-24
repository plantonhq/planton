# DigitalOcean DNS Zone

Deploys a DNS zone (domain) on DigitalOcean with optional inline DNS records covering every type the DigitalOcean API accepts -- A, AAAA, CNAME, MX, TXT, SRV, NS, CAA, and SOA -- plus the create-only apex-A convenience. DigitalOcean manages the authoritative nameservers for the zone, and record values can reference outputs from other Cloud Resources via ValueFromRef. The zone answers on DigitalOcean's nameservers the moment it exists, but the public internet only follows once the registrar's NS delegation is flipped -- and domain names are unique across ALL DigitalOcean accounts, so a name another account holds cannot be created.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **DigitalOcean Domain** -- a DNS zone served by DigitalOcean's nameservers (`ns1.digitalocean.com`, `ns2.digitalocean.com`, `ns3.digitalocean.com`); optionally seeded with an untracked apex A record via `ipAddress`
- **DNS Records** -- created only when `records` are provided; one DigitalOcean DNS record per value entry (multi-value records fan out), with type-specific fields for MX priority, SRV priority/weight/port, and CAA flags/tag enforced at validation time

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **A domain you control** -- adding a zone does not require owning the domain (DigitalOcean hosts it immediately), but public resolution starts only after your registrar delegates to DigitalOcean's nameservers. Note domain names are unique across ALL DigitalOcean accounts.
- **IP addresses or hostnames for records** -- A records take IPv4 addresses, AAAA take IPv6, CNAME/MX/NS/SRV/CAA take target hostnames (author them fully qualified WITH a trailing dot, `mail.example.com.` / `letsencrypt.org.`, or relative to the zone -- a bare `letsencrypt.org` is re-applied on every run because DigitalOcean reports it back with the dot).

## Deploy

### Console

Open the deployment store, find **DigitalOcean DNS Zone**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Simple Website Zone** preset in the [Presets](#presets) tab to create a zone with apex and www records.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDnsZone
metadata:
  name: example-com
  org: acme-corp
  env: prod
spec:
  domainName: example.com
  records:
    - name: "@"
      type: A
      values:
        - value: "203.0.113.10"
      ttlSeconds: 3600
```

```shell
planton apply -f do-dns-zone.yaml
```

This creates a DNS zone for `example.com` with a single A record pointing the apex at the specified IP address. A Stack Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a DNS zone. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Domain name** -- The `domainName` field must be a valid fully-qualified domain name (e.g., `example.com`). After provisioning, update the domain's nameservers at your registrar to DigitalOcean's set (the `name_servers` output). DNS propagation can take up to 48 hours.

**Record types** -- Each record in the `records` list specifies a `type` (DigitalOcean accepts A, AAAA, CNAME, MX, TXT, SRV, NS, CAA, SOA; ALIAS and PTR are rejected at validation time), a `name` (use `@` for the apex), one or more `values` (each value becomes its own record -- two A values make round-robin), and an optional `ttlSeconds`. MX records require `priority`, SRV records require `priority`/`weight`/`port`, and CAA records require `flags`/`tag` -- all enforced before any provisioner runs.

**TTL** -- The `ttlSeconds` field controls how long resolvers cache the record; omit it to take DigitalOcean's default (1800 seconds). Use shorter TTLs (300) during migrations, longer (3600-86400) for stable records. Records that share a name share one TTL on DigitalOcean: give them the same value or leave them all unset, or the lone custom TTL is rewritten server-side on every run.

**ValueFromRef in record values** -- Record `values` accept ValueFromRef references, so records can point at outputs of other Cloud Resources (a Droplet's `ipv4_address`, a load balancer's IP) instead of hardcoded values.

**`ipAddress`** -- a create-only convenience that seeds an apex A record the platform never tracks afterwards. Prefer declaring the apex record in `records`; use `ipAddress` only when migrating a configuration that already relies on it. Applied at creation only: later edits are ignored rather than recreating the zone, so adopting an existing zone whose manifest carries it is safe.

**Destroy blast radius** -- destroying the zone deletes every record in it, including records created by the standalone DNS record kind and records added by hand in the control panel. Deleting a zone also releases the domain name for any DigitalOcean account to claim. Enumerate what lives in a shared zone before destroying it.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies (record values may optionally reference any resource's outputs).

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `zone_name` | Domain name of the DNS zone | DNS record `domain` references, App Platform custom domains |
| `zone_id` | The zone's resource identifier -- the domain name itself, not a UUID | API operations, imports |
| `name_servers` | DigitalOcean's fixed authoritative nameserver set | Domain registrar NS delegation |
| `urn` | The domain's uniform resource name (`do:domain:example.com`) | DigitalOcean project assignment, audit |
| `record_ids` | Numeric ids of the inline records, keyed `<record name>-<record index>-<value index>` | API operations on a single record, state import (`{domain},{record_id}`) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Simple website zone** -- apex A record plus a www CNAME. Minimal configuration for getting a domain live quickly. Start from the **Simple Website Zone** preset.

**Production zone with email** -- website records plus MX pairs with priorities, an SPF policy, and CAA certificate authority pinning. Start from the **Production Zone with Email** preset.

## Works With

- [**DigitalOcean DNS Record**](/cloud-catalog/digital-ocean-dns-record) -- standalone records referencing this zone's `zone_name` output, for records owned by other teams or charts
- [**DigitalOcean App Platform App**](/cloud-catalog/digital-ocean-app) -- app custom domains reference the zone so App Platform manages their DNS records in it
