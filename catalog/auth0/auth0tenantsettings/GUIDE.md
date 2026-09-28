# Auth0 Tenant Settings Guide

## Security
## Platform Security Posture

The certifications below are Auth0's own published claims about their hosted platform (verify current status on Auth0's compliance page). They describe the vendor's service — never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Auth0's published certifications and security standards:

- SOC 2 Type II (annual audit)
- ISO 27001, ISO 27018 (privacy controls)
- HIPAA BAA available on enterprise plans
- PCI DSS Level 1 Service Provider
- FedRAMP Authorized (moderate baseline)
- CSA STAR Level 2
- GDPR compliant with Data Processing Agreement

## Data Protection

- **Data residency**: US, EU, AU regions
- **Encryption in transit**: TLS 1.2+
- **Encryption at rest**: AES-256
- **Penetration testing**: Annual third-party assessments

## Tenant-Settings Security Notes

### Public by Design

The name, logo, languages and support contacts are shown to anyone who opens the tenant's login page. Put nothing in them you would not publish.

### The Logo Is Fetched From Where You Host It

The login page loads `pictureUrl` from its host on every visit. Host the logo somewhere you control, over HTTPS, and keep the URL answering: a moved or deleted image breaks the page's look (never the sign-in itself).

### A Support Address Is an Invitation

People who can't sign in will write to `supportEmail`. Point it at a monitored mailbox that knows it must never ask for a password.

### Open Registration Is Open to Everyone

`flags.enableDynamicClientRegistration` opens `/oidc/register` with no token: anyone can register an application in the tenant. Auth0 contains each one as a third-party application -- it reaches only the APIs and scopes the tenant's default third-party permissions grant, and signs in only through domain-level connections -- so those two settings are the real boundary. Configure them before turning registration on, and grant third parties nothing you would not grant a stranger.

### Session Lifetimes Are the Unattended-Laptop Setting

`idleSessionLifetime` is how long a session left open on a shared or lost device keeps working; `sessionLifetime` is how long before anyone must prove who they are again. Shorter is safer and costs a sign-in; the Short Idle Sessions preset is a starting point for consoles.

### Leave the Legacy Paths Off

The `flags.allowLegacy*` flags, `flags.enableIdtokenApi2` and `flags.enableLegacyProfile` keep old grant types and token behaviors alive. Declare them `false` unless an integration you can name still needs one, and `flags.enablePublicSignupUserExistsError: false` so the signup API does not tell strangers which addresses have accounts.

## Adoption

Every tenant exists before this resource does, so the first deploy of an Auth0TenantSettings is always an adoption: the resource starts managing a tenant that already carries every setting.

### Unset Means Unmanaged -- With Exceptions

A field the spec leaves unset is never sent, and the tenant keeps its value. The provider cannot keep that promise for six settings (and one nested block), and these reset (or, for the login route, keep planning a change) on the first deploy and after any import when the spec leaves them unset:

| Setting | What an unset field does on a tenant that has one |
|---|---|
| `defaultRedirectionUri` | Plans its removal on every run; Auth0 keeps the value, as nothing is sent |
| `skipNonVerifiableCallbackUriConfirmationPrompt` | Resets to Auth0's unset state |
| `mtls` | Removes the mTLS endpoint aliases on the first deploy (not after an import) |
| `errorPage` | Returns the tenant to Auth0's default error page |
| `defaultTokenQuota` | Removes the tenant's default token quotas |
| `countryCodes` | Removes the tenant's phone country filter |
| `sessions.anonymous` (when `sessions` is declared) | Removes the anonymous-session settings |

Before the first deploy against a tenant that has any of these, read them (`GET /api/v2/tenants/settings`, or the dashboard's Settings tabs) and declare the live values in the manifest.

### Importing a Tenant Already Managed Elsewhere

If Terraform or Pulumi already manages the tenant outside Planton, import rather than re-apply: the tenant imports by any string (Planton uses `tenant`), and its default domain by the fixed identifier `custom_domain_default`. Declare the settings the other tool managed, with their live values, and the first plan shows no change for them.

### What the Flags Do

`flags` holds the tenant's behavior switches; each is managed only when set, and one declared flag leaves the others as they are. The ones that change what people and clients see: `enableClientConnections` (every connection turned on for each new application -- Auth0 recommends `false`), `enableCustomDomainInEmails` (email links on the custom domain, once one is ready), `enableDynamicClientRegistration` (open registration, above), `enableSso` (skip the "continue as" confirmation; matters only on older tenants), `enablePublicSignupUserExistsError` and `noDiscloseEnterpriseConnections` (the dashboard shows both inverted), `revokeRefreshTokenGrant` and `useScopeDescriptionsForConsent`. The rest keep legacy behavior (above) or change the dashboard itself. The deprecated `require_pushed_authorization_requests` flag is not offered: Auth0 requires PAR per application.

### Clearing a Default

`defaultAudience` and `defaultDirectory` take a value or a reference, never an empty string, so removing a tenant's default audience or directory is done in the dashboard (Settings, General, API Authorization Settings). Removing the field only stops managing it.

## Plan Boundaries

Auth0's documentation gates these settings on a plan, an add-on or Early Access, and the spec comment on each says so:

| Setting | Needs |
|---|---|
| `pushedAuthorizationRequestsSupported` | The Enterprise plan with the Highly Regulated Identity add-on |
| `mtls` | The Enterprise plan with the Highly Regulated Identity add-on |
| `sessions.anonymous` | Early Access, on the Enterprise plan |
| `defaultTokenQuota` | Early Access Auth0 enables on request |
| `clientIdMetadataDocumentSupported` | Early Access, per Auth0's Management API |
| `dynamicClientRegistrationSecurityMode` | Only customers that had third-party applications before April 2026 |
| `sessionLifetime`, `idleSessionLifetime` | Longer limits on the Enterprise plan: Auth0 caps them at 720 and 72 hours elsewhere, silently |

## Permissions
## Management API Scopes

Auth0 tenant settings require the following Management API scopes. They must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:tenant_settings` | Read the tenant's current settings |
| Update | `update:tenant_settings` | Change the tenant's settings |
| Read default domain | `read:custom_domains` | Read the tenant's default domain (only when `defaultCustomDomain` is set) |
| Set default domain | `update:custom_domains` | Set the tenant's default domain (only when `defaultCustomDomain` is set) |

There is no create or delete scope to grant: a tenant can't be created or deleted through the Management API, and its settings have no delete.

## Compliance
## Tenant-Settings Compliance Notes

### Tenant Behavior as Auditable Configuration

The tenant's face (its name, logo and support contacts) and its behavior (session lifetimes, what its OAuth endpoints accept, its legacy switches) are version-controlled with the rest of its identity configuration, so a change to either is reviewed like any other change. Changes are also recorded in Auth0's tenant logs.

## Cost
## Pricing Model

Auth0 pricing is based on the plan tier and monthly active users, not on settings. The friendly name, logo and support contacts are editable on every plan, the Free plan included; the settings in Plan Boundaries need a higher plan, an add-on, or Early Access. Page customization beyond the logo is gated on a custom domain rather than a plan: Universal Login page templates need one (Auth0's Free plan includes one custom domain), and this component doesn't touch them.

## Cost Impact

Changing these settings has no billing impact.
