# Auth0TenantSettings

Manages the settings of an existing [Auth0 tenant](https://auth0.com/docs/get-started/tenant-settings): how it presents itself to the people who sign in through it (the name Universal Login shows, the logo, its languages and support contacts), how long their sessions last, what its OAuth and OpenID Connect endpoints accept (including what third-party and MCP clients rely on), the defaults every application inherits, its sign-in and MFA behavior, its error page, its behavior flags, and the domain its emails link to.

## When to Use

- **Your product's name on the login page**: Universal Login reads "Log in to *friendly name* to continue to *application*". Without a friendly name it shows the tenant's identifier (for example `acme-prod`).
- **Your logo instead of Auth0's**: the login and consent pages show the tenant's picture.
- **A way to get help**: people who can't sign in see your support address and page.
- **Sessions that fit the application**: how long a session lasts, how long it survives unused, and whether it outlives the browser.
- **An API MCP clients can reach**: Client ID Metadata Document registration, the RFC 8707 `resource` parameter, and Dynamic Client Registration.
- **Email links on your domain**: once a custom domain is verified, make it the tenant's default so verification and password-reset links point there, never at `auth0.com`.
- **Infrastructure as code**: the tenant's face and behavior are version-controlled beside the clients and connections it serves.

## Quick Start

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0TenantSettings
metadata:
  name: tenant-settings
  org: acme-corp
  env: production
spec:
  friendlyName: Acme
  pictureUrl: https://assets.acme.com/logo.png
  supportEmail: support@acme.com
  supportUrl: https://acme.com/support
  idleSessionLifetime: 8
  sessionCookie:
    mode: persistent
```

## Key Behaviors

- **The tenant is the credential's**: the settings managed are those of the tenant the provider connection's credential belongs to. One resource per tenant.
- **Unset is unmanaged**: a field left out is never sent, and the tenant keeps whatever value it already carries. At least one field must be set.
- **Six exceptions on adoption**: the provider cannot leave `defaultRedirectionUri`, `skipNonVerifiableCallbackUriConfirmationPrompt`, `mtls`, `errorPage`, `defaultTokenQuota` and `countryCodes` alone once this resource manages the tenant. Declare their live values before the first deploy (the guide's Adoption section).
- **Nothing is ever deleted**: Auth0's Management API cannot create or delete a tenant, and has no delete for its settings. Destroying this resource forgets the settings and leaves their last-applied values in place. To return a setting to a specific value, set that value before removing the field.
- **The default domain**: `defaultCustomDomain` reads an `Auth0CustomDomainVerification`'s `domain`, so only a verified domain becomes the default. The tenant's canonical domain is also accepted, to make it the default again. Auth0 has no way to unset a default, so destroy leaves the last one in place.
- **Defaults by reference**: `defaultAudience` reads an `Auth0ResourceServer`'s `identifier`, and `defaultDirectory` an `Auth0Connection`'s `name`.
- **Plan boundaries**: pushed authorization requests and mTLS need the Enterprise plan with the Highly Regulated Identity add-on; anonymous sessions, token quotas and Client ID Metadata Document registration are Early Access.
- **The application's half of the sentence** is each client's name (the `Auth0Client` kind's `name`, defaulting to its `metadata.name`).
- **Permissions**: the credential needs `read:tenant_settings` and `update:tenant_settings` on the tenant's Management API, and `read:custom_domains` and `update:custom_domains` when `defaultCustomDomain` is set (`iac/permissions.yaml`).

## Outputs

| Output | Description |
|---|---|
| `friendly_name` | The tenant's name as people see it |
| `picture_url` | The URL of the tenant's logo |
| `support_email` | The support address the tenant's pages offer |
| `support_url` | The support page the tenant's pages link to |
| `default_custom_domain` | The tenant's default domain as set by this resource; empty when unmanaged |
| `default_audience` | The API identifier every access token defaults to |
| `default_directory` | The connection the password grant signs people in through by default |
| `client_id_metadata_document_supported` | Whether the tenant registers applications from a Client ID Metadata Document |
| `resource_parameter_profile` | `audience` or `compatibility` |
| `enable_dynamic_client_registration` | Whether any client can register an application through `/oidc/register` |
| `dynamic_client_registration_security_mode` | The security mode of dynamically registered applications |
| `session_lifetime` | Hours a session lasts however active |
| `idle_session_lifetime` | Hours a session survives unused |
| `session_cookie_mode` | `persistent` or `non-persistent` |
| `enabled_locales` | The tenant's languages, its default first |

## Auth0 Documentation

- [Tenant settings](https://auth0.com/docs/get-started/tenant-settings)
- [Sessions](https://auth0.com/docs/manage-users/sessions)
- [Dynamic Client Registration](https://auth0.com/docs/get-started/applications/dynamic-client-registration)
- [Register applications with CIMD](https://auth0.com/docs/get-started/auth0-overview/create-applications/register-applications-with-cimd)
- [Terraform auth0_tenant](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/tenant)
- [Pulumi auth0.Tenant](https://www.pulumi.com/registry/packages/auth0/api-docs/tenant/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
