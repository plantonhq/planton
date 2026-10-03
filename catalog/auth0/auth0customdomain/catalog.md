# Auth0 Custom Domain

Serves an Auth0 tenant's sign-in on a domain you own, such as `id.example.com`, in place of the tenant's `auth0.com` address. Universal Login, the links in the tenant's emails, and the issuer of its tokens then carry your name. Creating the domain returns the DNS record that proves you control it; an Auth0 Custom Domain Verification completes it.

## What Gets Created

When you deploy this Infra Component, the IaC module creates:

- **A custom domain** in the tenant your Auth0 connection's credential belongs to, with its certificate type, TLS policy, client-IP header, metadata, and passkey relying party
- **The DNS record to publish** -- returned in the outputs, ready for a DNS record Infra Component to read

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with the tenant's domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A DNS Infra Component for the record** -- for example a Cloudflare DNS Record in the zone that serves the domain, reading this domain's outputs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **The connection's Machine-to-Machine application** must hold `create:custom_domains`, `read:custom_domains`, `update:custom_domains` and `delete:custom_domains` on the tenant's Management API (Auth0 dashboard: Applications, APIs, Auth0 Management API, Machine To Machine Applications).
- **A card on file** on a Free tenant, which Auth0 requires before it allows the one custom domain the plan includes (it is not charged). Self-managed certificates need the Enterprise plan.

## Deploy

### Console

Open the deployment store, find **Auth0 Custom Domain**, and click **Deploy**. Start from the **Auth0-Managed Domain on Cloudflare** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0CustomDomain
metadata:
  name: sign-in-domain
  org: acme-corp
  env: prod
spec:
  domain: id.acme.com
  type: auth0_managed_certs
```

```shell
planton apply -f auth0-custom-domain.yaml
```

The domain is created in `pending_verification`, and its outputs carry the CNAME to publish. An Infra Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Who holds the certificate** -- `auth0_managed_certs` lets Auth0 issue and renew it: publish one CNAME and keep it, DNS-only (no CDN proxying, no CNAME flattening), because every renewal re-checks it. `self_managed_certs` puts your own proxy and certificate in front, needs the Enterprise plan, and is proven with a TXT record.

**The record completes it** -- Compose the domain's `dns_record_name`, `dns_record_type` and `dns_record_value` into a DNS record, then order an Auth0 Custom Domain Verification after that record. All three install together.

**Both domains keep working** -- The tenant's canonical domain still answers. A token's issuer is the domain that served the request, so trust `https://<your domain>/` in applications that sign in through it, and set the tenant's default domain (Auth0 Tenant Settings) so email links use it too.

**Passkeys** -- Set `relyingPartyIdentifier` to a parent domain (`acme.com` for `id.acme.com`) so one passkey signs in across your domains; Auth0 recommends it.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The tenant is the one the Auth0 connection's credential belongs to.

### What This Kind Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The custom domain's identifier (`cd_...`) | An Auth0 Custom Domain Verification's `customDomainId` |
| `domain` | The custom domain's name | Application issuer settings, documentation |
| `status` | `pending_verification`, then `ready` | Audits |
| `origin_domain_name` | The tenant host the domain serves from | A self-managed proxy's origin |
| `dns_record_name` | The record's name | A DNS record's `name` |
| `dns_record_type` | `CNAME` or `TXT` | A DNS record's `type` |
| `dns_record_value` | The CNAME target or TXT text | A DNS record's `content` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Auth0-managed domain on Cloudflare** -- The domain, its CNAME in a Cloudflare zone, and the verification, in one install. Start from the **Auth0-Managed Domain on Cloudflare** preset.

**Passkeys across your domains** -- A custom domain whose passkeys bind to the parent domain. Start from the **Passkeys Across Your Domains** preset.

**Your own proxy** -- A self-managed domain behind your reverse proxy and certificate. Start from the **Self-Managed Certificate** preset.

## Works With

- [**Auth0 Custom Domain Verification**](/infra-catalog/auth0-custom-domain-verification) -- waits until the domain is verified and serving.
- [**Cloudflare DNS Record**](/infra-catalog/cloudflare-dns-record) -- publishes the record that proves control of the domain.
- [**Auth0 Tenant Settings**](/infra-catalog/auth0-tenant-settings) -- makes the domain the tenant's default, so email links use it.
