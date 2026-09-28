# Auth0 Custom Domain Verification

Completes an Auth0 custom domain. It asks Auth0 to check the domain's DNS record, then waits until the domain is ready to serve sign-in: verified, and for an Auth0-managed domain, holding the certificate Auth0 issued. Order it after the DNS record that proves control of the domain.

## What Gets Created

When you deploy this Cloud Resource, the IaC module:

- **Verifies the custom domain** -- Auth0 checks its DNS record and, for an Auth0-managed domain, issues its certificate
- **Waits until the domain is ready** -- and fails naming the last status Auth0 reported if it is not ready in time
- **Reads the verified domain back** -- its name and origin host, and the proxy key of a self-managed domain

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with the tenant's domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **An Auth0 Custom Domain** -- the domain to verify, referenced by `customDomainId`.
- **The DNS record that proves it** -- for example a Cloudflare DNS Record reading the domain's `dns_record_*` outputs, declared as a `depends_on` relationship of this Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **The connection's Machine-to-Machine application** must hold `create:custom_domains` (Auth0 files verification under create) and `read:custom_domains` on the tenant's Management API.

## Deploy

### Console

Open the deployment store, find **Auth0 Custom Domain Verification**, and click **Deploy**. Start from the **Verify After the Record** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0CustomDomainVerification
metadata:
  name: sign-in-domain-verification
  org: acme-corp
  env: prod
  relationships:
    - kind: CloudflareDnsRecord
      name: sign-in-domain-cname
      type: depends_on
spec:
  customDomainId:
    valueFrom:
      kind: Auth0CustomDomain
      name: sign-in-domain
      fieldPath: status.outputs.id
```

```shell
planton apply -f auth0-custom-domain-verification.yaml
```

When the Stack Job succeeds, the domain is ready and people can sign in on it.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Order it after the record** -- Auth0 can verify only a record that already resolves publicly, so declare the DNS record as a `depends_on` relationship. Keep the record DNS-only: a CDN proxy or CNAME flattening hides it from Auth0.

**A one-time action** -- There is nothing to update. Destroying this Cloud Resource leaves the domain verified; destroy the Auth0 Custom Domain to remove the domain.

**The proxy key** -- For a self-managed domain, `cname_api_key` is the key your proxy sends to Auth0 in the `cname-api-key` header. Auth0 returns it once; Planton keeps it as a secret output.

## Outputs and Dependencies

### What This Component Consumes

| Field | Foreign Key | Required |
|-------|-------------|----------|
| `customDomainId` | Auth0 Custom Domain (`status.outputs.id`) | Yes |

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `custom_domain_id` | The verified custom domain's identifier | Audits |
| `domain` | The verified domain's name | Auth0 Tenant Settings' `defaultCustomDomain` |
| `origin_domain_name` | The tenant host the domain serves from | A self-managed proxy's origin |
| `cname_api_key` | The self-managed proxy's key (secret) | A self-managed proxy's configuration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Verify after the record** -- The verification of an Auth0-managed domain, ordered after its Cloudflare CNAME. Start from the **Verify After the Record** preset.

**A self-managed domain** -- The verification that hands your proxy its key. Start from the **Self-Managed Domain** preset.

## Works With

- [**Auth0 Custom Domain**](/cloud-catalog/auth0-custom-domain) -- the domain this verifies.
- [**Cloudflare DNS Record**](/cloud-catalog/cloudflare-dns-record) -- publishes the record Auth0 checks.
- [**Auth0 Tenant Settings**](/cloud-catalog/auth0-tenant-settings) -- makes the verified domain the tenant's default.
