# CloudflareZeroTrustOrganization

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustOrganizationSpec configures the Zero Trust organization:
the account-wide (or zone-wide) Access login experience -- the team domain,
login page design, session defaults, MFA policy, and the Access
service-key rotation cadence.

Three provider facts shape how this resource behaves:
  - It is a SINGLETON UPSERT: Cloudflare has no create call for an
    organization -- both create and update are the same PUT, so applying
    this resource mutates whatever organization already exists on the
    account. The organization itself is born when Zero Trust is first
    enabled (the team-name onboarding step), never by this resource.
  - Destroy is a NO-OP: deleting this resource abandons the live
    configuration exactly as last applied; nothing is reverted at
    Cloudflare.
  - The API accepts every field on the same upsert -- there is no partial
    surface. Unset fields are simply not sent, leaving the live value
    untouched.

## Example

```yaml
# Complete example manifest for CloudflareZeroTrustOrganization. Configures
# the account's Zero Trust organization: the Access login page design,
# session defaults, MFA policy, and the Access service-key rotation cadence.
# A singleton upsert: applying mutates the organization the account already
# carries, and destroy abandons the configuration (nothing is reverted).
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustOrganization
metadata:
  name: acme-zero-trust-org
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  auth_domain: acme.cloudflareaccess.com
  name: Acme Zero Trust
  session_duration: 24h
  warp_auth_session_duration: 12h
  deny_unmatched_requests: false
  login_design:
    background_color: "#0f172a"
    text_color: "#e2e8f0"
    header_text: Sign in to Acme
    footer_text: Managed by Acme IT
  mfa_config:
    allowed_authenticators:
      - totp
      - security_key
    session_duration: 12h
  allow_authenticate_via_warp: true
  auto_redirect_to_identity: false
  key_rotation_interval_days: 90
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` |  |  |  |
| `spec.zoneId` | `string \| valueFrom` |  |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.authDomain` | `string` | yes |  |  |
| `spec.name` | `string` |  |  |  |
| `spec.sessionDuration` | `string` |  |  |  |
| `spec.warpAuthSessionDuration` | `string` |  |  |  |
| `spec.userSeatExpirationInactiveTime` | `string` |  |  |  |
| `spec.denyUnmatchedRequests` | `bool` |  |  |  |
| `spec.denyUnmatchedRequestsExemptedZoneNames` | `[]string` |  |  |  |
| `spec.customPages` | `CloudflareZeroTrustOrganizationCustomPages` |  |  |  |
| `spec.customPages.forbidden` | `string` |  |  |  |
| `spec.customPages.identityDenied` | `string` |  |  |  |
| `spec.loginDesign` | `CloudflareZeroTrustOrganizationLoginDesign` |  |  |  |
| `spec.loginDesign.backgroundColor` | `string` |  |  |  |
| `spec.loginDesign.textColor` | `string` |  |  |  |
| `spec.loginDesign.logoPath` | `string` |  |  |  |
| `spec.loginDesign.headerText` | `string` |  |  |  |
| `spec.loginDesign.footerText` | `string` |  |  |  |
| `spec.mfaConfig` | `CloudflareZeroTrustOrganizationMfaConfig` |  |  |  |
| `spec.mfaConfig.allowedAuthenticators` | `[]string` |  |  |  |
| `spec.mfaConfig.sessionDuration` | `string` |  |  |  |
| `spec.mfaConfig.amrMatchingSessionDuration` | `string` |  |  |  |
| `spec.mfaConfig.requiredAaguids` | `string` |  |  |  |
| `spec.mfaSshPivKeyRequirements` | `CloudflareZeroTrustOrganizationMfaSshPivKeyRequirements` |  |  |  |
| `spec.mfaSshPivKeyRequirements.pinPolicy` | `string` |  |  |  |
| `spec.mfaSshPivKeyRequirements.touchPolicy` | `string` |  |  |  |
| `spec.mfaSshPivKeyRequirements.sshKeyType` | `[]string` |  |  |  |
| `spec.mfaSshPivKeyRequirements.sshKeySize` | `[]int64` |  |  |  |
| `spec.mfaSshPivKeyRequirements.requireFipsDevice` | `bool` |  |  |  |
| `spec.allowAuthenticateViaWarp` | `bool` |  |  |  |
| `spec.autoRedirectToIdentity` | `bool` |  |  |  |
| `spec.mfaRequiredForAllApps` | `bool` |  |  |  |
| `spec.isUiReadOnly` | `bool` |  |  |  |
| `spec.uiReadOnlyToggleReason` | `string` |  |  |  |
| `spec.keyRotationIntervalDays` | `int32` |  |  |  |

## Field Details

### spec.accountId

`string`

The Cloudflare account that owns this organization. Set this for the
account-wide organization (the common case). Mutually exclusive with
zone_id.

- rule: account_id must be a 32-character hex string

### spec.zoneId

`string | valueFrom`

The Cloudflare zone this organization is scoped to, as a literal zone ID
or a reference to a CloudflareDnsZone. Mutually exclusive with
account_id. NOTE: a zone-scoped organization cannot be adopted by
import -- the provider's importer is account-only.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.authDomain

`string` · required

The FULL team domain users sign in through, including the
.cloudflareaccess.com suffix (e.g. "acme.cloudflareaccess.com" -- the
form Cloudflare's API returns and the provider's own tests write).
This is the closest thing the organization has to an identity: the
upsert mutates whatever organization exists, keyed only by the account
or zone.

REQUIRED because Cloudflare requires it on EVERY write: the upsert is
a full-body PUT, and omitting auth_domain is rejected with API error
11004 "access.api.error.invalid_auth_domain" (live-measured). Declare
your CURRENT team domain unless you deliberately intend to rename it
-- a rename changes the login URL for every Access user.

- rule: {"required":true}

### spec.name

`string`

The organization's display name, shown on the Access login page.

Declare it on any account whose organization already carries a name
(live-measured): the API echoes the existing name on every read, so a
manifest that omits this field sees a permanent "name -> null" re-plan
-- the same declare-it-or-accept-drift class as other echoed optional
fields on singleton upserts.

### spec.sessionDuration

`string`

How long an Access session lasts before users re-authenticate, as a Go
duration (e.g. "24h", "2h45m"; ns/us/ms/s/m/h units).

### spec.warpAuthSessionDuration

`string`

How long a WARP-authenticated session lasts, as a duration using minutes
or hours only (e.g. "30m", "12h").

### spec.userSeatExpirationInactiveTime

`string`

How long a user can be inactive before their seat is released, as a Go
duration. Cloudflare requires at least one month (730h).

### spec.denyUnmatchedRequests

`bool` · optional (explicit presence)

When true, requests that match no Access application are denied instead
of passed through.

### spec.denyUnmatchedRequestsExemptedZoneNames

`[]string`

Zone names exempted from deny_unmatched_requests.

### spec.customPages

`CloudflareZeroTrustOrganizationCustomPages`

Custom Access pages shown instead of the default block screens. Values
are custom-page UIDs (Cloudflare's Access custom pages, managed outside
this resource today).

### spec.customPages.forbidden

`string`

The UID of the custom page shown when a user is denied after failing a
non-identity rule.

### spec.customPages.identityDenied

`string`

The UID of the custom page shown when a user's identity is denied.

### spec.loginDesign

`CloudflareZeroTrustOrganizationLoginDesign`

The look of the Access login page.

### spec.loginDesign.backgroundColor

`string`

The page background color, as a CSS color (e.g. "#1e293b").

### spec.loginDesign.textColor

`string`

The text color, as a CSS color.

### spec.loginDesign.logoPath

`string`

The path to the logo shown on the page.

### spec.loginDesign.headerText

`string`

The heading text above the login options.

### spec.loginDesign.footerText

`string`

The footer text below the login options.

### spec.mfaConfig

`CloudflareZeroTrustOrganizationMfaConfig`

Organization-wide multi-factor authentication policy.

### spec.mfaConfig.allowedAuthenticators

`[]string`

The authenticator methods users may satisfy MFA with. Cloudflare accepts
exactly: totp, biometrics, security_key, ssh_piv_key.

- rule: allowed_authenticators entries must be drawn from totp, biometrics, security_key, ssh_piv_key

### spec.mfaConfig.sessionDuration

`string`

How long an MFA assertion stays valid, as a duration using minutes or
hours only (min "0m", max "720h").

### spec.mfaConfig.amrMatchingSessionDuration

`string`

How long an authentication-method-reference match stays valid, as a
duration using minutes or hours only (min "0m", max "720h").

### spec.mfaConfig.requiredAaguids

`string`

The NAME of a Cloudflare List of FIDO2 AAGUIDs restricting which
security keys are accepted. A single list name, not the AAGUIDs
themselves.

### spec.mfaSshPivKeyRequirements

`CloudflareZeroTrustOrganizationMfaSshPivKeyRequirements`

Hardware requirements applied when ssh_piv_key is an allowed
authenticator.

### spec.mfaSshPivKeyRequirements.pinPolicy

`string`

When the key requires its PIN: never, once (per session), or always.

- rule: pin_policy must be one of never, once, always

### spec.mfaSshPivKeyRequirements.touchPolicy

`string`

When the key requires a touch: never, always, or cached.

- rule: touch_policy must be one of never, always, cached

### spec.mfaSshPivKeyRequirements.sshKeyType

`[]string`

The accepted key algorithms: ecdsa, ed25519, rsa.

- rule: ssh_key_type entries must be drawn from ecdsa, ed25519, rsa

### spec.mfaSshPivKeyRequirements.sshKeySize

`[]int64`

The accepted key sizes in bits: 256, 384, 521 (ECDSA), 2048, 3072, 4096
(RSA).

- rule: ssh_key_size entries must be drawn from 256, 384, 521, 2048, 3072, 4096

### spec.mfaSshPivKeyRequirements.requireFipsDevice

`bool` · optional (explicit presence)

When true, only FIPS-certified hardware keys are accepted.

### spec.allowAuthenticateViaWarp

`bool` · optional (explicit presence)

When true, users can authenticate to Access applications through an
active WARP session instead of the login page.

### spec.autoRedirectToIdentity

`bool` · optional (explicit presence)

When true, users skip the identity-provider picker and are sent straight
to the sole configured provider.

### spec.mfaRequiredForAllApps

`bool` · optional (explicit presence)

When true, global MFA settings apply to applications by default. The
organization must have MFA enabled with at least one authentication
method and a session duration configured. Cloudflare rejects
allowed_authenticators containing ONLY ssh_piv_key when the organization
has any non-infrastructure applications (PIV keys pair with
infrastructure apps only).

### spec.isUiReadOnly

`bool` · optional (explicit presence)

When true, every Zero Trust setting becomes read-only in the dashboard
regardless of user permission -- changes can then only land through the
API or IaC. The natural companion of managing the organization from
this resource.

### spec.uiReadOnlyToggleReason

`string`

The reason shown in the dashboard for the read-only lock.

### spec.keyRotationIntervalDays

`int32` · optional (explicit presence)

The Access service-key rotation cadence in days (21-365). This folds
Cloudflare's key-configuration surface into the organization: like the
organization itself it is a singleton upsert with a no-op destroy.
Unset leaves the account's current cadence unmanaged. Account scope
only.

- rule: {"int32":{"lte":365,"gte":21}}

## Validation Rules

- `spec.account_xor_zone`: set exactly one of account_id or zone_id
- `spec.key_rotation_account_only`: key_rotation_interval_days applies only to an account-scoped organization -- unset it or switch to account_id

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustOrganization, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.auth_domain` | `string` | The team domain users sign in through, without the .cloudflareaccess.com suffix -- what Access applications and WARP enrollment reference. |
| `status.outputs.account_id` | `string` | The Cloudflare account the organization was applied to (empty for a zone-scoped organization). The harness and import recipes key on it. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
