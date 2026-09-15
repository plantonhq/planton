# GcpIdentityPlatformConfig

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpIdentityPlatformConfigSpec defines a project's Identity Platform
configuration — the sign-in methods (email/password, phone, anonymous),
authorized domains, multi-factor authentication, blocking functions,
SMS-region policy, quotas, and the identity providers (Google, Facebook,
OIDC, SAML) end users can authenticate with.

This is a PROJECT SINGLETON with one-way creation: applying it for the
first time permanently initializes Identity Platform on the project
(the project must have billing enabled), and GCP provides no way to
de-initialize — destroying this resource abandons the configuration in
place rather than deleting anything. Every setting remains freely
updatable after initialization.

The three identity-provider lists compose the project-level IdP config
resources, so one manifest takes a project from nothing to a working
sign-in surface. Tenant-scoped configuration lives on the separate
GcpIdentityPlatformTenant kind (enable multi_tenant.allow_tenants here
first).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIdentityPlatformConfig
metadata:
  name: my-sample-auth-config
spec:
  # GCP project whose Identity Platform is configured. Must have BILLING
  # enabled — initialization fails without it. Omit to use the provider's
  # default project. NOTE: the first deploy PERMANENTLY initializes
  # Identity Platform on this project; there is no de-initialize.
  projectId:
    value: my-gcp-project-123

  # First-party sign-in methods. Each arm present here is sent explicitly
  # (including enabled=false); arms omitted stay unmanaged.
  signIn:
    email:
      # Explicit send — a false here actively disables the method.
      enabled: true
      # When false, users can sign in via email link alone.
      passwordRequired: true
    phoneNumber:
      enabled: true
      # Fixed verification codes for CI and reviewers — no real SMS.
      # Never ship real numbers here.
      testPhoneNumbers:
        "+15555550100": "123456"
    anonymous:
      # Guest accounts, typically upgraded to a real provider later.
      enabled: true
    # Most apps keep this false so an email maps to exactly one account.
    allowDuplicateEmails: false

  # Domains authorized for OAuth redirects and hosted sign-in flows.
  # localhost and the Firebase subdomains are authorized by default when
  # this list is empty.
  authorizedDomains:
    - app.example.com

  # Multi-factor authentication policy.
  mfa:
    # DISABLED, ENABLED (users may enroll), or MANDATORY.
    state: ENABLED
    # Only PHONE_SMS is accepted here; TOTP goes through providerConfigs.
    enabledProviders:
      - PHONE_SMS
    providerConfigs:
      - state: ENABLED
        totpProviderConfig:
          # Adjacent 30-second code windows accepted (clock-skew
          # tolerance); GCP allows 0-10.
          adjacentIntervals: 5

  # Cloud Functions invoked synchronously during sign-up/sign-in — they
  # can allow, block, or modify the operation. Their latency is added to
  # every affected authentication.
  blockingFunctions:
    triggers:
      # eventType: beforeCreate (before an account is created) or
      # beforeSignIn (before a sign-in completes).
      - eventType: beforeSignIn
        # A literal HTTPS URL or a GcpCloudFunction reference (its
        # function_url output is exactly this value).
        functionUri:
          value: https://us-central1-my-gcp-project-123.cloudfunctions.net/auth-gatekeeper
    # Forward only the tokens the function's logic actually inspects.
    forwardInboundCredentials:
      accessToken: false
      idToken: true
      refreshToken: false

  # Temporary sign-up ceiling — GCP applies the three fields as one unit,
  # so all are set together (quota 1-1000).
  signUpQuota:
    quota: 100
    # Window length as a seconds duration.
    quotaDuration: 86400s
    # When the override takes effect (RFC3339 UTC).
    startTime: "2026-09-01T00:00:00Z"

  # Which regions may receive SMS — the toll-fraud control. Exactly one
  # arm: allowlistOnly (allow only these) or allowByDefault (deny these).
  smsRegionConfig:
    allowlistOnly:
      # Two-letter CLDR region codes.
      allowedRegions:
        - US
        - DE

  # Restrictions on what client apps can do through the Identity Toolkit
  # API directly.
  clientPermissions:
    # When true, accounts are created only by your backend (admin SDK).
    disabledUserSignup: false
    # When true, account deletion happens only through your backend.
    disabledUserDeletion: false

  # Whether sign-in/sign-up requests are written to Cloud Logging. Sent
  # explicitly when set (true AND false reach the API); omit to leave
  # GCP's current value unmanaged.
  requestLoggingEnabled: true

  # Multi-tenancy: allowTenants must be true before any
  # GcpIdentityPlatformTenant can be created in this project.
  multiTenant:
    allowTenants: true
    # Optional GCP location for tenant data:
    # defaultTenantLocation: global

  # Delete anonymous users automatically after ~30 days of inactivity.
  autodeleteAnonymousUsers: true

  # Well-known identity providers enabled project-wide. clientId and
  # clientSecret come from each provider's own developer console —
  # consent-screen OAuth clients have no programmatic creation path.
  defaultSupportedIdps:
    # idpId is one of the ten canonical values (apple.com, facebook.com,
    # gc.apple.com, github.com, google.com, linkedin.com, microsoft.com,
    # playgames.google.com, twitter.com, yahoo.com). Immutable.
    - idpId: google.com
      clientId: 000000000000-placeholder.apps.googleusercontent.com
      # Managed secret — stored as a reference, resolved at deploy.
      clientSecret: GOCSPX-placeholder

  # Custom OIDC identity providers. Names must start with "oidc." and
  # are immutable.
  oauthIdpConfigs:
    - name: oidc.corp-sso
      displayName: Corporate SSO
      # Where Identity Platform fetches the provider's discovery document.
      issuer: https://accounts.example.com
      clientId: corp-sso-client
      # Required for the authorization-code flow below; managed secret.
      clientSecret: placeholder-oidc-secret
      responseType:
        # Authorization-code flow — the more secure choice (needs the
        # clientSecret above).
        code: true
        idToken: false

  # Inbound SAML identity providers (enterprise SSO). Names must match
  # saml.<slug> (lowercase start, alphanumeric end) and are immutable.
  inboundSamlConfigs:
    - name: saml.okta-prod
      displayName: Okta (Production)
      idpConfig:
        # Both values come from the IdP's metadata XML.
        idpEntityId: http://www.okta.com/exkexample
        ssoUrl: https://example.okta.com/app/example/sso/saml
        # Whether to sign outbound authentication requests.
        signRequest: false
        # The IdP's public signing certificates (PEM) — not secrets.
        idpCertificates:
          - x509Certificate: |
              -----BEGIN CERTIFICATE-----
              MIIC...placeholder...
              -----END CERTIFICATE-----
      spConfig:
        # Where the IdP posts SAML responses — must be https://.
        callbackUri: https://app.example.com/__/auth/handler
        # This project's SP entity ID, as registered with the IdP.
        spEntityId: app-example-com

  # Governs ONLY the composed IdP configs above — the config singleton
  # itself cannot be deleted (destroy always abandons it in place).
  # DELETE (default), PREVENT, or ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.signIn` | `GcpIdentityPlatformConfigSignIn` |  |  |  |
