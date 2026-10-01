# Auth0TenantSettings

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0TenantSettingsSpec manages the settings of an existing Auth0 tenant:
how it presents itself to the people who sign in through it (its name, logo,
languages and support contacts), how long their sessions last, what its
OAuth and OpenID Connect endpoints accept (including the settings third-party
and MCP clients depend on), the defaults every application inherits, its
sign-in and MFA behavior, its error page, and its behavior flags.

The tenant is the one the provider connection's credential belongs to. Auth0's
Management API cannot create or delete a tenant, so this kind never does either:
it manages the settings of the tenant that already exists.

The fields are grouped the way Auth0's dashboard documents them:
- identity: friendly_name, picture_url, support_email, support_url,
  enabled_locales, sandbox_version
- sessions: session_lifetime, idle_session_lifetime,
  ephemeral_session_lifetime, idle_ephemeral_session_lifetime,
  session_cookie, sessions
- OAuth and OpenID Connect, for first-, third-party and MCP clients:
  client_id_metadata_document_supported, resource_parameter_profile,
  dynamic_client_registration_security_mode,
  pushed_authorization_requests_supported, acr_values_supported,
  disable_acr_values_supported,
  allow_organization_name_in_authentication_api, default_redirection_uri,
  allowed_logout_urls, oidc_logout, mtls,
  skip_non_verifiable_callback_uri_confirmation_prompt
- defaults: default_audience, default_directory, default_token_quota,
  default_custom_domain
- sign-in and MFA: customize_mfa_in_postlogin_action,
  phone_consolidated_experience, country_codes
- the error page: error_page
- behavior flags: flags

Each field maps to one tenant setting (PATCH /api/v2/tenants/settings). A field
left unset is NOT MANAGED: the module never sends it, and whatever value the
tenant already carries stays untouched. Setting a field manages that setting;
clearing it back to unset stops managing it but does NOT revert the live value.
Auth0 has no delete for tenant settings, so destroy abandons the last-applied
values. To return a setting to a specific value, set that value explicitly
before removing the field.

Six settings are the exception, because the provider cannot leave them alone
once this resource manages the tenant: default_redirection_uri,
skip_non_verifiable_callback_uri_confirmation_prompt, mtls, error_page,
default_token_quota and country_codes (and sessions.anonymous, when sessions
is declared). Their comments say what an unset value does on a tenant that
already carries one; declare the live values when adopting an existing
tenant.

The tenant's default custom domain (default_custom_domain) is managed the same
way as the other settings: unset, Auth0 keeps whatever default the tenant has.

The credential needs read:tenant_settings and update:tenant_settings on the
tenant's Management API, and read:custom_domains and update:custom_domains when
default_custom_domain is set (iac/permissions.yaml).

https://auth0.com/docs/get-started/tenant-settings
https://auth0.com/docs/customize/custom-domains/multiple-custom-domains/default-domain
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/tenant
https://www.pulumi.com/registry/packages/auth0/api-docs/tenant/

## Example

```yaml
# Auth0 Tenant Settings Test Manifest
# This file is used for testing the Auth0TenantSettings component. It exercises
# the whole spec, so the offline plan and preview proofs cover every argument
# and block the modules can send, including settings a live lane's tenant may
# not be entitled to (mtls and pushed authorization requests need the
# Enterprise plan with the Highly Regulated Identity add-on; anonymous
# sessions, token quotas and Client ID Metadata Document registration are
# Early Access).
#
# Applying it REWRITES the settings of the tenant the credential belongs to:
# run it only against a test tenant nobody signs in to.
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: the test tenant's domain (e.g., your-test-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - read:tenant_settings
#    - update:tenant_settings

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0TenantSettings
metadata:
  name: test-tenant-settings
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # Identity: "Log in to <friendlyName> to continue to <application>"
  friendlyName: Acme
  pictureUrl: https://assets.example.com/logo.png
  supportEmail: support@example.com
  supportUrl: https://example.com/support
  enabledLocales:
    - en
    - fr-CA
  sandboxVersion: "22"

  # Sessions, in hours
  sessionLifetime: 168
  idleSessionLifetime: 72
  ephemeralSessionLifetime: 72
  idleEphemeralSessionLifetime: 0.5
  sessionCookie:
    mode: persistent
  sessions:
    oidcLogoutPromptEnabled: true
    anonymous:
      activateCookie: true
      lifetimeInMinutes: 43200

  # OAuth and OpenID Connect
  clientIdMetadataDocumentSupported: true
  resourceParameterProfile: compatibility
  dynamicClientRegistrationSecurityMode: strict
  pushedAuthorizationRequestsSupported: true
  acrValuesSupported:
    - urn:mace:incommon:iap:silver
  allowOrganizationNameInAuthenticationApi: false
  defaultRedirectionUri: https://example.com/login
  allowedLogoutUrls:
    - https://example.com
  oidcLogout:
    rpLogoutEndSessionEndpointDiscovery: true
  mtls:
    enableEndpointAliases: true
  skipNonVerifiableCallbackUriConfirmationPrompt: false

  # Defaults
  defaultAudience:
    value: https://api.example.com
  defaultDirectory:
    value: Username-Password-Authentication
  defaultTokenQuota:
    clients:
      clientCredentials:
        enforce: true
        perDay: 10000
        perHour: 1000
    organizations:
      clientCredentials:
        enforce: false
        perDay: 50000

  # Sign-in and MFA
  customizeMfaInPostloginAction: true
  phoneConsolidatedExperience: false
  countryCodes:
    list:
      - US
      - CA
    mode: allow

  # The error page
  errorPage:
    url: https://example.com/error
    showLogLink: false

  # Behavior flags
  flags:
    enableClientConnections: false
    enableDynamicClientRegistration: false
    enablePublicSignupUserExistsError: false
    allowLegacyRoGrantTypes: false
    allowLegacyDelegationGrantTypes: false
    allowLegacyTokeninfoEndpoint: false
    revokeRefreshTokenGrant: false
    useScopeDescriptionsForConsent: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.friendlyName` | `string` |  |  |  |
