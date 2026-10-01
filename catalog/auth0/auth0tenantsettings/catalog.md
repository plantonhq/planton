# Auth0 Tenant Settings

Sets how an existing Auth0 tenant presents itself and behaves: the name Universal Login shows, the logo on its login and consent pages, its languages and support contacts, how long sessions last, what its OAuth endpoints accept (including what MCP clients rely on), the defaults every application inherits, its error page and its behavior flags. One Cloud Resource per tenant.

## What Gets Created

Nothing new: a tenant can't be created or deleted through Auth0's Management API. When you deploy this Cloud Resource, the IaC module sets the settings of the tenant your Auth0 connection's credential belongs to:

- **Friendly name** -- the name in "Log in to *friendly name* to continue to *application*", and in the emails Auth0 sends for the tenant
- **Logo** -- shown on the login and consent pages instead of Auth0's
- **Support email and page** -- offered to people who can't sign in
- **Languages** -- the languages its pages and emails are offered in, the default first
- **Sessions** -- how long a login lasts, how long it survives unused, and whether it outlives the browser
- **OAuth and OpenID Connect** -- Client ID Metadata Document registration, the `resource` parameter, pushed authorization requests, accepted ACR values, the login route, logout URLs and logout discovery, mTLS
- **Defaults** -- the API every token is for and the connection the password grant uses, when a request names none, and default token quotas
- **Sign-in and MFA** -- MFA factor choice in Actions, the unified phone provider, a phone country filter
- **Error page** -- your own page, or Auth0's with or without a log link
- **Flags** -- Dynamic Client Registration, application connections, legacy grant types, and the tenant's other switches
- **Default domain** -- the domain the tenant's emails link to, when you set one

Only what you set is changed; every other setting keeps its value.

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with the tenant's domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **The connection's Machine-to-Machine application** must hold `read:tenant_settings` and `update:tenant_settings` on the tenant's Management API (Auth0 dashboard: Applications, APIs, Auth0 Management API, Machine To Machine Applications).
- **A public HTTPS URL for your logo**, if you set one (about 150 x 150 pixels).

## Deploy

### Console

Open the deployment store, find **Auth0 Tenant Settings**, and click **Deploy**. Start from the **Product-Branded Login** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0TenantSettings
metadata:
  name: tenant-settings
  org: acme-corp
  env: prod
spec:
  friendlyName: Acme
  pictureUrl: https://assets.acme.com/logo.png
  supportEmail: support@acme.com
  supportUrl: https://acme.com/support
```

```shell
planton apply -f auth0-tenant-settings.yaml
```

The tenant's login page reads "Log in to Acme to continue to *application*" with your logo. A Stack Job tracks the change in real time.

### InfraChart

When the tenant's settings deploy beside the API, the connection and the custom domain they name, wire each by ValueFromRef so the pipeline applies the settings after what they point at:

```yaml
spec:
  friendlyName: Acme
  defaultAudience:
    valueFrom:
      kind: Auth0ResourceServer
      name: backend-api
      fieldPath: status.outputs.identifier
  defaultDirectory:
    valueFrom:
      kind: Auth0Connection
      name: users
      fieldPath: status.outputs.name
  defaultCustomDomain:
    valueFrom:
      kind: Auth0CustomDomainVerification
      name: sign-in-domain-verification
      fieldPath: status.outputs.domain
