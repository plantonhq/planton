# Cloudflare Zero Trust Access Identity Provider

Connects Cloudflare Access to an identity source — Google, Okta, Azure AD, GitHub, a generic OIDC or SAML provider, or Cloudflare's own one-time PIN — so users can sign in to Access-protected applications. The provider type is immutable at Cloudflare: changing it replaces the provider and invalidates every policy rule that referenced the old ID. Enabling SCIM provisioning mints a bearer secret that Cloudflare shows exactly once.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Access Identity Provider** — one identity provider at the account (or zone) with the chosen `type` and its per-type connection config
- **SCIM Endpoint and Secret** — created only when `scimConfig.enabled` is true; the bearer secret is returned once in the `scim_secret` output and redacted on every later read

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with an API token holding Account → Access: Organizations, Identity Providers, and Groups → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **Zero Trust enabled on the account** — the organization (team name) must already exist (a CloudflareZeroTrustOrganization Infra Component) or every Access create fails at the API.
- **IdP application credentials** (only for OAuth/OIDC/SAML types) — a client ID and secret from the external provider, or SAML metadata. `onetimepin` needs none.
- **A SAML certificate set** (only for `config.enableEncryption`) — created out-of-band via the Access SAML-certificate API; `samlCertificateSetId` names it.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Access Identity Provider**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the account-or-zone scope, the provider type, and the type's connection parameters. Start from the **One-time PIN**, **GitHub OAuth**, or **Okta with SCIM** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessIdentityProvider
metadata:
  name: otp-fallback
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: one-time-pin
  type: onetimepin
```

```shell
planton apply -f identity-provider.yaml
```

This creates Cloudflare's own one-time PIN provider — users see a PIN option on the Access login page, and no IdP application is required. One caveat: Cloudflare allows exactly one `onetimepin` provider per account — if yours already has one, a second create is refused with 409 "a onetimepin connection already exists"; adopt the existing provider by import instead (see the GUIDE). An Infra Job tracks the provisioning in real time.

### InfraChart

For a zone-scoped provider whose zone is deployed in the same InfraPipeline, wire `zoneId` with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  name: github
  type: github
  config:
    clientId: your-github-oauth-app-id
    clientSecret:
      value: your-github-oauth-app-secret
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then creates the provider scoped to the resolved zone.

## Key Configuration

These are the most important decisions when configuring an Access identity provider. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Type is a one-way door** — changing `type` is not an update but a replacement with a new provider ID, and every Access policy rule that referenced the old ID (azure_ad, github_organization, gsuite, okta, saml, oidc, login_method, auth_context) now points at a deleted object. To migrate: create a second provider, attach policies to the new ID, then destroy the old one. The spellings are Cloudflare's own — `azureAD` and `google-apps` are exact; lowercasing them fails validation here instead of at the API.

**Account scope or zone scope** — exactly one of `accountId` or `zoneId` must be set. Account-scoped providers (the common case) are usable by every application in the account; zone-scoped providers serve a single zone.

**Config is a union** — which `config` fields apply depends on `type`, and the spec enforces the pairing at manifest validation: `directoryId` only on `azureAD`, `authUrl`/`tokenUrl`/`certsUrl` only on `oidc`, the SAML fields only on `saml`, `oktaAccount` only on `okta`. A GitHub provider with `oktaAccount` set is rejected before it reaches the API. `onetimepin` takes no config at all — omit the object; the module sends Cloudflare the empty config it expects.

**The SCIM secret is shown once** — enabling `scimConfig` mints a bearer token that Cloudflare returns only on that create and redacts on every later read; it does not survive import. Capture `status.outputs.scim_secret` into a secret store in the same change that turns SCIM on. If lost, refresh it via the Access API's `refresh_scim_secret` endpoint — do not destroy and recreate the provider just to see it again (that changes the provider ID). `seatDeprovision` requires `userDeprovision`, and SCIM is forbidden on `onetimepin`.

**Read-only is a latch** — `readOnly: true` tells Cloudflare to refuse API updates and deletes on this provider; clearing it requires an explicit apply. Use it on the corporate IdP you cannot afford to lose to a bad pull request.

**Creation validates shape, not sign-in** — an OAuth provider creates successfully with credentials users cannot actually log in with. Verify the login flow at the Access login page after deploying, not from the apply result.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** (zone-scoped providers) | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `identity_provider_id` | The provider's UUID | Access policy rules and Access application `allowedIdps` lists |
| `scim_base_url` | Cloudflare's SCIM v2.0 endpoint (present when SCIM is enabled) | Configuring provisioning at the external identity provider |
| `scim_secret` | The SCIM bearer token, minted once (sensitive) | The credential paired with `scim_base_url` at the identity provider |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One-time PIN fallback** — Cloudflare's own PIN-by-email provider, no external IdP; the standard fallback for contractors and break-glass access. Start from the **One-time PIN** preset.

**Social OAuth sign-in** — a GitHub (or Google) OAuth App connected to Access; the fast path for developer-facing tools. Start from the **GitHub OAuth** preset.

**Corporate IdP with SCIM** — Okta (or Azure AD) with `scimConfig` enabled so deprovisioned users lose their sessions and seats automatically instead of at next re-authentication. Start from the **Okta with SCIM** preset.

## Works With

- [**Cloudflare Zero Trust Organization**](/infra-catalog/cloudflare-zero-trust-organization) — the team-name prerequisite every Access resource needs
- [**Cloudflare Zero Trust Access Application**](/infra-catalog/cloudflare-zero-trust-access-application) — the door users reach after signing in through this provider
- [**Cloudflare Zero Trust Access Policy**](/infra-catalog/cloudflare-zero-trust-access-policy) — the guards whose rules match this provider's identities
- [**Cloudflare Zero Trust Access Service Token**](/infra-catalog/cloudflare-zero-trust-access-service-token) — machine credentials, the non-human counterpart to this kind
- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the `zoneId` foreign key for a zone-scoped provider
