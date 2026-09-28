# Auth0 Branding

Sets how an existing Auth0 tenant's Universal Login looks: the logo, favicon, colors and font every login page shares, the page template the login box renders inside, and the no-code theme of the login box itself. One Cloud Resource per tenant.

## What Gets Created

Nothing new: a tenant has one branding and one theme. When you deploy this Cloud Resource, the IaC module sets them on the tenant your Auth0 connection's credential belongs to:

- **Branding** -- the logo, favicon, primary color, page background and font of every login, signup, reset and consent page, and the page template, when you set any of them
- **Theme** -- the login box's borders, colors, fonts, page background and layout, when you declare one; every field you leave out takes Auth0's default

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with the tenant's domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Auth0 Custom Domain and Verification** -- only for a page template, which Auth0 accepts only on a tenant with a custom domain.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **The connection's Machine-to-Machine application** must hold `read:branding`, `update:branding`, `delete:branding` and `read:custom_domains` on the tenant's Management API (Auth0 dashboard: Applications, APIs, Auth0 Management API, Machine To Machine Applications).
- **Public HTTPS URLs** for your logo, favicon and font, if you set them.

## Deploy

### Console

Open the deployment store, find **Auth0 Branding**, and click **Deploy**. Start from the **Logo and Colors** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Branding
metadata:
  name: branding
  org: acme-corp
  env: prod
spec:
  logoUrl: https://assets.acme.com/logo.png
  colors:
    primary: "#0059d6"
    pageBackground: "#000000"
```

```shell
planton apply -f auth0-branding.yaml
```

Every login page of the tenant shows your logo on your background, with your accent on its buttons and links. A Stack Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Unset is unmanaged** -- A branding setting you leave out is never sent, and the tenant keeps its current value. A spec with only a theme leaves the logo, favicon, colors and font untouched.

**A theme is applied whole** -- Declare only the theme fields that matter; everything else is sent at Auth0's default, so the login box always matches the declaration. Identifiers are the exception: sent only when declared, and only on tenants with the feature.

**The page template needs a custom domain** -- `universalLoginTemplate` is a Liquid HTML page containing `{%- auth0:head -%}` and `{%- auth0:widget -%}`. Auth0 refuses it on the canonical domain, so order it after an Auth0 Custom Domain Verification.

**Destroy is partial** -- Auth0 has no delete for branding. Destroying this Cloud Resource removes the page template and deletes the theme, and keeps the last-applied logo, favicon, colors and font.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies. The tenant is the one the Auth0 connection's credential belongs to; a page template needs the tenant's custom domain to be verified first (order it with `depends_on`).

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `theme_id` | The id of the theme applied; empty without a theme | Audits, importing the theme |
| `logo_url` | The logo the tenant's pages show | Audits, email templates that match the login page |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Logo and colors** -- Your logo, accent and page background on every login page. Start from the **Logo and Colors** preset.

**Your own page** -- The login box inside your page's header and footer, on your custom domain. Start from the **Page Template on a Custom Domain** preset.

**A themed login box** -- Your product's buttons, inputs and layout, with no CSS. Start from the **Themed Login Box** preset.

## Works With

- [**Auth0 Custom Domain**](/cloud-catalog/auth0-custom-domain) and [**Auth0 Custom Domain Verification**](/cloud-catalog/auth0-custom-domain-verification) -- the domain a page template needs.
- [**Auth0 Tenant Settings**](/cloud-catalog/auth0-tenant-settings) -- the tenant's name in the login page's sentence, and its support contacts.
- [**Auth0 Client (Application)**](/cloud-catalog/auth0-client) -- names the application the branded page signs people in to.