| `spec.pictureUrl` | `string` |  |  |  |
| `spec.supportEmail` | `string` |  |  |  |
| `spec.supportUrl` | `string` |  |  |  |
| `spec.defaultCustomDomain` | `string \| valueFrom` |  |  | Auth0CustomDomainVerification (`status.outputs.domain`) |
| `spec.enabledLocales` | `[]string` |  |  |  |
| `spec.sandboxVersion` | `string` |  |  |  |
| `spec.sessionLifetime` | `double` |  |  |  |
| `spec.idleSessionLifetime` | `double` |  |  |  |
| `spec.ephemeralSessionLifetime` | `double` |  |  |  |
| `spec.idleEphemeralSessionLifetime` | `double` |  |  |  |
| `spec.sessionCookie` | `Auth0TenantSettingsSessionCookie` |  |  |  |
| `spec.sessionCookie.mode` | `string` |  |  |  |
| `spec.sessions` | `Auth0TenantSettingsSessions` |  |  |  |
| `spec.sessions.oidcLogoutPromptEnabled` | `bool` | yes |  |  |
| `spec.sessions.anonymous` | `Auth0TenantSettingsAnonymousSessions` |  |  |  |
| `spec.sessions.anonymous.activateCookie` | `bool` |  |  |  |
| `spec.sessions.anonymous.lifetimeInMinutes` | `int32` |  |  |  |
| `spec.clientIdMetadataDocumentSupported` | `bool` |  |  |  |
| `spec.resourceParameterProfile` | `string` |  |  |  |
| `spec.dynamicClientRegistrationSecurityMode` | `string` |  |  |  |
| `spec.pushedAuthorizationRequestsSupported` | `bool` |  |  |  |
| `spec.acrValuesSupported` | `[]string` |  |  |  |
| `spec.disableAcrValuesSupported` | `bool` |  |  |  |
| `spec.allowOrganizationNameInAuthenticationApi` | `bool` |  |  |  |
| `spec.defaultRedirectionUri` | `string` |  |  |  |
| `spec.allowedLogoutUrls` | `[]string` |  |  |  |
| `spec.oidcLogout` | `Auth0TenantSettingsOidcLogout` |  |  |  |
| `spec.oidcLogout.rpLogoutEndSessionEndpointDiscovery` | `bool` | yes |  |  |
| `spec.mtls` | `Auth0TenantSettingsMtls` |  |  |  |
| `spec.mtls.disable` | `bool` |  |  |  |
| `spec.mtls.enableEndpointAliases` | `bool` |  |  |  |
| `spec.skipNonVerifiableCallbackUriConfirmationPrompt` | `bool` |  |  |  |
| `spec.defaultAudience` | `string \| valueFrom` |  |  | Auth0ResourceServer (`status.outputs.identifier`) |
| `spec.defaultDirectory` | `string \| valueFrom` |  |  | Auth0Connection (`status.outputs.name`) |
| `spec.defaultTokenQuota` | `Auth0TenantSettingsDefaultTokenQuota` |  |  |  |
| `spec.defaultTokenQuota.clients` | `Auth0TenantSettingsTokenQuota` |  |  |  |
| `spec.defaultTokenQuota.clients.clientCredentials` | `Auth0TenantSettingsClientCredentialsQuota` | yes |  |  |
| `spec.defaultTokenQuota.clients.clientCredentials.enforce` | `bool` |  |  |  |
| `spec.defaultTokenQuota.clients.clientCredentials.perDay` | `int32` |  |  |  |
| `spec.defaultTokenQuota.clients.clientCredentials.perHour` | `int32` |  |  |  |
| `spec.defaultTokenQuota.organizations` | `Auth0TenantSettingsTokenQuota` |  |  |  |
| `spec.defaultTokenQuota.organizations.clientCredentials` | `Auth0TenantSettingsClientCredentialsQuota` | yes |  |  |
| `spec.defaultTokenQuota.organizations.clientCredentials.enforce` | `bool` |  |  |  |
| `spec.defaultTokenQuota.organizations.clientCredentials.perDay` | `int32` |  |  |  |
| `spec.defaultTokenQuota.organizations.clientCredentials.perHour` | `int32` |  |  |  |
| `spec.customizeMfaInPostloginAction` | `bool` |  |  |  |
| `spec.phoneConsolidatedExperience` | `bool` |  |  |  |
| `spec.countryCodes` | `Auth0TenantSettingsCountryCodes` |  |  |  |
| `spec.countryCodes.list` | `[]string` | yes |  |  |
| `spec.countryCodes.mode` | `string` | yes |  |  |
| `spec.errorPage` | `Auth0TenantSettingsErrorPage` |  |  |  |
| `spec.errorPage.html` | `string` |  |  |  |
| `spec.errorPage.showLogLink` | `bool` |  |  |  |
| `spec.errorPage.url` | `string` |  |  |  |
| `spec.flags` | `Auth0TenantSettingsFlags` |  |  |  |
| `spec.flags.allowLegacyDelegationGrantTypes` | `bool` |  |  |  |
| `spec.flags.allowLegacyRoGrantTypes` | `bool` |  |  |  |
| `spec.flags.allowLegacyTokeninfoEndpoint` | `bool` |  |  |  |
| `spec.flags.dashboardInsightsView` | `bool` |  |  |  |
| `spec.flags.dashboardLogStreamsNext` | `bool` |  |  |  |
| `spec.flags.disableClickjackProtectionHeaders` | `bool` |  |  |  |
| `spec.flags.disableFieldsMapFix` | `bool` |  |  |  |
| `spec.flags.disableManagementApiSmsObfuscation` | `bool` |  |  |  |
| `spec.flags.enableAdfsWaadEmailVerification` | `bool` |  |  |  |
| `spec.flags.enableApisSection` | `bool` |  |  |  |
| `spec.flags.enableClientConnections` | `bool` |  |  |  |
| `spec.flags.enableCustomDomainInEmails` | `bool` |  |  |  |
| `spec.flags.enableDynamicClientRegistration` | `bool` |  |  |  |
| `spec.flags.enableIdtokenApi2` | `bool` |  |  |  |
| `spec.flags.enableLegacyLogsSearchV2` | `bool` |  |  |  |
| `spec.flags.enableLegacyProfile` | `bool` |  |  |  |
| `spec.flags.enablePipeline2` | `bool` |  |  |  |
| `spec.flags.enablePublicSignupUserExistsError` | `bool` |  |  |  |
| `spec.flags.enableSso` | `bool` |  |  |  |
| `spec.flags.mfaShowFactorListOnEnrollment` | `bool` |  |  |  |
| `spec.flags.noDiscloseEnterpriseConnections` | `bool` |  |  |  |
| `spec.flags.removeAlgFromJwks` | `bool` |  |  |  |
| `spec.flags.revokeRefreshTokenGrant` | `bool` |  |  |  |
| `spec.flags.useScopeDescriptionsForConsent` | `bool` |  |  |  |

