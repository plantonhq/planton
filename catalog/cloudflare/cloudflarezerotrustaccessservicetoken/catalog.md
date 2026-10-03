# Cloudflare Zero Trust Access Service Token

Deploys an Access service token: a machine credential — a client-ID / client-secret pair that non-human clients present in the `CF-Access-Client-ID` / `CF-Access-Client-Secret` request headers to pass through Access-protected applications without an identity-provider login. The secret is returned only at creation and at each rotation, never on reads; capture it into a secret store at deploy or a lost secret means rotating. Rotation is first-class: the token ID and client ID survive it, so Access policies keep matching.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Access Service Token** — one service token at the account (or zone) with the given display name and duration
- **Client ID and Client Secret** — exported as outputs; the secret is marked sensitive and is never readable again except at rotation

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with an API token holding Account → Access: Service Tokens → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **Zero Trust enabled on the account** — the organization (team name) must already exist (a CloudflareZeroTrustOrganization Infra Component) or every Access create fails at the API.
- **A secret store ready to receive the credential** — `status.outputs.client_secret` cannot be recovered from Cloudflare later; capturing it is part of the deploy, not an afterthought.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Access Service Token**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the account-or-zone scope, the token name, and the duration and rotation fields. Start from the **Minimal** preset in the [Presets](#presets) tab to pre-populate the standard one-year token.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessServiceToken
metadata:
  name: ci-deployer
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: ci-deployer
```

```shell
planton apply -f service-token.yaml
```

This creates an account-scoped token with Cloudflare's one-year default duration — capture `status.outputs.client_id` and `status.outputs.client_secret` into your CI secret store before anything else. An Infra Job tracks the provisioning in real time.

### InfraChart

For a zone-scoped token whose zone is deployed in the same InfraPipeline, wire `zoneId` with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  name: ci-deployer
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then creates the token scoped to the resolved zone.

## Key Configuration

These are the most important decisions when configuring a service token. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The secret is shown twice in the token's life** — at create, and at each rotation; never again. Reads never return it, and an imported token lands with an empty secret. If you did not capture it, the token still exists and policies still match its ID, but no client can authenticate until you rotate. Write `client_id` and `client_secret` to your secret store in the same change that creates the token.

**Rotate in place, never destroy-and-recreate** — incrementing `clientSecretVersion` mints a new secret while `service_token_id` and `client_id` stay the same, so Access policy `service_token` rules keep matching. Destroy-and-recreate mints a new token ID and every policy that listed the old ID stops matching.

**Rotation is a pair** — `clientSecretVersion` and `previousClientSecretExpiresAt` must be set together or both left unset (the normal non-rotating state; Cloudflare treats the initial secret as version 1). The expiry controls how long the OLD secret keeps working: extend it into the future so services can migrate, or set it in the past to kill a compromised secret immediately.

**Duration versus rotation** — `duration` is how long the token itself is valid: a Go-style duration (`8760h` is Cloudflare's one-year default) or `forever` for a non-expiring token. Rotation replaces the secret inside that lifetime — a `forever` token still needs rotation when the secret leaks.

**Account scope or zone scope** — exactly one of `accountId` or `zoneId`. Account-scoped tokens (the common case) work with every application in the account; zone-scoped tokens serve one zone.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** (zone-scoped tokens) | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `service_token_id` | The token's UUID | Access policy `service_token` include rules |
| `client_id` | The value for the `CF-Access-Client-ID` header | Machine-client configuration |
| `client_secret` | The value for the `CF-Access-Client-Secret` header (sensitive; returned only at create and rotation) | Machine-client secret stores |
| `expires_at` | When the token expires (RFC3339) | Expiry monitoring and renewal automation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**CI deployer token** — a named token with the one-year default, captured into the CI secret store at deploy; the everyday machine credential. Start from the **Minimal** preset.

**Scheduled rotation** — increment the version and give the old secret a 30-day expiry so services migrate at their own pace; the token ID never changes. Start from the **Rotation** preset.

**Compromised-secret kill** — the same rotation shape with `previousClientSecretExpiresAt` set in the past, invalidating the leaked secret immediately while the new one takes over.

**Long-lived automation identity** — `duration: forever` for infrastructure that cannot tolerate annual expiry, paired with a standing rotation practice instead of an expiry date.

## Works With

- [**Cloudflare Zero Trust Access Policy**](/infra-catalog/cloudflare-zero-trust-access-policy) — a `service_token` include rule that lists this token's ID
- [**Cloudflare Zero Trust Access Application**](/infra-catalog/cloudflare-zero-trust-access-application) — the application the machine client calls
- [**Cloudflare Zero Trust Access Identity Provider**](/infra-catalog/cloudflare-zero-trust-access-identity-provider) — human sign-in, the other door through Access
- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the `zoneId` foreign key for a zone-scoped token