```

The InfraPipeline resolves the graph, deploys the API, the connection and the verified domain first, then applies the tenant's settings with the resolved values.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Unset is unmanaged** -- Every field is optional; a field you leave out is never sent, and the tenant keeps its current value. Set at least one.

**Declare what the provider resets** -- Six settings are the exception: on the first deploy the provider clears the tenant's error page, default token quotas, phone country filter, non-verifiable callback prompt and mTLS aliases when you leave them unset, and keeps planning to remove a login route (`defaultRedirectionUri`) you leave unset. Read the tenant's live values and declare them first.

**Sessions** -- `idleSessionLifetime` (hours unused) and `sessionLifetime` (hours in all) decide how often people sign in again; `sessionCookie.mode` whether closing the browser ends the session. Auth0 caps both lifetimes lower outside the Enterprise plan.

**Opening the tenant to MCP clients** -- `clientIdMetadataDocumentSupported`, `resourceParameterProfile: compatibility` and `flags.enableDynamicClientRegistration` let an MCP client register itself and ask for a token for your API. Dynamic registration is open to anyone: configure the tenant's default third-party permissions first.

**The other half of the sentence** -- Universal Login names the application after the Auth0 client. Give each client a people-facing name with the Auth0 Client kind's `name`, so the page reads "Log in to Acme to continue to Acme Console".

**Email links on your domain** -- Set `defaultCustomDomain` to a verified custom domain (reference the Auth0 Custom Domain Verification's `domain` output), so verification and password-reset emails link to your domain. The credential then also needs `read:custom_domains` and `update:custom_domains`.

**Destroy leaves the settings in place** -- Auth0 has no delete for tenant settings, so destroying this Cloud Resource stops managing them and keeps their last-applied values. To return a setting to a specific value, set that value before removing the field or the resource.

## Outputs and Dependencies

### What This Component Consumes

| Field | Foreign Key | Required |
|-------|-------------|----------|
| `defaultCustomDomain` | Auth0 Custom Domain Verification (`status.outputs.domain`) | No |
| `defaultAudience` | Auth0 Resource Server (`status.outputs.identifier`) | No |
| `defaultDirectory` | Auth0 Connection (`status.outputs.name`) | No |

The tenant is the one the Auth0 connection's credential belongs to.

### What This Component Provides

After provisioning, `status.outputs` contains the settings as the tenant carries them:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `friendly_name` | The tenant's name as people see it | Audits, documentation |
| `picture_url` | The URL of the tenant's logo | Audits |
| `support_email` | The support address the tenant's pages offer | Support runbooks |
| `support_url` | The support page the tenant's pages link to | Support runbooks |
| `default_custom_domain` | The tenant's default domain, when set here | Audits |
| `default_audience` | The API every access token defaults to | Audits |
| `default_directory` | The connection the password grant uses by default | Audits |
| `client_id_metadata_document_supported` | Whether CIMD registration is on | MCP readiness checks |
| `resource_parameter_profile` | `audience` or `compatibility` | MCP readiness checks |
| `enable_dynamic_client_registration` | Whether open registration is on | MCP readiness checks, security reviews |
| `dynamic_client_registration_security_mode` | The mode of registered applications | Security reviews |
| `session_lifetime` | Hours a session lasts in all | Security reviews |
| `idle_session_lifetime` | Hours a session survives unused | Security reviews |
| `session_cookie_mode` | Whether the session outlives the browser | Security reviews |
| `enabled_locales` | The tenant's languages, default first | Localization |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Product-branded login** -- The tenant introduces itself as your product, with your logo and your support contacts. Start from the **Product-Branded Login** preset.

**MCP-ready tenant** -- MCP clients register themselves and ask for tokens with the `resource` parameter. Start from the **MCP-Ready Tenant** preset.

**Short idle sessions** -- A session left unused for an hour ends; every session ends after a day. Start from the **Short Idle Sessions** preset.

## Works With

- [**Auth0 Client (Application)**](/cloud-catalog/auth0-client) -- names the application in the login page's sentence.
- [**Auth0 Connection**](/cloud-catalog/auth0-connection) -- the sign-in methods the branded page offers, and the tenant's default directory.
- [**Auth0 Resource Server**](/cloud-catalog/auth0-resource-server) -- the API the tenant's default audience names.
- [**Auth0 Custom Domain Verification**](/cloud-catalog/auth0-custom-domain-verification) -- the verified domain the tenant's emails link to.