## Field Details

### spec.friendlyName

`string`

friendly_name is the tenant's name as people see it: Universal Login's
"Log in to <friendly_name> to continue to <application>", and the name in
the emails Auth0 sends for the tenant. Unset, Auth0 shows the tenant's
identifier (for example "acme-prod").

### spec.pictureUrl

`string`

picture_url is the URL of the logo shown for the tenant on its login
and consent pages (Auth0 recommends 150 x 150 pixels). Unset, Auth0's own
logo is shown.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uri":true}}

### spec.supportEmail

`string`

support_email is the address the tenant's login and error pages offer to a
person who needs help signing in.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"email":true}}

### spec.supportUrl

`string`

support_url is the page the tenant's login and error pages link to for
help.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uri":true}}

### spec.defaultCustomDomain

`string | valueFrom`

default_custom_domain is the domain that speaks for the tenant when a
request does not say which of its domains it came through: the links in the
emails Auth0 sends (verification, password reset, invitations) and the
notifications the Management API triggers. Set it to the tenant's custom
domain so a person who signs in at id.example.com also receives links to
id.example.com, never to the tenant's canonical auth0.com domain.

Only a verified domain can be the default, so reference the
Auth0CustomDomainVerification (its status.outputs.domain): the default is
then set only after Auth0 has verified the domain. A literal is also
accepted, including the tenant's canonical domain (e.g.
"example.eu.auth0.com") to make it the default again. Unset, the tenant's
default is not managed; clearing the field stops managing it and leaves the
last-applied default in place (Auth0 has no way to unset a default).

