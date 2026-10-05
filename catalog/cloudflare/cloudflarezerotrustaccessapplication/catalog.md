# Cloudflare Zero Trust Access Application

Deploys a Cloudflare Zero Trust Access application: the protected resource -- a self-hosted web app, a SaaS app, an SSH/VNC/RDP target, an app launcher, or an MCP endpoint -- that Cloudflare Access guards. The application binds one or more standalone Access policies (referenced by ID) to the resource and configures how users reach and authenticate to it: session lifetime, identity providers, cookie posture, CORS, and -- for SaaS types -- the SAML or OIDC federation Cloudflare provides. An application is account-scoped or zone-scoped (exactly one of `accountId` or `zoneId`); account scope is the common case and lets the app reuse account-level policies and groups.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Access Application** -- the guarded application with its type, protected `domain` and additional `destinations`, attached policies in evaluation order, session and cookie settings, and (for `saas` types) the SAML/OIDC federation configuration whose issued credentials surface as outputs

Policies themselves are separate resources -- this component attaches existing CloudflareZeroTrustAccessPolicy resources; it does not create them.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Access edit permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust subscription** -- Access applications require a Zero Trust plan enabled on the Cloudflare account.
- **At least one Access policy** -- create a CloudflareZeroTrustAccessPolicy first (or have an existing policy ID); an application with no attached allow policy admits no one.
- **A zone for the protected domain (self-hosted types only)** -- `domain` must be a hostname in a zone on this account, resolving to the origin (directly or through a Cloudflare Tunnel).

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Access Application**, and click **Deploy**. The creation wizard captures the scope (account or zone), the application type and protected domain, the attached policies, and session settings. Start from the **Self-hosted web application** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: internal-dashboard
  org: acme-corp
  env: prod
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  name: internal-dashboard
  type: self_hosted
  domain: dashboard.example.com
  sessionDuration: 24h
  policies:
    - policy:
        valueFrom:
          kind: CloudflareZeroTrustAccessPolicy
          name: allow-staff
          fieldPath: status.outputs.policy_id
      precedence: 1
```

```shell
planton apply -f cloudflare-zero-trust-access-application.yaml
```

This creates an account-scoped self-hosted application protecting `dashboard.example.com`, governed by the referenced `allow-staff` policy, with a 24-hour session. An Infra Job tracks the provisioning in real time.

### InfraChart

Deploy the group, policy, and application as one composition, wiring the application to the policy with ValueFromRef:

```yaml
spec:
  policies:
    - policy:
        valueFrom:
          kind: CloudflareZeroTrustAccessPolicy
          name: allow-staff
          fieldPath: status.outputs.policy_id
      precedence: 1
```

The InfraPipeline resolves the dependency graph, provisions the policy first, then creates the application with the resolved policy ID attached.

## Key Configuration

These are the most important decisions when configuring a Zero Trust Access application. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope (`accountId` vs `zoneId`)** -- Exactly one must be set. Account scope is the common case: the application can reuse account-level policies and Access groups. Zone scope pins the application to a single zone; `zoneId` accepts a literal or a CloudflareDnsZone reference.

**Type (`type`)** -- Defaults to `self_hosted` and determines which other fields apply: `domain` is required for `self_hosted`/`ssh`/`vnc`/`rdp`, `saasApp` is required for `saas` and `dash_sso`, and the launcher styling fields only matter for `app_launcher`. Choosing the wrong type is not a tweak -- most of the spec silently stops applying.

**Policies (`policies`)** -- Standalone CloudflareZeroTrustAccessPolicy resources attached in evaluation order (`precedence`, lower first; list order breaks ties). Access denies by default, so an application with no allow policy locks everyone out -- including you.

**Session duration (`sessionDuration`)** -- A duration string like `24h`; empty uses the account default. Shorten it for sensitive applications; the `mfaConfig.sessionDuration` re-prompt is a separate, tighter clock for the second factor.

**Identity providers (`allowedIdps`, `autoRedirectToIdentity`)** -- Empty allows every configured IdP. Restricting to one IdP and setting `autoRedirectToIdentity` skips the chooser page, which is the polished single-IdP experience but locks sign-in to that provider's availability.

**SaaS federation (`saasApp`)** -- For `saas` types, choose `authType` `saml` or `oidc`. OIDC issues a `client_id`/`client_secret` pair as outputs to paste into the SaaS provider's SSO settings; SAML exports the SSO endpoint, entity ID, and public key instead.

**CORS (`corsHeaders` vs `optionsPreflightBypass`)** -- Mutually exclusive. Bypass lets preflight requests skip Access entirely; explicit CORS headers keep Access in the path. Pick bypass only when the protected API must serve browsers from other origins that cannot carry the Access cookie.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareZeroTrustAccessPolicy** | `policies[].policy` | `status.outputs.policy_id` |
| **CloudflareDnsZone** (optional, zone scope) | `zoneId` | `status.outputs.zone_id` |

`allowedIdps[]` and `scimConfig.idpUid` also accept references to identity-provider outputs alongside literal IdP IDs.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `aud` | The application's audience (AUD) tag | Validating the Cloudflare Access JWT in the origin service or Worker behind the app |
| `saas_client_id` | OAuth client ID (SaaS OIDC apps) | Pasted into the SaaS provider's SSO settings |
| `saas_client_secret` | OAuth client secret (SaaS OIDC apps). Sensitive. | Pasted into the SaaS provider's SSO settings |
| `saas_sso_endpoint` | SSO endpoint URL (SaaS SAML apps) | Configuring the SaaS provider's SAML connection |
| `saas_public_key` | IdP-facing certificate (SaaS SAML apps) | Configuring the SaaS provider's SAML connection |
| `saas_idp_entity_id` | IdP entity ID (SaaS SAML apps) | Configuring the SaaS provider's SAML connection |

`status.outputs` also carries `application_id` and the resolved `domain`.

## Common Patterns

**Self-hosted web app behind Access** -- An account-scoped `self_hosted` application protecting an internal hostname, governed by a referenced allow policy, commonly fronting an origin published through a Cloudflare Tunnel. Start from the **Self-hosted web application** preset.

**SaaS federation (OIDC)** -- Cloudflare acts as the identity provider for a SaaS product: the application issues the OAuth client credentials as outputs, and the SaaS provider's SSO settings consume them. Start from the **SaaS application (OIDC)** preset.

## Works With

- [**Cloudflare Zero Trust Access Policy**](/infra-catalog/cloudflare-zero-trust-access-policy) -- attached via `policies[]` to decide who is admitted
- [**Cloudflare Zero Trust Access Group**](/infra-catalog/cloudflare-zero-trust-access-group) -- reusable identity groups the attached policies reference
- [**Cloudflare Zero Trust Access Identity Provider**](/infra-catalog/cloudflare-zero-trust-access-identity-provider) -- the IdPs `allowedIdps` restricts sign-in to
- [**Cloudflare Zero Trust Tunnel**](/infra-catalog/cloudflare-zero-trust-tunnel) -- publishes the private origin a self-hosted application protects
- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- scopes a zone-level application and hosts the protected hostname
