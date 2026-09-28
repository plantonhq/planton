# Auth0 Tenant Settings

Sets how an existing Auth0 tenant presents itself to the people who sign in through it: the name Universal Login shows, the logo on its login and consent pages, and the support contacts it offers. One Cloud Resource per tenant.

## What Gets Created

Nothing new: a tenant can't be created or deleted through Auth0's Management API. When you deploy this Cloud Resource, the IaC module sets the settings of the tenant your Auth0 connection's credential belongs to:

- **Friendly name** -- the name in "Log in to *friendly name* to continue to *application*", and in the emails Auth0 sends for the tenant
- **Logo** -- shown on the login and consent pages instead of Auth0's
- **Support email and page** -- offered to people who can't sign in
- **Default domain** -- the domain the tenant's emails link to, when you set one

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

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Unset is unmanaged** -- Every field is optional; a field you leave out is never sent, and the tenant keeps its current value. Set at least one.

**The other half of the sentence** -- Universal Login names the application after the Auth0 client. Give each client a people-facing name with the Auth0 Client kind's `name`, so the page reads "Log in to Acme to continue to Acme Console".

**Email links on your domain** -- Set `defaultCustomDomain` to a verified custom domain (reference the Auth0 Custom Domain Verification's `domain` output), so verification and password-reset emails link to your domain. The credential then also needs `read:custom_domains` and `update:custom_domains`.

**Destroy leaves the settings in place** -- Auth0 has no delete for tenant settings, so destroying this Cloud Resource stops managing them and keeps their last-applied values. To return a setting to a specific value, set that value before removing the field or the resource.

## Outputs and Dependencies

### What This Component Consumes

| Field | Foreign Key | Required |
|-------|-------------|----------|
| `defaultCustomDomain` | Auth0 Custom Domain Verification (`status.outputs.domain`) | No |

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

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Product-branded login** -- The tenant introduces itself as your product, with your logo and your support contacts. Start from the **Product-Branded Login** preset.

## Works With

- [**Auth0 Client (Application)**](/cloud-catalog/auth0-client) -- names the application in the login page's sentence.
- [**Auth0 Connection**](/cloud-catalog/auth0-connection) -- the sign-in methods the branded page offers.
- [**Auth0 Custom Domain Verification**](/cloud-catalog/auth0-custom-domain-verification) -- the verified domain the tenant's emails link to.
