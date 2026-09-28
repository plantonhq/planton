# Auth0Branding

Manages how an existing [Auth0 tenant](https://auth0.com/docs/customize/login-pages/universal-login)'s Universal Login looks, in two layers: the branding every login, signup, reset and consent page shares (logo, favicon, colors, font, and the page template the login box renders inside), and the no-code theme of the login box itself (button and input shapes, the full palette, font sizes, the page layout and the logo's placement).

## When to Use

- **Your logo and colors on every login page**, on any plan, the Free plan included.
- **Your own page around the login box**: a page template with your header, footer and styles, once the tenant has a custom domain.
- **Your product's look inside the login box**, without writing CSS: the theme's borders, colors, fonts and layout.
- **Infrastructure as code**: the tenant's login look is version-controlled beside its clients, connections and domain.

## Quick Start

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Branding
metadata:
  name: branding
  org: acme-corp
  env: production
spec:
  logoUrl: https://assets.acme.com/logo.png
  faviconUrl: https://assets.acme.com/favicon.png
  colors:
    primary: "#0059d6"
    pageBackground: "#000000"
```

## Fields

| Field | What it sets |
|---|---|
| `logoUrl` | The logo every Universal Login page shows (about 150 x 150 pixels); it takes precedence over the tenant's `pictureUrl` on the login pages |
| `faviconUrl` | The icon browsers show in the tab |
| `colors.primary` | The accent of buttons and links, a hex color |
| `colors.pageBackground` | The page around the login box: a hex color or Auth0's gradient JSON |
| `fontUrl` | A custom font file the pages load |
| `universalLoginTemplate` | The full page template, a Liquid HTML document containing `{%- auth0:head -%}` and `{%- auth0:widget -%}` |
| `theme` | The login box's theme: `displayName`, `borders`, `colors`, `fonts` (with the `bodyText`, `buttonsText`, `inputLabels`, `links`, `subtitle` and `title` text styles), `pageBackground`, `widget`, and `identifiers` |

## Key Behaviors

- **The tenant is the credential's**: the branding managed is that of the tenant the provider connection's credential belongs to. A tenant has one branding and one theme, so one resource per tenant.
- **Unset is unmanaged**: a branding setting left out is never sent, and the tenant keeps whatever it carries. The branding is declared only when the spec sets one of `logoUrl`, `faviconUrl`, `colors`, `fontUrl` or `universalLoginTemplate`, so a spec that declares only a theme leaves the logo, favicon, colors and font untouched. At least one setting or a theme must be set.
- **A theme is sent whole**: once `theme` is declared, every block Auth0 requires (borders, colors, fonts with all six text styles, the page background, the widget) is sent, and every field left out is sent at Auth0's default, so the theme always matches the declaration. The defaults are Auth0's own; each text style has its own (body text, links and subtitle 87.5%, buttons and input labels 100%, title 150%; only links bold).
- **Identifiers are sent only when declared**, and only on tenants where Auth0 has enabled the feature. Once applied they can be changed but not removed: removing the block leaves the last-applied values in place.
- **The page template needs a custom domain**: Auth0 refuses a template on the canonical domain, so compose an `Auth0CustomDomain` and its `Auth0CustomDomainVerification` first. Removing the template returns the pages to Auth0's default page.
- **Theme and branding overlap on the login box**: the theme's `widget.logoUrl`, when set, replaces `logoUrl` inside the box, and the theme's `fonts.fontUrl` replaces `fontUrl` there.
- **Permissions**: the credential needs `read:branding`, `update:branding` and `delete:branding`, and `read:custom_domains` (the provider checks for a custom domain on every read, update and delete), on the tenant's Management API (`iac/permissions.yaml`).

## Destroy

Auth0 has no delete for the tenant's branding, so destroying this resource:

- **removes the page template** (on a tenant with a custom domain), returning the pages to Auth0's default page;
- **deletes the theme**, returning the login box to Auth0's look. The theme is the tenant's one theme: applying a theme adopts a theme the tenant already carries, and destroy deletes it;
- **leaves the last-applied logo, favicon, colors and font in place**. To return one to a specific value, set that value before removing the field or the resource.

## Plans

Branding and the theme are on every Auth0 plan, the Free plan included. The page template needs a custom domain on the tenant; the Free plan includes one (a card on file, not charged).

## Outputs

| Output | Description |
|---|---|
| `theme_id` | The id of the theme the branding applied; empty when the spec declares no theme |
| `logo_url` | The logo the tenant's pages show after the deployment; empty when the spec manages no branding setting |

## Auth0 Documentation

- [Customize Universal Login themes](https://auth0.com/docs/customize/login-pages/universal-login/customize-themes)
- [Customize Universal Login page templates](https://auth0.com/docs/customize/login-pages/universal-login/customize-templates)
- [Terraform auth0_branding](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/branding)
- [Terraform auth0_branding_theme](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/branding_theme)
- [Pulumi auth0.Branding](https://www.pulumi.com/registry/packages/auth0/api-docs/branding/)
- [Pulumi auth0.BrandingTheme](https://www.pulumi.com/registry/packages/auth0/api-docs/brandingtheme/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