| `spec.signIn.email` | `GcpIdentityPlatformConfigSignInEmail` |  |  |  |
| `spec.signIn.email.enabled` | `bool` | yes |  |  |
| `spec.signIn.email.passwordRequired` | `bool` |  |  |  |
| `spec.signIn.phoneNumber` | `GcpIdentityPlatformConfigSignInPhone` |  |  |  |
| `spec.signIn.phoneNumber.enabled` | `bool` | yes |  |  |
| `spec.signIn.phoneNumber.testPhoneNumbers` | `map<string, string>` |  |  |  |
| `spec.signIn.anonymous` | `GcpIdentityPlatformConfigSignInAnonymous` |  |  |  |
| `spec.signIn.anonymous.enabled` | `bool` | yes |  |  |
| `spec.signIn.allowDuplicateEmails` | `bool` |  |  |  |
| `spec.authorizedDomains` | `[]string` |  |  |  |
| `spec.mfa` | `GcpIdentityPlatformConfigMfa` |  |  |  |
| `spec.mfa.state` | `string` |  |  |  |
| `spec.mfa.enabledProviders` | `[]string` |  |  |  |
| `spec.mfa.providerConfigs` | `[]GcpIdentityPlatformConfigMfaProviderConfig` |  |  |  |
| `spec.mfa.providerConfigs[].state` | `string` |  |  |  |
| `spec.mfa.providerConfigs[].totpProviderConfig` | `GcpIdentityPlatformConfigMfaTotp` |  |  |  |
| `spec.mfa.providerConfigs[].totpProviderConfig.adjacentIntervals` | `int32` |  |  |  |
| `spec.blockingFunctions` | `GcpIdentityPlatformConfigBlockingFunctions` |  |  |  |
| `spec.blockingFunctions.triggers` | `[]GcpIdentityPlatformConfigBlockingFunctionTrigger` | yes |  |  |
| `spec.blockingFunctions.triggers[].eventType` | `string` | yes |  |  |
| `spec.blockingFunctions.triggers[].functionUri` | `string \| valueFrom` | yes |  | GcpCloudFunction (`status.outputs.function_url`) |
| `spec.blockingFunctions.forwardInboundCredentials` | `GcpIdentityPlatformConfigForwardInboundCredentials` |  |  |  |
| `spec.blockingFunctions.forwardInboundCredentials.accessToken` | `bool` |  |  |  |
| `spec.blockingFunctions.forwardInboundCredentials.idToken` | `bool` |  |  |  |
| `spec.blockingFunctions.forwardInboundCredentials.refreshToken` | `bool` |  |  |  |
| `spec.signUpQuota` | `GcpIdentityPlatformConfigSignUpQuota` |  |  |  |
| `spec.signUpQuota.quota` | `int64` |  |  |  |
| `spec.signUpQuota.quotaDuration` | `string` |  |  |  |
| `spec.signUpQuota.startTime` | `string` |  |  |  |
| `spec.smsRegionConfig` | `GcpIdentityPlatformConfigSmsRegionConfig` |  |  |  |
| `spec.smsRegionConfig.allowByDefault` | `GcpIdentityPlatformConfigSmsAllowByDefault` |  |  |  |
| `spec.smsRegionConfig.allowByDefault.disallowedRegions` | `[]string` |  |  |  |
| `spec.smsRegionConfig.allowlistOnly` | `GcpIdentityPlatformConfigSmsAllowlistOnly` |  |  |  |
| `spec.smsRegionConfig.allowlistOnly.allowedRegions` | `[]string` |  |  |  |
| `spec.clientPermissions` | `GcpIdentityPlatformConfigClientPermissions` |  |  |  |
| `spec.clientPermissions.disabledUserSignup` | `bool` |  |  |  |
| `spec.clientPermissions.disabledUserDeletion` | `bool` |  |  |  |
| `spec.requestLoggingEnabled` | `bool` |  |  |  |
| `spec.multiTenant` | `GcpIdentityPlatformConfigMultiTenant` |  |  |  |
| `spec.multiTenant.allowTenants` | `bool` |  |  |  |
| `spec.multiTenant.defaultTenantLocation` | `string` |  |  |  |
| `spec.autodeleteAnonymousUsers` | `bool` |  |  |  |
| `spec.defaultSupportedIdps` | `[]GcpIdentityPlatformConfigDefaultSupportedIdp` |  |  |  |
| `spec.defaultSupportedIdps[].idpId` | `string` | yes |  |  |
| `spec.defaultSupportedIdps[].clientId` | `string` | yes |  |  |
| `spec.defaultSupportedIdps[].clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.defaultSupportedIdps[].enabled` | `bool` |  | `true` |  |
| `spec.oauthIdpConfigs` | `[]GcpIdentityPlatformConfigOauthIdp` |  |  |  |
| `spec.oauthIdpConfigs[].name` | `string` | yes |  |  |
| `spec.oauthIdpConfigs[].displayName` | `string` |  |  |  |
| `spec.oauthIdpConfigs[].issuer` | `string` | yes |  |  |
| `spec.oauthIdpConfigs[].clientId` | `string` | yes |  |  |
| `spec.oauthIdpConfigs[].clientSecret` | `string` (sensitive) |  |  |  |
| `spec.oauthIdpConfigs[].enabled` | `bool` |  | `true` |  |
| `spec.oauthIdpConfigs[].responseType` | `GcpIdentityPlatformConfigOauthResponseType` |  |  |  |
| `spec.oauthIdpConfigs[].responseType.code` | `bool` |  |  |  |
| `spec.oauthIdpConfigs[].responseType.idToken` | `bool` |  |  |  |
| `spec.inboundSamlConfigs` | `[]GcpIdentityPlatformConfigInboundSaml` |  |  |  |
| `spec.inboundSamlConfigs[].name` | `string` | yes |  |  |
| `spec.inboundSamlConfigs[].displayName` | `string` | yes |  |  |
| `spec.inboundSamlConfigs[].enabled` | `bool` |  | `true` |  |
| `spec.inboundSamlConfigs[].idpConfig` | `GcpIdentityPlatformConfigSamlIdpConfig` | yes |  |  |
| `spec.inboundSamlConfigs[].idpConfig.idpEntityId` | `string` | yes |  |  |
| `spec.inboundSamlConfigs[].idpConfig.ssoUrl` | `string` | yes |  |  |
| `spec.inboundSamlConfigs[].idpConfig.signRequest` | `bool` |  |  |  |
| `spec.inboundSamlConfigs[].idpConfig.idpCertificates` | `[]GcpIdentityPlatformConfigSamlCertificate` |  |  |  |
| `spec.inboundSamlConfigs[].idpConfig.idpCertificates[].x509Certificate` | `string` |  |  |  |
| `spec.inboundSamlConfigs[].spConfig` | `GcpIdentityPlatformConfigSamlSpConfig` |  |  |  |
| `spec.inboundSamlConfigs[].spConfig.callbackUri` | `string` |  |  |  |
| `spec.inboundSamlConfigs[].spConfig.spEntityId` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.adoptExisting` | `bool` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project whose Identity Platform is configured. Can be a
literal project ID or a reference to a GcpProject resource. If
omitted, the provider's default project is used. The project must
have BILLING enabled — initialization fails without it.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.signIn

`GcpIdentityPlatformConfigSignIn`

First-party sign-in methods. Each arm you set is sent explicitly
(including enabled=false) so the project's state always matches the
manifest; arms you omit are left unmanaged.

### spec.signIn.email

`GcpIdentityPlatformConfigSignInEmail`

Email/password sign-in.

### spec.signIn.email.enabled

`bool` · required · optional (explicit presence)

Whether email sign-in is enabled. Required inside the arm: declaring
the arm takes the method under management, and the switch says which
way (false actively disables it).

- rule: {"required":true}

### spec.signIn.email.passwordRequired

`bool`

Whether a password is required for email accounts. When false,
users can sign in via email link alone.

### spec.signIn.phoneNumber

`GcpIdentityPlatformConfigSignInPhone`

Phone-number (SMS code) sign-in.

### spec.signIn.phoneNumber.enabled

`bool` · required · optional (explicit presence)

Whether phone sign-in is enabled. Required inside the arm: declaring
the arm takes the method under management, and the switch says which
way (false actively disables it).

- rule: {"required":true}

### spec.signIn.phoneNumber.testPhoneNumbers

`map<string, string>`

Test phone numbers mapped to fixed verification codes (e.g.
"+15555550100" -> "123456") — lets CI and reviewers exercise the
phone flow without receiving real SMS. Never ship real numbers here.

### spec.signIn.anonymous

`GcpIdentityPlatformConfigSignInAnonymous`

Anonymous (guest) sign-in — accounts created without credentials,
typically upgraded to a real provider later.

### spec.signIn.anonymous.enabled

`bool` · required · optional (explicit presence)

Whether anonymous (guest) sign-in is enabled. Required inside the arm:
declaring the arm takes the method under management, and the switch
says which way (false actively disables it).

- rule: {"required":true}

### spec.signIn.allowDuplicateEmails

`bool`

Whether multiple accounts may share one email address. Most apps
leave this false so an email maps to exactly one account.

### spec.authorizedDomains

`[]string`

Domains authorized for OAuth redirects and hosted sign-in flows
(e.g. "myapp.example.com"). localhost and the project's Firebase
subdomains are authorized by default when this list is left empty.

### spec.mfa

`GcpIdentityPlatformConfigMfa`

Multi-factor authentication policy for the project.

### spec.mfa.state

`string`

Project-wide MFA state (provider-validated values):
  "DISABLED"  -- MFA cannot be used
  "ENABLED"   -- users may enroll a second factor
  "MANDATORY" -- every user must present a second factor

- rule: mfa.state must be one of: DISABLED, ENABLED, MANDATORY

### spec.mfa.enabledProviders

`[]string`

Usable second factors. The provider accepts only "PHONE_SMS" here;
TOTP is configured through provider_configs.

- rule: {"repeated":{"items":{"cel":[{"id":"valid_mfa_provider","message":"mfa.enabled_providers entries must be PHONE_SMS (TOTP is configured via provider_configs)","expression":"this == 'PHONE_SMS'"}]}}}

### spec.mfa.providerConfigs

`[]GcpIdentityPlatformConfigMfaProviderConfig`

Per-provider MFA configuration (currently TOTP authenticator apps).

### spec.mfa.providerConfigs[].state

`string`

This provider's state: DISABLED, ENABLED, or MANDATORY.

- rule: provider_configs.state must be one of: DISABLED, ENABLED, MANDATORY

### spec.mfa.providerConfigs[].totpProviderConfig

`GcpIdentityPlatformConfigMfaTotp`

TOTP (authenticator-app) settings.

### spec.mfa.providerConfigs[].totpProviderConfig.adjacentIntervals

`int32`

How many adjacent 30-second code windows are accepted (clock-skew
tolerance). GCP accepts up to 10.

- rule: totp_provider_config.adjacent_intervals must be between 0 and 10

### spec.blockingFunctions

`GcpIdentityPlatformConfigBlockingFunctions`

Blocking functions — Cloud Functions invoked synchronously during
sign-up/sign-in that can allow, block, or modify the operation
(custom claims, domain allowlists, fraud checks).

### spec.blockingFunctions.triggers

`[]GcpIdentityPlatformConfigBlockingFunctionTrigger` · required

The trigger points and the Cloud Functions they invoke. GCP supports
the "beforeCreate" (before an account is created) and "beforeSignIn"
(before a sign-in completes) event types.

- rule: {"repeated":{"minItems":"1"}}

### spec.blockingFunctions.triggers[].eventType

`string` · required

The trigger point: "beforeCreate" or "beforeSignIn".

- rule: event_type must be beforeCreate or beforeSignIn
- rule: {"required":true}

### spec.blockingFunctions.triggers[].functionUri

`string | valueFrom` · required

The HTTPS endpoint of the Cloud Function to invoke — a literal URL
or a reference to a GcpCloudFunction resource (its function_url
output is exactly this value).

- references: GcpCloudFunction (`status.outputs.function_url`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudFunction, name: <that resource's name>, fieldPath: status.outputs.function_url}} -- a bare string does not parse

### spec.blockingFunctions.forwardInboundCredentials

`GcpIdentityPlatformConfigForwardInboundCredentials`

Which of the user's tokens are forwarded to the blocking function —
grant only what the function's logic actually inspects.

### spec.blockingFunctions.forwardInboundCredentials.accessToken

`bool`

Forward the user's OAuth access token.

### spec.blockingFunctions.forwardInboundCredentials.idToken

`bool`

Forward the user's OIDC ID token.

### spec.blockingFunctions.forwardInboundCredentials.refreshToken

`bool`

Forward the user's refresh token.

### spec.signUpQuota

`GcpIdentityPlatformConfigSignUpQuota`

Temporary sign-up quota override — all three fields are set together
(GCP needs the ceiling, the window length, and when it starts).

- rule: sign_up_quota requires quota, quota_duration, and start_time together — GCP applies the override as one unit

### spec.signUpQuota.quota

`int64`

Sign-ups allowed during the window — between 1 and 1000 (the API's
documented range).

- rule: sign_up_quota.quota must be between 1 and 1000

### spec.signUpQuota.quotaDuration

`string`

How long the override stays active, as a seconds duration (e.g.
"86400s" for one day).

- rule: sign_up_quota.quota_duration must be a seconds duration such as 86400s

### spec.signUpQuota.startTime

`string`

When the override takes effect (RFC3339 UTC, e.g.
"2026-09-01T00:00:00Z").

- rule: sign_up_quota.start_time must be an RFC3339 UTC timestamp such as 2026-09-01T00:00:00Z

### spec.smsRegionConfig

`GcpIdentityPlatformConfigSmsRegionConfig`

Which regions may receive SMS (verification codes, MFA) — the
toll-fraud control. Exactly one policy arm.

- rule: set exactly one of allow_by_default or allowlist_only

### spec.smsRegionConfig.allowByDefault

`GcpIdentityPlatformConfigSmsAllowByDefault`

Allow SMS to every region EXCEPT the listed ones.

### spec.smsRegionConfig.allowByDefault.disallowedRegions

`[]string`

Two-letter CLDR region codes to disallow (e.g. "KP", "RU").

### spec.smsRegionConfig.allowlistOnly

`GcpIdentityPlatformConfigSmsAllowlistOnly`

Allow SMS ONLY to the listed regions — the tighter toll-fraud
posture.

### spec.smsRegionConfig.allowlistOnly.allowedRegions

`[]string`

Two-letter CLDR region codes to allow (e.g. "US", "DE").

### spec.clientPermissions

`GcpIdentityPlatformConfigClientPermissions`

Restrictions on what client applications can do through the
Identity Toolkit API directly.

### spec.clientPermissions.disabledUserSignup

`bool`

When true, end users CANNOT sign up themselves through the API —
accounts are created only by your backend (admin SDK).

### spec.clientPermissions.disabledUserDeletion

`bool`

When true, end users CANNOT delete their own accounts through the
API — deletion happens only through your backend.

### spec.requestLoggingEnabled

`bool` · optional (explicit presence)

Whether sign-in/sign-up requests are written to Cloud Logging. Sent
explicitly when set (true or false); leave unset to keep GCP's
current value unmanaged.

### spec.multiTenant

`GcpIdentityPlatformConfigMultiTenant`

Multi-tenancy: allow_tenants must be true before any
GcpIdentityPlatformTenant can be created in this project.

### spec.multiTenant.allowTenants

`bool`

Whether this project may contain tenants. Must be true before any
GcpIdentityPlatformTenant is created in the project.

### spec.multiTenant.defaultTenantLocation

`string`

Default GCP location for tenant data (e.g. "global").

### spec.autodeleteAnonymousUsers

`bool`

When true, anonymous users are deleted automatically after ~30 days
of inactivity — keeps abandoned guest sessions from accumulating.

### spec.defaultSupportedIdps

`[]GcpIdentityPlatformConfigDefaultSupportedIdp`

Default supported identity providers (Google, Facebook, Apple, ...)
enabled for the whole project, each with the OAuth client obtained
from that provider's own developer console.

### spec.defaultSupportedIdps[].idpId

`string` · required

Which provider this is — the provider's canonical IdP ID:
apple.com, facebook.com, gc.apple.com, github.com, google.com,
linkedin.com, microsoft.com, playgames.google.com, twitter.com,
yahoo.com. Immutable: changing it replaces the IdP config.

- rule: idp_id must be one of: apple.com, facebook.com, gc.apple.com, github.com, google.com, linkedin.com, microsoft.com, playgames.google.com, twitter.com, yahoo.com
- rule: {"required":true}

### spec.defaultSupportedIdps[].clientId

`string` · required

The OAuth client ID issued by the identity provider's own developer
console (e.g. Google Cloud Console for google.com, Meta for
facebook.com). Consent-screen clients have no programmatic creation
path — obtaining these is a documented console step.

- rule: {"required":true}

### spec.defaultSupportedIdps[].clientSecret

`string` · required · sensitive

The OAuth client secret paired with client_id. A secret value: the
platform stores it as a managed-secret reference and resolves it
just-in-time at deploy — it never sits in plaintext in the control
plane. No Planton resource produces it (it comes from the external
provider's console), so it is a plain secret string.

- rule: {"required":true}

### spec.defaultSupportedIdps[].enabled

`bool` · optional (explicit presence)

Whether users can sign in with this provider (default true). Both
IaC engines send the value explicitly so behavior is identical
regardless of engine.

- default: `true`

### spec.oauthIdpConfigs

`[]GcpIdentityPlatformConfigOauthIdp`

Custom OIDC identity providers for the whole project.

### spec.oauthIdpConfigs[].name

`string` · required

Resource name for this OIDC config — must start with "oidc."
(e.g. "oidc.corp-sso"). The API's naming rule, validated up front.
Immutable.

- rule: oauth_idp_configs.name must start with 'oidc.' — e.g. oidc.corp-sso
- rule: {"required":true}

### spec.oauthIdpConfigs[].displayName

`string`

Human-readable name shown in consoles and sign-in UIs.

### spec.oauthIdpConfigs[].issuer

`string` · required

The OIDC issuer URL (e.g. "https://accounts.example.com") — where
Identity Platform fetches the provider's discovery document.

- rule: {"required":true}

### spec.oauthIdpConfigs[].clientId

`string` · required

The OAuth client ID registered with the OIDC provider.

- rule: {"required":true}

### spec.oauthIdpConfigs[].clientSecret

`string` · sensitive

The OAuth client secret — required for the authorization-code flow
(response_type.code), unused for the implicit id_token flow. A
secret value handled as a managed-secret reference end to end.

### spec.oauthIdpConfigs[].enabled

`bool` · optional (explicit presence)

Whether users can sign in with this provider (default true). Sent
explicitly by both engines.

- default: `true`

### spec.oauthIdpConfigs[].responseType

`GcpIdentityPlatformConfigOauthResponseType`

Which OAuth response type to request from the provider.

### spec.oauthIdpConfigs[].responseType.code

`bool`

Authorization-code flow — the provider returns a code exchanged
server-side (requires client_secret). The more secure flow.

### spec.oauthIdpConfigs[].responseType.idToken

`bool`

Implicit flow — the provider returns an ID token directly.

### spec.inboundSamlConfigs

`[]GcpIdentityPlatformConfigInboundSaml`

Inbound SAML identity providers (enterprise SSO) for the whole
project.

### spec.inboundSamlConfigs[].name

`string` · required

Resource name for this SAML config — must start with "saml.",
contain only alphanumerics/hyphens/underscores/periods, and the part
after "saml." must start with a lowercase letter, end alphanumeric,
and be at least 2 characters (the API's naming rule, validated up
front). E.g. "saml.okta-prod". Immutable.

- rule: inbound_saml_configs.name must start with 'saml.' followed by a lowercase letter, contain only alphanumerics/hyphens/underscores/periods, end alphanumeric, and have at least 2 characters after the prefix — e.g. saml.okta-prod
- rule: {"required":true}

### spec.inboundSamlConfigs[].displayName

`string` · required

Human-readable name shown in consoles and sign-in UIs.

- rule: {"required":true}

### spec.inboundSamlConfigs[].enabled

`bool` · optional (explicit presence)

Whether users can sign in with this provider (default true). Sent
explicitly by both engines.

- default: `true`

### spec.inboundSamlConfigs[].idpConfig

`GcpIdentityPlatformConfigSamlIdpConfig` · required

The external identity provider's side of the SAML exchange.

- rule: {"required":true}

### spec.inboundSamlConfigs[].idpConfig.idpEntityId

`string` · required

The IdP's entity ID from its metadata XML.

- rule: {"required":true}

### spec.inboundSamlConfigs[].idpConfig.ssoUrl

`string` · required

The IdP's SSO URL — where users are sent to authenticate.

- rule: {"required":true}

### spec.inboundSamlConfigs[].idpConfig.signRequest

`bool`

Whether to sign outbound authentication requests.

### spec.inboundSamlConfigs[].idpConfig.idpCertificates

`[]GcpIdentityPlatformConfigSamlCertificate`

The IdP's X.509 signing certificates (PEM) used to verify SAML
responses. Public certificates, not secrets.

### spec.inboundSamlConfigs[].idpConfig.idpCertificates[].x509Certificate

`string`

The certificate in PEM format.

### spec.inboundSamlConfigs[].spConfig

`GcpIdentityPlatformConfigSamlSpConfig`

This project's side of the SAML exchange (the service provider).

### spec.inboundSamlConfigs[].spConfig.callbackUri

`string`

Where the IdP posts SAML responses — must be an https:// URL (the
API's rule, validated up front).

- rule: sp_config.callback_uri must start with https://

### spec.inboundSamlConfigs[].spConfig.spEntityId

`string`

This project's SP entity ID, as registered with the IdP.

### spec.deletionPolicy

`string`

Deletion policy for the COMPOSED identity-provider configs (the
default-supported/OIDC/SAML entries above) — the project config
itself cannot be deleted (destroy always abandons it in place):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the IdP configs are deleted; users can no longer sign
               in through them (accounts themselves are untouched)
  "PREVENT" -- destroy FAILS; protects a live sign-in surface
  "ABANDON" -- the IdP configs are removed from management but keep
               working in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.adoptExisting

`bool`

Adopt the project's EXISTING Identity Platform configuration instead
of initializing it. Initialization is one-way and once-only: GCP
rejects a second initializeAuth with 400 "Identity Platform has
already been enabled for this project" (live-verified), and there is
no de-initialize. Set true for any project where Identity Platform
was ever enabled — by a console click, a Firebase Auth setup, or a
previous deployment of this kind (destroy abandons the configuration
in place, so a re-deploy after destroy also needs it). The module
then imports the singleton (projects/{project}/config) and applies
this spec as an update. Leave false only for a project whose
Identity Platform has never been enabled.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpIdentityPlatformConfig, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.config_name` | `string` | The configuration's resource name: projects/{project}/config. |
| `status.outputs.api_key` | `string` | The auto-provisioned API key client apps initialize the Identity Platform / Firebase Auth SDK with. A live credential in the API-key sense (restrict it by domain/app in the console); the engines mark it secret in state. |
| `status.outputs.firebase_subdomain` | `string` | The project's Firebase subdomain (e.g. "my-project" for my-project.firebaseapp.com) — the default hosted sign-in domain. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.blockingFunctions.triggers[].functionUri` | GcpCloudFunction | `status.outputs.function_url` |

## See Also

- [Overview](../README.md)
