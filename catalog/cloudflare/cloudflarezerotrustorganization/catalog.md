# Cloudflare Zero Trust Organization

Configures the Zero Trust organization: the account-wide Access login experience — team domain, login page design, session defaults, MFA policy — plus the Access service-key rotation cadence. This is a singleton upsert over the account's real login infrastructure: Cloudflare creates the organization when Zero Trust is first enabled, applying this resource mutates it, and destroy abandons the configuration exactly as last applied. One field, `authDomain`, has account-wide blast radius — changing it logs every user out of everything.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Organization Configuration** — one upsert of the account (or zone) organization singleton, sending only the fields the spec sets
- **Key-Rotation Cadence** — created only when `keyRotationIntervalDays` is set; the Access service-key rotation configuration (account scope only), itself a singleton upsert with a no-op destroy

Neither surface has a delete at Cloudflare: destroy reverts nothing.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with an API token holding Account → Access: Organizations, Identity Providers, and Groups → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **Zero Trust already onboarded** — the organization is born when Zero Trust is first enabled on the account (the team-name onboarding step), never by this resource. Applying against a fresh account without that onboarding fails at the API.
- **Custom Access pages** (only for `customPages`) — the block-page UIDs come from Access custom pages managed outside this resource today.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Organization**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the account-or-zone scope, the team domain and login design, and the session, MFA, and rotation settings. Start from the **Branded login** preset in the [Presets](#presets) tab to pre-populate the everyday shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustOrganization
metadata:
  name: acme-zero-trust-org
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  authDomain: acme.cloudflareaccess.com
  name: Acme Zero Trust
  sessionDuration: 24h
```

```shell
planton apply -f organization.yaml
```

This asserts the team domain (`acme.cloudflareaccess.com`), the login page's display name, and a 24-hour Access session on the organization the account already carries — every unset field keeps its live value. An Infra Job tracks the provisioning in real time.

### InfraChart

For a zone-scoped organization whose zone is deployed in the same InfraPipeline, wire `zoneId` with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  authDomain: acme.cloudflareaccess.com
  name: Acme Zero Trust
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then applies the organization configuration against the resolved zone. Note that a zone-scoped organization cannot be adopted by import, and it cannot carry `keyRotationIntervalDays`.

## Key Configuration

These are the most important decisions when configuring the Zero Trust organization. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**`authDomain` is the blast radius** — the team domain is the URL every user signs in through and every WARP client enrolls against, declared as the FULL domain including the `.cloudflareaccess.com` suffix (the form Cloudflare's API returns). It is required on every write — the upsert is a full-body PUT and omitting it is rejected (API error 11004) — so declare your CURRENT team domain unless you deliberately intend a rename. Changing it invalidates active sessions and breaks bookmarks, IdP redirect URIs, and WARP enrollment in one apply. Treat it like a production hostname: set it once, correctly, and never touch it casually.

**Destroy reverts nothing** — there is no "undo by destroy" on this singleton: the organization keeps the last-applied configuration forever. To revert a setting, apply the previous value. The folded key-rotation cadence behaves identically.

**Unset means unmanaged** — fields you leave unset are never sent, so dashboard-set values survive and partial adoption is safe (manage only the login design, say). The flip side: removing a field from the manifest does not clear it at Cloudflare — apply the empty or default value explicitly instead.

**MFA pairing rules are API-side** — `mfaRequiredForAllApps` needs MFA enabled with at least one allowed authenticator and an MFA session duration configured. And Cloudflare rejects `allowedAuthenticators` containing only `ssh_piv_key` while the organization has any non-infrastructure applications — PIV keys pair with infrastructure apps only.

**The dashboard lock** — `isUiReadOnly: true` makes every Zero Trust setting read-only in the dashboard regardless of user permission, leaving the API (and therefore IaC) as the only write path — the natural companion of managing the organization from a manifest. Set `uiReadOnlyToggleReason` so dashboard users see why.

**Unmatched-request posture** — `denyUnmatchedRequests: true` denies requests that match no Access application instead of passing them through, with `denyUnmatchedRequestsExemptedZoneNames` carving out exceptions; the fail-closed stance for accounts where everything should sit behind Access.

**Key rotation cadence** — `keyRotationIntervalDays` (21–365) folds the Access service-key rotation surface into this resource; account scope only. Unset leaves the account's current cadence unmanaged.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** (zone-scoped organizations) | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `auth_domain` | The team domain, without the `.cloudflareaccess.com` suffix | Access application and WARP enrollment configuration |
| `account_id` | The account the organization was applied to (empty for zone scope) | Import recipes and account-scoped siblings |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Branded login** — brand the Access login page (colors, logo, header/footer), set the session default, and leave MFA, custom pages, and key rotation exactly as the dashboard has them. Start from the **Branded login** preset.

**MFA hardening** — MFA required for every application by default (security keys or TOTP), unmatched requests denied, the dashboard locked read-only so IaC is the only write path, and the service key rotating every 30 days. Start from the **MFA hardening** preset.

**Partial adoption** — manage one surface (say, `sessionDuration` and `autoRedirectToIdentity`) while the rest of the organization stays dashboard-owned; safe because unset fields are never sent.

## Works With

- [**Cloudflare Zero Trust Access Identity Provider**](/infra-catalog/cloudflare-zero-trust-access-identity-provider) — the sign-in methods behind the login page this resource styles
- [**Cloudflare Zero Trust Access Application**](/infra-catalog/cloudflare-zero-trust-access-application) — the doors this login guards
- [**Cloudflare Zero Trust Access Policy**](/infra-catalog/cloudflare-zero-trust-access-policy) — who gets in
- [**Cloudflare Zero Trust Gateway Settings**](/infra-catalog/cloudflare-zero-trust-gateway-settings) — the traffic-filtering half of Zero Trust