- references: Auth0CustomDomainVerification (`status.outputs.domain`)
- rule: write as {value: <literal>} or {valueFrom: {kind: Auth0CustomDomainVerification, name: <that resource's name>, fieldPath: status.outputs.domain}} -- a bare string does not parse

### spec.enabledLocales

`[]string`

enabled_locales are the languages the tenant's Universal Login pages and
emails are offered in, as Auth0 language codes ("en", "fr-CA", "pt-BR").
The FIRST entry is the tenant's default language: the one a person sees
when the browser asks for none of the others. Empty, the tenant's languages
are not managed.

- rule: {"repeated":{"items":{"string":{"in":["am","ar","ar-EG","ar-SA","az","bg","bn","bs","ca-ES","cnr","cs","cy","da","de","el","en","en-CA","es","es-419","es-AR","es-MX","et","eu-ES","fa","fi","fr","fr-CA","fr-FR","gl-ES","gu","he","hi","hr","hu","hy","id","is","it","ja","ka","kk","kn","ko","lt","lv","mk","ml","mn","mr","ms","my","nb","nl","nn","no","pa","pl","pt","pt-BR","pt-PT","ro","ru","sk","sl","so","sq","sr","sv","sw","ta","te","th","tl","tr","uk","ur","vi","zgh","zh-CN","zh-HK","zh-MO","zh-TW"]}}}}

### spec.sandboxVersion

`string` · optional (explicit presence)

sandbox_version is the Node.js runtime the tenant's extensibility code
runs on -- custom database action scripts and custom social connections
(e.g. "22"; the dashboard's Settings, Advanced, Extensibility, Runtime).
Moving it can break a script written for the older runtime: run the
dashboard's custom-database-script check first. Unset, the runtime is not
managed.

- rule: {"string":{"maxLen":"8"}}

### spec.sessionLifetime

`double` · optional (explicit presence)

session_lifetime is the number of hours a login session lasts however
active the person is ("Require log in after" in the dashboard): after it,
they sign in again. A value of an hour or more is a whole number of hours;
below an hour, Auth0 keeps it in minutes (0.5 is 30 minutes). Auth0 caps
it at 720 hours (30 days) on non-Enterprise plans and 8760 hours (365
days) on Enterprise plans, silently. Unset, the lifetime is not managed
(Auth0's own default is 168 hours).

- rule: a session_lifetime of an hour or more must be a whole number of hours (Auth0 stores hours as integers); use a value below 1 for a lifetime in minutes
- rule: {"double":{"gte":0.01}}

### spec.idleSessionLifetime

`double` · optional (explicit presence)

idle_session_lifetime is the number of hours a login session survives
without the person using it ("Inactivity timeout" in the dashboard). A
value of an hour or more is a whole number of hours; below an hour, Auth0
keeps it in minutes. Auth0 caps it at 72 hours (3 days) on non-Enterprise
plans and 2400 hours (100 days) on Enterprise plans, silently. Unset, the
idle lifetime is not managed (Auth0's own default is 72 hours).

- rule: an idle_session_lifetime of an hour or more must be a whole number of hours (Auth0 stores hours as integers); use a value below 1 for a lifetime in minutes
- rule: {"double":{"gte":0.01}}

### spec.ephemeralSessionLifetime

`double` · optional (explicit presence)

ephemeral_session_lifetime is session_lifetime for a non-persistent
session (session_cookie.mode "non-persistent", or a person who unticks
"remember me"): the hours it lasts however active the person is. A value of
an hour or more is a whole number of hours; below an hour, Auth0 keeps it
in minutes. Unset, it is not managed (Auth0's own default is 72 hours).

- rule: an ephemeral_session_lifetime of an hour or more must be a whole number of hours (Auth0 stores hours as integers); use a value below 1 for a lifetime in minutes
- rule: {"double":{"gte":0.0167}}

### spec.idleEphemeralSessionLifetime

`double` · optional (explicit presence)

idle_ephemeral_session_lifetime is idle_session_lifetime for a
non-persistent session: the hours it survives unused. A value of an hour or
more is a whole number of hours; below an hour, Auth0 keeps it in minutes.
Unset, it is not managed (Auth0's own default is 24 hours).

- rule: an idle_ephemeral_session_lifetime of an hour or more must be a whole number of hours (Auth0 stores hours as integers); use a value below 1 for a lifetime in minutes
- rule: {"double":{"gte":0.0167}}

### spec.sessionCookie

`Auth0TenantSettingsSessionCookie`

session_cookie chooses whether the login session outlives the browser.
Unset, the cookie's behavior is not managed.

### spec.sessionCookie.mode

`string` · optional (explicit presence)

mode is the session cookie's lifetime:
- "persistent": the session survives closing the browser, until its
  session_lifetime or idle_session_lifetime runs out.
- "non-persistent": closing the browser ends the session; the ephemeral
  lifetimes apply while it is open. Biometrics as a first factor does not
  work with non-persistent sessions.
Unset, the mode is not managed.

- rule: {"string":{"in":["persistent","non-persistent"]}}

### spec.sessions

`Auth0TenantSettingsSessions`

sessions holds the tenant's logout-confirmation and anonymous-session
settings. Unset, neither is managed.

### spec.sessions.oidcLogoutPromptEnabled

`bool` · required · optional (explicit presence)

oidc_logout_prompt_enabled decides whether Auth0 asks the person to
confirm an RP-initiated logout request that does not carry the hints
proving where it came from ("RP-Initiated Logout End-User Confirmation" in
the dashboard). true asks, which is Auth0's default; false logs out
without asking. Required when sessions is declared.

- rule: {"required":true}

### spec.sessions.anonymous

`Auth0TenantSettingsAnonymousSessions`

anonymous configures anonymous sessions: sessions and access tokens for
people who have not signed in yet (a shopping cart, a wishlist), carried
into their account when they do. Early Access, on the Enterprise plan.
When sessions is declared, leaving anonymous unset removes the tenant's
anonymous-session settings; declare the live value when adopting a tenant
that has them.

### spec.sessions.anonymous.activateCookie

`bool` · optional (explicit presence)

activate_cookie decides whether anonymous-session requests return the
auth0_anon cookie. Set false when you carry the session with a transfer
ticket instead. Unset, Auth0 issues the cookie.

### spec.sessions.anonymous.lifetimeInMinutes

`int32` · optional (explicit presence)

lifetime_in_minutes is how long an anonymous session lasts, from 1 minute
to 525600 (a year). Unset, Auth0 applies 43200 (30 days).

- rule: {"int32":{"lte":525600,"gte":1}}

### spec.clientIdMetadataDocumentSupported

`bool` · optional (explicit presence)

client_id_metadata_document_supported lets the tenant register an
application from a Client ID Metadata Document (CIMD): a JSON file the
application hosts on its own HTTPS domain, whose URL is its client ID --
the registration an MCP client that has never met the tenant can use,
without a shared secret. A CIMD client is always a third-party application
in strict mode, and its logins fail while the tenant has active Rules.
Auth0's Management API marks the setting Early Access. Unset, it is not
managed (Auth0's own default is false).

### spec.resourceParameterProfile

`string` · optional (explicit presence)

resource_parameter_profile chooses how a client names the API it wants a
token for:
- "audience": only the audience parameter; a resource parameter is passed
  on to the upstream identity provider untouched.
- "compatibility": audience first, and when a request has none, the RFC
  8707 resource parameter names the API -- what MCP clients send. Auth0
  consumes the parameter and does not pass it upstream.
Unset, the profile is not managed.

- rule: {"string":{"in":["audience","compatibility"]}}

### spec.dynamicClientRegistrationSecurityMode

`string` · optional (explicit presence)

dynamic_client_registration_security_mode is the security mode Auth0
gives each application a client registers for itself through Dynamic
Client Registration (flags.enable_dynamic_client_registration):
- "strict": the enhanced security controls of a third-party application
  (PKCE required, only authorization_code and refresh_token grants, only
  domain-level connections).
- "permissive": the behavior third-party applications had before those
  controls.
Auth0 lets only customers that had third-party applications before April
2026 configure this setting; other tenants' registrations are always
strict. Unset, it is not managed.

- rule: {"string":{"in":["strict","permissive"]}}

### spec.pushedAuthorizationRequestsSupported

`bool` · optional (explicit presence)

pushed_authorization_requests_supported opens the tenant's /oauth/par
endpoint, so an application can push its authorization request over the
back channel instead of the browser's address bar (RFC 9126). An
application can then be set to require it. Needs the Enterprise plan with
the Highly Regulated Identity add-on. Unset, it is not managed.

### spec.acrValuesSupported

`[]string`

acr_values_supported are the Authentication Context Class Reference values
the tenant accepts, published in its OpenID configuration; an
authentication request naming any other value is refused (e.g.
"urn:mace:incommon:iap:silver"). Empty, the list is not managed; to remove
it, set disable_acr_values_supported instead.

### spec.disableAcrValuesSupported

`bool` · optional (explicit presence)

disable_acr_values_supported, set true, removes the tenant's list of
accepted ACR values (acr_values_supported must then be empty). Unset, it
is not managed.

### spec.allowOrganizationNameInAuthenticationApi

`bool` · optional (explicit presence)

allow_organization_name_in_authentication_api lets /authorize and the
SAML endpoints accept an Auth0 organization's name as well as its ID, and
adds an org_name claim beside org_id to ID and access tokens. A name can
be renamed and reused where an ID cannot, so an application that
authorizes on the org_name claim must check it with that in mind. Unset,
it is not managed.

### spec.defaultRedirectionUri

`string` · optional (explicit presence)

default_redirection_uri is the tenant's login route ("Tenant Login URI"
in the dashboard): the HTTPS address of a page in your application that
starts a login, which Auth0 redirects to when it must start one itself --
a person opening a bookmarked login page or an expired password-reset link
-- and the application has no login route of its own. An empty string
removes it.

Declare the live value when adopting an existing tenant: the provider has
no computed value for it, so leaving it unset on a tenant that has one
plans a change to remove it on every run (Auth0 keeps the value, as
nothing is sent).

- rule: default_redirection_uri must be an absolute https:// URL (or an empty string to remove it)

### spec.allowedLogoutUrls

`[]string`

allowed_logout_urls are the addresses Auth0 may send a person to after
logout when the logout request names no application (with SSO, the one
list every application shares). A URL may carry wildcards the way
application URLs do. Empty, the list is not managed.

### spec.oidcLogout

`Auth0TenantSettingsOidcLogout`

oidc_logout holds the tenant's RP-initiated logout discovery setting.
Unset, it is not managed.

### spec.oidcLogout.rpLogoutEndSessionEndpointDiscovery

`bool` · required · optional (explicit presence)

rp_logout_end_session_endpoint_discovery decides whether the tenant's
OpenID configuration advertises its logout endpoint as
end_session_endpoint, so an OpenID Connect library finds it on its own.
Required when oidc_logout is declared.

- rule: {"required":true}

### spec.mtls

`Auth0TenantSettingsMtls`

mtls configures the tenant's mutual-TLS endpoints, for applications that
authenticate with a client certificate or bind their tokens to one. Needs
the Enterprise plan with the Highly Regulated Identity add-on.

Declare the live value when adopting an existing tenant: the first deploy
of this resource sends an unset mtls as "no mTLS configuration", which
removes the tenant's mTLS endpoint aliases. After that first deploy, an
unset mtls is left alone.

- rule: set either mtls.disable: true (remove the tenant's mTLS configuration) or mtls.enable_endpoint_aliases: true, not both

### spec.mtls.disable

`bool` · optional (explicit presence)

disable, set true, removes the tenant's mTLS configuration.

### spec.mtls.enableEndpointAliases

`bool` · optional (explicit presence)

enable_endpoint_aliases publishes the tenant's mTLS endpoints under their
own aliases in its OpenID configuration (mtls_endpoint_aliases), so a
client that authenticates with a certificate reaches the endpoints that
ask for one.

### spec.skipNonVerifiableCallbackUriConfirmationPrompt

`bool` · optional (explicit presence)

skip_non_verifiable_callback_uri_confirmation_prompt decides whether Auth0
asks the person to confirm a login whose callback it cannot verify -- a
custom scheme such as myapp:// or a localhost address, where another app
on the device could be posing as yours. true skips the prompt; false shows
it, which Auth0 recommends.

Declare the live value when adopting an existing tenant; leaving it unset
on an adopted tenant resets it (to Auth0's unset state).

### spec.defaultAudience

`string | valueFrom`

default_audience is the API every authorization request is treated as
asking for when it names none: every access token the tenant issues then
carries this API's identifier as its audience. Setting it changes the
tokens every application receives, and can break one that expects an
opaque token. Reference the Auth0ResourceServer (its
status.outputs.identifier), or give the identifier as a literal. Unset, it
is not managed. This field cannot carry an empty value, so a default
audience is removed in the dashboard (Settings, General, API Authorization
Settings).

- references: Auth0ResourceServer (`status.outputs.identifier`)
- rule: write as {value: <literal>} or {valueFrom: {kind: Auth0ResourceServer, name: <that resource's name>, fieldPath: status.outputs.identifier}} -- a bare string does not parse

### spec.defaultDirectory

`string | valueFrom`

default_directory is the connection the Resource Owner Password flow (the
/oauth/token password grant) signs people in through when the request
names no realm, and the default connection of Universal Login. It must be
a database, passwordless (email or sms), AD/LDAP, Azure AD or ADFS
connection. Reference the Auth0Connection (its status.outputs.name), or
give the connection's name as a literal. Unset, it is not managed. This
field cannot carry an empty value, so a default directory is removed in
the dashboard.

- references: Auth0Connection (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: Auth0Connection, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.defaultTokenQuota

`Auth0TenantSettingsDefaultTokenQuota`

default_token_quota caps how many tokens the client-credentials grant
issues, per application and per organization, unless an application or
organization sets its own quota. Auth0 offers token quotas as Early Access
it enables on request.

Declare the live value when adopting an existing tenant; leaving it unset
on an adopted tenant resets it (removes the tenant's default quotas). A
declared quota with no clients or organizations limit also removes them.

### spec.defaultTokenQuota.clients

`Auth0TenantSettingsTokenQuota`

clients is the quota every application gets unless it sets its own.

### spec.defaultTokenQuota.clients.clientCredentials

`Auth0TenantSettingsClientCredentialsQuota` · required

client_credentials is the quota on tokens from the client-credentials
grant.

- rule: {"required":true}

### spec.defaultTokenQuota.clients.clientCredentials.enforce

`bool` · optional (explicit presence)

enforce decides what a request over the quota gets: true refuses it;
false issues the token and records the overrun in the tenant's logs.
Unset, the quota is enforced.

### spec.defaultTokenQuota.clients.clientCredentials.perDay

`int32` · optional (explicit presence)

per_day is the most tokens issued in a day. Unset, no daily cap.

- rule: {"int32":{"gte":1}}

### spec.defaultTokenQuota.clients.clientCredentials.perHour

`int32` · optional (explicit presence)

per_hour is the most tokens issued in an hour. Unset, no hourly cap.

- rule: {"int32":{"gte":1}}

### spec.defaultTokenQuota.organizations

`Auth0TenantSettingsTokenQuota`

organizations is the quota every organization gets unless it sets its
own.

### spec.defaultTokenQuota.organizations.clientCredentials

`Auth0TenantSettingsClientCredentialsQuota` · required

client_credentials is the quota on tokens from the client-credentials
grant.

- rule: {"required":true}

### spec.defaultTokenQuota.organizations.clientCredentials.enforce

`bool` · optional (explicit presence)

enforce decides what a request over the quota gets: true refuses it;
false issues the token and records the overrun in the tenant's logs.
Unset, the quota is enforced.

### spec.defaultTokenQuota.organizations.clientCredentials.perDay

`int32` · optional (explicit presence)

per_day is the most tokens issued in a day. Unset, no daily cap.

- rule: {"int32":{"gte":1}}

### spec.defaultTokenQuota.organizations.clientCredentials.perHour

`int32` · optional (explicit presence)

per_hour is the most tokens issued in an hour. Unset, no hourly cap.

- rule: {"int32":{"gte":1}}

### spec.customizeMfaInPostloginAction

`bool` · optional (explicit presence)

customize_mfa_in_postlogin_action lets a post-login Action choose which
MFA factor, or set of factors, a person is challenged with
(api.authentication.challengeWith and challengeWithAny), instead of the
tenant's MFA policy deciding. Universal Login only. Unset, it is not
managed.

### spec.phoneConsolidatedExperience

`bool` · optional (explicit presence)

phone_consolidated_experience routes every MFA and passwordless text and
voice message through the tenant's one phone provider (the Unified Phone
Experience), instead of Guardian's own phone configuration. Configure the
tenant's phone provider before turning it on. Unset, it is not managed.

### spec.countryCodes

`Auth0TenantSettingsCountryCodes`

country_codes limits the phone numbers people can sign in with, by
country calling code.

Declare the live value when adopting an existing tenant; leaving it unset
on an adopted tenant resets it (removes the tenant's country filter). The
provider notes the tenant needs Auth0's country-codes feature enabled.

### spec.countryCodes.list

`[]string` · required

list is the countries, as uppercase ISO 3166-1 alpha-2 codes ("US", "GB").

- rule: {"repeated":{"minItems":"1","items":{"string":{"pattern":"^[A-Z]{2}$"}}}}

### spec.countryCodes.mode

`string` · required

mode decides what the list means: "allow" accepts only numbers from the
listed countries; "deny" refuses numbers from them.

- rule: {"required":true,"string":{"in":["allow","deny"]}}

### spec.errorPage

`Auth0TenantSettingsErrorPage`

error_page replaces the page Auth0 shows when a login fails before it can
send the person back to an application (a missing parameter, an expired
link, an invalid callback URL).

Declare the live value when adopting an existing tenant; leaving it unset
on an adopted tenant resets it (back to Auth0's default error page). A
declared error_page with no html, url or show_log_link: true also returns
the tenant to Auth0's default page.

- rule: error_page.url must be an absolute URL (or empty)

### spec.errorPage.html

`string` · optional (explicit presence)

html is a page Auth0 renders itself, written in Liquid: it can show the
error with {{error}}, {{error_description}}, {{tracking}}, {{client_id}}
and {{connection}} (escape each with | escape).

### spec.errorPage.showLogLink

`bool` · optional (explicit presence)

show_log_link decides whether Auth0's default error page links to the
failure in the tenant's logs.

### spec.errorPage.url

`string` · optional (explicit presence)

url is a page of your own Auth0 redirects to instead, with the error in
its query string (error, error_description, tracking, client_id,
connection, lang).

### spec.flags

`Auth0TenantSettingsFlags`

flags are the tenant's behavior switches, each managed only when set.
Unset, none is managed.

### spec.flags.allowLegacyDelegationGrantTypes

`bool` · optional (explicit presence)

allow_legacy_delegation_grant_types lets applications use the legacy
/delegation endpoint. Keep it false unless an old integration needs it.

### spec.flags.allowLegacyRoGrantTypes

`bool` · optional (explicit presence)

allow_legacy_ro_grant_types lets applications use the legacy /oauth/ro
endpoint (resource owner and legacy passwordless). Keep it false unless an
old integration needs it.

### spec.flags.allowLegacyTokeninfoEndpoint

`bool` · optional (explicit presence)

allow_legacy_tokeninfo_endpoint keeps the legacy /tokeninfo endpoint
answering.

### spec.flags.dashboardInsightsView

`bool` · optional (explicit presence)

dashboard_insights_view turns on the dashboard's newer insights activity
page.

### spec.flags.dashboardLogStreamsNext

`bool` · optional (explicit presence)

dashboard_log_streams_next gives the dashboard beta access to log
streaming changes.

### spec.flags.disableClickjackProtectionHeaders

`bool` · optional (explicit presence)

disable_clickjack_protection_headers, set true, stops classic Universal
Login prompts sending the headers that keep them from being framed by
another site. Leave it false unless a classic page must be embedded.

### spec.flags.disableFieldsMapFix

`bool` · optional (explicit presence)

disable_fields_map_fix turns off Auth0's correction of SAML field
mappings that repeat an attribute.

### spec.flags.disableManagementApiSmsObfuscation

`bool` · optional (explicit presence)

disable_management_api_sms_obfuscation, true by Auth0's default, returns
SMS phone numbers in full from Management API reads; false masks them.

### spec.flags.enableAdfsWaadEmailVerification

`bool` · optional (explicit presence)

enable_adfs_waad_email_verification asks people signing in through an
Azure AD or ADFS connection to verify their email on their first login.

### spec.flags.enableApisSection

`bool` · optional (explicit presence)

enable_apis_section shows the APIs section of the tenant's dashboard.

### spec.flags.enableClientConnections

`bool` · optional (explicit presence)

enable_client_connections, true by Auth0's default, turns on every
existing connection for each new application. Auth0 recommends false, so
each application signs in only through the connections you enable for it.

### spec.flags.enableCustomDomainInEmails

`bool` · optional (explicit presence)

enable_custom_domain_in_emails makes the emails Auth0 sends link to the
tenant's custom domain. The tenant needs a custom domain whose status is
ready first (Auth0CustomDomainVerification); default_custom_domain picks
which one.

### spec.flags.enableDynamicClientRegistration

`bool` · optional (explicit presence)

enable_dynamic_client_registration opens the tenant's /oidc/register
endpoint: any client -- an MCP client meeting your API for the first time,
or anyone else -- can register a third-party application WITHOUT a token.
Each registration is a third-party application
(dynamic_client_registration_security_mode) that reaches only the APIs and
scopes the tenant's default third-party permissions grant and signs in
only through domain-level connections, so configure those before turning
it on.

### spec.flags.enableIdtokenApi2

`bool` · optional (explicit presence)

enable_idtoken_api2 lets an ID token authorize some Management API v2
requests (legacy behavior).

### spec.flags.enableLegacyLogsSearchV2

`bool` · optional (explicit presence)

enable_legacy_logs_search_v2 uses the older v2 search of the tenant's
logs.

### spec.flags.enableLegacyProfile

`bool` · optional (explicit presence)

enable_legacy_profile makes ID tokens and /userinfo carry the whole user
profile instead of only OpenID Connect claims (legacy behavior).

### spec.flags.enablePipeline2

`bool` · optional (explicit presence)

enable_pipeline2, true by Auth0's default, turns on Auth0's advanced API
authorization scenarios.

### spec.flags.enablePublicSignupUserExistsError

`bool` · optional (explicit presence)

enable_public_signup_user_exists_error makes the public signup API answer
user_exists when an identifier is taken -- which lets anyone test whether
an address has an account. false answers generically. The dashboard's "Use
a generic response in public signup API error message" shows the opposite
of this value.

### spec.flags.enableSso

`bool` · optional (explicit presence)

enable_sso, true, sends a person with a live session straight on to the
application without asking them to confirm the login. It matters only on
older tenants: newer tenants always behave as true. The provider sends it
only when it changes.

### spec.flags.mfaShowFactorListOnEnrollment

`bool` · optional (explicit presence)

mfa_show_factor_list_on_enrollment lets a person enrolling in MFA choose
among the tenant's factors.

### spec.flags.noDiscloseEnterpriseConnections

`bool` · optional (explicit presence)

no_disclose_enterprise_connections, true, stops the tenant publishing its
enterprise connections and their identity-provider domains in the file
Lock reads -- which Home Realm Discovery in Lock relies on. The dashboard's
"Enable Publishing of Enterprise Connections Information with IdP domains"
shows the opposite of this value.

### spec.flags.removeAlgFromJwks

`bool` · optional (explicit presence)

remove_alg_from_jwks leaves the alg property out of the keys the tenant
publishes at /.well-known/jwks.json.

### spec.flags.revokeRefreshTokenGrant

`bool` · optional (explicit presence)

revoke_refresh_token_grant makes revoking a refresh token through
/oauth/revoke also delete the grant behind it (so the person consents
again).

### spec.flags.useScopeDescriptionsForConsent

`bool` · optional (explicit presence)

use_scope_descriptions_for_consent shows each scope's description, not its
name, on the consent page.

## Validation Rules

- `spec.at_least_one_setting`: configure at least one tenant setting -- an Auth0TenantSettings resource that manages nothing would deploy nothing
- `spec.acr_values_or_disable`: set either acr_values_supported (the ACR values the tenant accepts) or disable_acr_values_supported: true (no list at all), not both -- Auth0 cannot both keep and remove the list

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0TenantSettings, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.friendly_name` | `string` | friendly_name is the tenant's name as people see it. |
| `status.outputs.picture_url` | `string` | picture_url is the URL of the tenant's logo. |
| `status.outputs.support_email` | `string` | support_email is the support address the tenant's pages offer. |
| `status.outputs.support_url` | `string` | support_url is the support page the tenant's pages link to. |
| `status.outputs.default_custom_domain` | `string` | default_custom_domain is the tenant's default domain as set by this resource; empty when the spec leaves the default unmanaged. |
| `status.outputs.default_audience` | `string` | default_audience is the API identifier every access token defaults to; empty when the tenant has none. |
| `status.outputs.default_directory` | `string` | default_directory is the connection the password grant signs people in through by default; empty when the tenant has none. |
| `status.outputs.client_id_metadata_document_supported` | `bool` | client_id_metadata_document_supported is whether the tenant registers applications from a Client ID Metadata Document. |
| `status.outputs.resource_parameter_profile` | `string` | resource_parameter_profile is how a client names the API it wants a token for: "audience" or "compatibility". |
| `status.outputs.enable_dynamic_client_registration` | `bool` | enable_dynamic_client_registration is whether any client can register a third-party application through the tenant's /oidc/register endpoint. |
| `status.outputs.dynamic_client_registration_security_mode` | `string` | dynamic_client_registration_security_mode is the security mode of the applications Dynamic Client Registration creates; empty when the tenant reports none. |
| `status.outputs.session_lifetime` | `double` | session_lifetime is the hours a login session lasts however active the person is. |
| `status.outputs.idle_session_lifetime` | `double` | idle_session_lifetime is the hours a login session survives unused. |
| `status.outputs.session_cookie_mode` | `string` | session_cookie_mode is whether the session outlives the browser: "persistent" or "non-persistent". |
| `status.outputs.enabled_locales` | `[]string` | enabled_locales are the tenant's languages, its default first. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.defaultCustomDomain` | Auth0CustomDomainVerification | `status.outputs.domain` |
| `spec.defaultAudience` | Auth0ResourceServer | `status.outputs.identifier` |
| `spec.defaultDirectory` | Auth0Connection | `status.outputs.name` |

## See Also

- [Overview](../README.md)
